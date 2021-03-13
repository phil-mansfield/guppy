package lib

/* check.go contains the core functions of guppy's "check" mode. */

// Check runs the guppy "check" command on the provided Args. The user must also
// specify whether guppy or mpi_guppy is being run through the mode flag. This
// function will weither crash upon encountering errors or will print warnings,
// depending on what CheckStrictness is set to in args. If Check completes,
// it returns true if all tests passed and false otherwise.
func Check(mode RunMode, args *Args) bool {
	panic("NYI")
}
