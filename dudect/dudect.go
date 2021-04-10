// Package dudect implements a side channel leak detector.
//
// In order to get accurate readings, you'll want to disable as
// much of the Go runtime as possible. In particular, setting
// some of the following GODEBUG environment variables
//
//    asyncpreemptoff=1
//    gcshrinkstackoff=1
//    sbrk=1
//
// disabling garbage collection
//
//    debug.SetGCPercent(-1) // or GOGC=off
//
// and locking the goroutine running the tests to an OS thread
//
//    runtime.LockOSThread()
//    defer runtime.UnlockOSThread()
//
// This package is a Go transliteration of github.com/oreparaz/dudect.
package dudect

import (
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"sort"
)

const (
	enoughMeasurements = 10000
	numPercentiles     = 100
)

// Config configures a Test.
type Config struct {
	// ChunkSize is the size of the data
	// processed each iteration.
	//
	// For example, if testing
	//    func constantTimeCompare(a, b [16]byte) bool
	// set ChunkSize to 16.
	ChunkSize int
	// Measurements is the number of measurements per
	// call to Test.
	//
	// Memory usage is proportional to the number of tests.
	Measurements int
	// Output, if non-nil, is used to print informational
	// messages.
	Output io.Writer
}

// Context records test measurements.
type Context struct {
	cfg       *Config
	execTimes []int64 // len == Measurements
	inputData []byte  // len == Measurements*ChunkSize
	classes   []byte  // len == Measurements
	// perform this many tests in total:
	//   - 1 first order uncropped test,
	//   - numPercentiles tests
	//   - 1 second order test
	testCtxs    [1 + numPercentiles + 1]testCtx
	percentiles [numPercentiles]int64
}

func NewContext(cfg *Config) *Context {
	return &Context{
		cfg:       &(*cfg),
		execTimes: make([]int64, cfg.Measurements),
		classes:   make([]byte, cfg.Measurements),
		inputData: make([]byte, cfg.Measurements*cfg.ChunkSize),
	}
}

func (ctx *Context) printf(format string, args ...interface{}) {
	if ctx.cfg.Output != nil {
		fmt.Fprintf(ctx.cfg.Output, format, args...)
	}
}

// PrepFunc prepares input for testing.
//
// The length of data is cfg.ChunkSize*cfg.Measurements.
//
// Each element in classes must be either 0 or 1.
// The length of classes is cfg.Measurements.
//
// PrepFunc is only called once per test.
type PrepFunc func(cfg *Config, input, classes []byte)

// Prepare is an implementation of PrepFunc.
func Prepare(cfg *Config, input, classes []byte) {
	_, err := rand.Read(input)
	if err != nil {
		panic(err)
	}
	_, err = rand.Read(classes)
	if err != nil {
		panic(err)
	}
	for i, c := range classes {
		classes[i] = c & 1
		if c&1 != 0 {
			// Leave random
			continue
		}
		p := input[i*cfg.ChunkSize : (i+1)*cfg.ChunkSize]
		for j := range p {
			p[j] = 0
		}
	}
}

// Test samples fn and reports whether any leaks were found.
//
// Test should be called in a loop like so:
//
//    for {
//        select {
//        case <-timeout.C:
//            break
//        default:
//        }
//        if ctx.Test(fn, prep) {
//            break
//        }
//    }
//
// In other words, Test cannot provide fn is constant
// time. Instead, it tries to find leaks. For more
// information, see github.com/oreparaz/dudect.
//
// If prep is nil, Prepare is used instead.
func (ctx *Context) Test(fn func([]byte) bool, prep PrepFunc) bool {
	if prep == nil {
		prep = Prepare
	}
	prep(ctx.cfg, ctx.inputData, ctx.classes)
	ctx.measure(fn)

	first := ctx.percentiles[len(ctx.percentiles)-1] == 0
	if first {
		// Throw away the first batch of measurements.
		// This helps warming things up.
		ctx.prepPercentiles()
		return false
	}
	ctx.updateStats()
	return ctx.leaks()
}

func (ctx *Context) measure(fn func([]byte) bool) {
	for i := 0; i < ctx.cfg.Measurements; i++ {
		data := ctx.inputData[i*ctx.cfg.ChunkSize : (i+1)*ctx.cfg.ChunkSize]
		start := cpucycles()
		fn(data)
		last := cpucycles()
		ctx.execTimes[i] = last - start
	}
	// fmt.Println(ctx.execTimes)
}

