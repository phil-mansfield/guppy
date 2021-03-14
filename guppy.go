package main

import (
	"fmt"

	"github.com/phil-mansfield/guppy/lib"
	"github.com/phil-mansfield/guppy/lib/thread"
	"github.com/phil-mansfield/guppy/lib/error"
	read "github.com/phil-mansfield/guppy/go"
)

func main() {
	// Parse arguements.
	mode, configFile, cmdArgs := lib.ParseCommandLine()
	rawArgs := lib.ParseConfigFile(configFile)
	rawArgs.Overwrite(cmdArgs)

	// Do processing that doesn't external validation.
	args := rawArgs.Process()
	// Tell future functions that this isn't being run with MPI.
	args.RunMode = lib.GuppyMode
	
	// Run the chosen mode.
	switch mode {
	case "help":
		lib.PrintHelp(args)
	case "check":
		Check(args)
	case "convert":
		Convert(args)
	case "confirm":
		Confirm(args)
	default:
		error.External(
			"You attempted to run guppy in the mode '%s', but the only valid " +
				"modes are 'help', 'check', 'convert', and 'confirm'.", mode,
		)
	}
}

// Check runs guppy's "check" mode which tests for errors in the configuration
// arguments.
func Check(args *lib.Args) {
	ok := lib.Check(lib.GuppyMode, args)
	if ok {
		fmt.Println("No errors detected.")
	}
}

// Convert runs guppy's "convert" mode, which converts files into/out of the
// .gup format.
func Convert(args *lib.Args) {
	lib.Check(lib.GuppyMode, args)
	
	thread.Set(args.Threads)

	for snap := 0; snap < args.Snaps; snap++ {
		// Split particles into guppy data cubes.
		origHd, part := lib.CollectParticles(args, snap)
		buf := lib.SplitBuffer(args, part)
			
		for i, file := range args.GupFileNames {
			// Split off
			part.Split(args, i, buf)
			
			// Compress each file and write it to disk.
			hd := lib.CreateHeader(args, origHd, i)
			gupData := lib.Compress(args, buf)
			lib.Write(args, file, hd, origHd, gupData)
		}
	}
}

// Confirm run's guppy's "confirm" mode, which checks that
func Confirm(args *lib.Args) {
	lib.Check(lib.GuppyMode, args)

	thread.Set(args.Threads)

	for snap := 0; snap < args.Snaps; snap++ {
		// This is the one weak link in the "confirm" chain. If CollectParticles
		// has a bug in it, it might not be caught through "confirm".
		origHd, part := lib.CollectParticles(args, snap)

		// Loop over all the .gup files.
		for _, file := range args.GupFileNames {
			f := read.Open(file)
			lib.Confirm(origHd, part, f)
		}
	}
	fmt.Println("No errors detected.")
}
