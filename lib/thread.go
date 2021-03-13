package lib

/* thread.go contains functions useful for multi-threading. */

import (
	"runtime"
)

//
func SetThreads(n int) {
	if n > runtime.NumCPU() {
		ExternalErrorf("%d threads requested, but your system only has %d cores per node. If you want guppy to use the maximum number of threads per node, set Threads=-1.")
	}

	runtime.GOMAXPROCS(n)
}