func (ctx *Context) updateStats() {
	// Set i = 10 to discard the first few measurements.
	for i := 10; i < ctx.cfg.Measurements-1; i++ {
		diff := ctx.execTimes[i]
		if diff < 0 {
			// The cpu cycle counter overflowed, just throw away the measurement.
			continue
		}

		// t-test on the execution time
		ctx.testCtxs[0].push(float64(diff), ctx.classes[i])

		// t-test on cropped execution times, for several cropping thresholds.
		for j, pct := range ctx.percentiles {
			if diff < pct {
				ctx.testCtxs[j+1].push(float64(diff), ctx.classes[i])
			}
		}

		// second-order test (only if we have more than 10000 measurements).
		// Centered product pre-processing.
		if ctx.testCtxs[0].n[0] > 10000 {
			centered := float64(diff) - ctx.testCtxs[0].mean[ctx.classes[i]]
			ctx.testCtxs[1+numPercentiles].push(
				centered*centered, ctx.classes[i])
		}
	}
}

// leaks reports whether any leaks were found.
func (ctx *Context) leaks() bool {
	t := ctx.maxTest()
	maxT := math.Abs(t.compute())
	numTracesMaxT := t.n[0] + t.n[1]
	maxTau := maxT / math.Sqrt(numTracesMaxT)

	// print the number of measurements of the test that yielded max t.
	// sometimes you can see this number go down - this can be confusing
	// but can happen (different test)
	ctx.printf("meas: %7.2f M, ", (numTracesMaxT / 1e6))
	if numTracesMaxT < enoughMeasurements {
		ctx.printf("not enough measurements (%.0f still to go).\n",
			enoughMeasurements-numTracesMaxT)
		return false
	}

	//
	// We report the following statistics:
	//
	//    maxT: the t value
	//    maxTau: a t value normalized by sqrt(number of measurements).
	//            this way we can compare maxTau taken with different
	//            number of measurements. This is sort of "distance
	//            between distributions", independent of number of
	//            measurements.
	//    (5/tau)^2: how many measurements we would need to barely
	//               detect the leak, if present. "barely detect the
	//               leak" here means have a t value greater than 5.
	//
	// The first metric is standard; the other two aren't, but
	// are pretty sensible imho.
	ctx.printf("max t: %+7.2f, max tau: %.2e, (5/tau)^2: %.2e.",
		maxT, maxTau, 5*5/float64(maxTau*maxTau))

	// threshold values for Welch's t-test
	const (
		thresholdBananas  = 500 // test failed, with overwhelming probability
		thresholdModerate = 10  // test failed. Pankaj likes 4.5 but let's be more lenient
	)
	if maxT > thresholdBananas {
		ctx.printf(" Definitely not constant time.\n")
		return true
	}
	if maxT > thresholdModerate {
		ctx.printf(" Probably not constant time.\n")
		return true
	}
	if maxT < thresholdModerate {
		ctx.printf(" For the moment, maybe constant time.\n")
	}
	return false
}

func (ctx *Context) maxTest() testCtx {
	idx := 0
	var max float64
	for i, tctx := range ctx.testCtxs {
		if tctx.n[0] <= enoughMeasurements {
			continue
		}
		x := math.Abs(tctx.compute())
		if x > max {
			max = x
			idx = i
		}
	}
	return ctx.testCtxs[idx]
}

// prepPercentiles sets different thresholds for cropping measurements.
//
// The exponential tendency is meant to approximately match
// the measurements distribution, but there's not more science
// than that.
func (ctx *Context) prepPercentiles() {
	sort.Slice(ctx.execTimes, func(i, j int) bool {
		return ctx.execTimes[i] < ctx.execTimes[j]
	})
	for i := range ctx.percentiles {
		w := 1 - math.Pow(0.5, 10*float64(i+1)/numPercentiles)
		ctx.percentiles[i] = ctx.execTimes[int(w)*ctx.cfg.Measurements]
	}
}

type testCtx struct {
	mean [2]float64
	m2   [2]float64
	n    [2]float64
}

// push implements Welch's t-test.
//
// Welch's t-test test whether two populations have
// the same mean. This is basically Student's t-test
// for unequal variances and unequal sample sizes.
//
// See https://en.wikipedia.org/wiki/Welch%27s_t-test
func (ctx *testCtx) push(x float64, class uint8) {
	if class > 1 {
		panic("class > 1")
	}
	ctx.n[class]++
	// Estimate variance on the fly as per the Welford method.
	// This gives good numerical stability. See Knuth's TAOCP vol 2.
	delta := x - ctx.mean[class]
	ctx.mean[class] += delta / ctx.n[class]
	ctx.m2[class] += delta * (x - ctx.mean[class])
}

func (ctx *testCtx) compute() float64 {
	v0 := ctx.m2[0] / (ctx.n[0] - 1)
	v1 := ctx.m2[1] / (ctx.n[1] - 1)
	num := ctx.mean[0] - ctx.mean[1]
	den := math.Sqrt(v0/ctx.n[0] + v1/ctx.n[1])
	return num / den
}
