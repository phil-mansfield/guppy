package main

import (
	"github.com/phil-mansfield/guppy/lib"
	read "github.com/phil-mansfield/guppy/go"
)

func main() {
	// Parse arguements.
	configFile, mode, cmdArgs := lib.ParseCommandLine()
	configArgs := lib.ParseConfigFile(configFile)
	rawArgs := lib.CombineArguments(configArgs, cmdArgs)
	args := lib.ProcessArguments(rawArgs)

	// Run the chosen mode.
	switch mode {
	case "help":
		lib.PrintHelp(lib.GuppyMode)
	case "check":
		Check(args)
	case "convert":
		Convert(args)
	case "confirm":
		Confirm(args)
	default:
		lib.ExternalErrorf(
			"You attempted to run guppy in the mode '%s', but the only valid " +
				"modes are 'help', 'check', 'convert', and 'confirm'.", mode,
		)
	}
}

// Check runs guppy's "check" mode which tests for errors in the configuration
// arguments.
func Check(args *lib.Args) {
	lib.Check(lib.GuppyMode, lib.CrashOnError, args)
	fmt.Println("No errors detected.")
}

// Convert runs guppy's "convert" mode, which converts files into/out of the
// .gup format.
func Convert(args *lib.Args) {
	lib.Check(lib.GuppyMode, lib.WarnOnError, args)
	
	lib.StartThreads(args)

	for _, snap := range args.Snaps {
		// Split particles into guppy data cubes.
		part := lib.CollectParticles(args, snap)
		parts := part.Split(args, part)
		
		for file := range parts {
			// Compress each file and write it to disk.
			gupData := lib.Compress(args, parts[i])
			lib.Write(args, snap, file, gupData)
		}
	}
}

// Confirm run's guppy's "confirm" mode, which checks that
func Confirm(args *lib.Args) {
	lib.Check(lib.GuppyMode, lib.WarnOnError, args)

	lib.StartThreads(args)

	for _, snap := range args.Snaps {
		// This is the one weak link in the "confirm" chain. If CollectParticles
		// has a bug in it, it might not be caught through "confirm".
		part := lib.CollectParticles(args, snap)

		// Loop over all the .gup files.
		gupFiles := lib.GupFileNames(args)
		for file := range gupFiles {
			gupData := read.ReadAll(fupFiles[file])
			lib.ConfirmGupData(gupData, part)
		}
	}
	fmt.Println("No errors detected.")
}
