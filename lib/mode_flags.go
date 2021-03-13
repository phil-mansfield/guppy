package lib

// RunMode indicates whether a function is being run through guppy or
// mpi_guppy.
type RunMode int
const (
	GuppyMode RunMode = iota
	MpiGuppyMode
)

// CheckStrictness indicates how functions related to the "check" guppy mode
// should behave when it encounters an error.
type CheckStrictness int
const (
	CrashOnError CheckStrictness = iota
	WarnOnError
)
