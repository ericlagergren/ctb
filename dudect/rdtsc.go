package dudect

import _ "unsafe" // for go:linkname

/*
 Intel actually recommends calling CPUID to serialize the execution flow
 and reduce variance in measurement due to out-of-order execution.
 We don't do that here yet.
 see §3.2.1 http://www.intel.com/content/www/us/en/embedded/training/ia-32-ia-64-benchmark-code-execution-paper.html
*/
func cpucycles() int64

//go:linkname cpucycles runtime.cputicks
