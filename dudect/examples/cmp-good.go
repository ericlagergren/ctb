// +build main

package main

import (
	"crypto/rand"
	"crypto/subtle"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/elagergren/ctb/dudect"
)

func main() {
	debug.SetGCPercent(-1)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	const N = 512
	cfg := &dudect.Config{
		ChunkSize:    N,
		Measurements: 1000,
		Output:       os.Stderr,
	}
	ctx := dudect.NewContext(cfg)

	secret := make([]byte, N)
	_, err := rand.Read(secret)
	if err != nil {
		panic(err)
	}
	fn := func(data []byte) bool {
		return subtle.ConstantTimeCompare(data, secret) == 1
	}
	t := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-t.C:
			break
		default:
		}
		if ctx.Test(fn, nil) {
			break
		}
	}
}
