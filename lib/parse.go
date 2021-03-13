package lib

// RawArgs stores the unprocessed values which the user assigned to each config
// variable.
type RawArgs struct {
}

// Args stores configuration information. It is a post-processed version of
// RawArgs.
type Args struct {
	Snaps int
	Threads int
	RunMode RunMode

	GupFileNames []string
}

// ParseCommandLine parses the command line arguments and returns the mode guppy
// is being run in, the name of the config file, and any arguments which were
// set. Expects that the arguments are presented in the order:
// $ guppy <mode> <config file> [--<Arg1> <Value1>] [--<Arg2> <Value2>]
func ParseCommandLine() (mode, configFile string, args *RawArgs) {
	panic("NYI")
}

// ParseConfigFile parses arguements from a config file.
func ParseConfigFile(fileName string) *RawArgs {
	panic("NYI")
}

// Overwrite arguments in arg1 which have been set to non-default values in
// arg2.
func (arg1 *RawArgs) Overwrite(arg2 *RawArgs) {
	panic("NYI")
}

// Process converts the raw user input to a format which is more useful for
// internal functions. Very simple validation will be done here, but nothing
// which requires interacting with external files.
func (args *RawArgs) Process() *Args {
	panic("NYI")
}


