package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"syscall"

	"github.com/phil-mansfield/guppy/lib/config"
	"github.com/phil-mansfield/guppy/lib/format"
	"github.com/phil-mansfield/guppy/lib/thread"
	guppy "github.com/phil-mansfield/guppy/go"

	"reflect"
	"unsafe"
)

const (
	// *Mode constants are internal values indicating which mode the server is
	// being run in.
	ReadMode = iota
	WriteMode
	ExampleConfigMode
	CreatePipesMode
	DeletePipesMode

	// *Format constants are internal values indicating which format
	// output/input is in.
	RockstarFormat = iota
)

var (
	// Version is the version of the software. This can potentially be used
	// to differentiate between breaking changes to the input/output format.
	Version uint64 = 0x1
	RockstarFormatCode uint64 = 0xffffffff00000001
)

func main() {
	// Parse arguments
	confName, first, last, mode := ParseArguments()
	
	// Read the config file
	var conf *Config
	switch mode {
	case ReadMode, WriteMode, CreatePipesMode, DeletePipesMode:
		conf = ParseConfig(confName)

		if conf.Threads < 0 { conf.Threads = int64(runtime.NumCPU()) }
		thread.Set(int(conf.Threads))

		if last >= int(conf.Blocks) {
			fmt.Fprintf(os.Stderr, "LastBlock, %d, is >= than the " +
				"number of blocks, %d.\n", last, conf.Blocks)
			os.Exit(1)
		}
	}

	// Run the server in the selected mode
	switch mode {
	case ReadMode: Read(conf, first, last)
	case WriteMode: Write(conf, first, last)
	case CreatePipesMode: CreatePipes(conf)
	case DeletePipesMode: DeletePipes(conf)
	case ExampleConfigMode: ExampleConfig()
	default: panic("Internal error.")
	}
}

///////////////////////
// Parsing functions //
///////////////////////

// Config contains the values in the server's config files.
type Config struct {
	Format, GuppyFiles, PipeDirectory, Snapshots string
	Blocks, Threads int64
}

// ParseConfig parses the config file with the name confName.
func ParseConfig(confName string) *Config {	
	conf := &Config{ }
	
	vars := config.NewConfigVars("guppy_server")
	vars.String(&conf.Format, "Format", "")
	vars.String(&conf.GuppyFiles, "GuppyFiles", "")
	vars.String(&conf.PipeDirectory, "PipeDirectory", "")
	vars.String(&conf.Snapshots, "Snapshots", "")
	vars.Int(&conf.Blocks, "Blocks", -1)
	vars.Int(&conf.Threads, "Threads", -1)

	err := config.ReadConfig(confName, vars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read config file, '%s': %s\n",
			confName, err.Error())
		os.Exit(1)
	}

	return conf
}

// ParseArguments parses command line arguments, and returns the name of the
// config file, the first and last blocks analyzed, and the server mode,
// respectively.
func ParseArguments() (confName string, first, last int, mode int) {
	if len(os.Args) < 2 { Usage() }
	
	modeName, args := os.Args[1], os.Args[2:]
	switch modeName {
	case "read": return ParseReadArguments(args)
	case "write": return ParseWriteArguments(args)
	case "example_config": return ParseExampleConfigArguments(args)
	case "create_pipes": return ParseCreatePipesArguments(args)
	case "delete_pipes": return ParseDeletePipesArguments(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode, '%s'\n\n", modeName)
		Usage()
		return "", -1, -1, -1
	}
}

// ParseReadArguments parses arguments for the "read" server mode.
func ParseReadArguments(args []string) (
	confName string, first, last int, mode int,
) {
	if len(args) != 3 { Usage() }

	confName, firstStr, lastStr := args[0], args[1], args[2]

	if _, err := os.Stat(confName); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ConfigName, '%s' does not exist\n\n",
			confName)
		Usage()
	}
	
	first, err := strconv.Atoi(firstStr)
	if err != nil || first < 0 {
		fmt.Fprintf(os.Stderr, "FirstPipe, '%s' is not a valid argument.\n\n",
			firstStr)
		Usage()
	}

	last, err = strconv.Atoi(lastStr)
	if err != nil || first < 0 {
		fmt.Fprintf(os.Stderr, "LastPipe, '%s' is not a valid argument.\n\n",
			lastStr)
		Usage()
	}

	return confName, first, last, ReadMode
}

// ParseReadArguments parses arguments for the "write" server mode.
func ParseWriteArguments(args []string) (
	confName string, first, last int, mode int,
) {
	if len(args) != 3 { Usage() }

	confName, firstStr, lastStr := args[0], args[1], args[2]

	if _, err := os.Stat(confName); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ConfigName, '%s' does not exist\n\n",
			confName)
		Usage()
	}
	
	first, err := strconv.Atoi(firstStr)
	if err != nil || first < 0 {
		fmt.Fprintf(os.Stderr, "FirstPipe, '%s' is not a valid argument.\n\n",
			firstStr)
		Usage()
	}

	last, err = strconv.Atoi(lastStr)
	if err != nil || first < 0 {
		fmt.Fprintf(os.Stderr, "LastPipe, '%s' is not a valid argument.\n\n",
			lastStr)
		Usage()
	}

	return confName, first, last, WriteMode
}

// ParseReadArguments parses arguments for the "example_config" server mode.
func ParseExampleConfigArguments(args []string) (
	confName string, first, last int, mode int,
) {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "Wrong number of arguments for " +
			"'example_config' mode.\n\n")
		Usage()
	}
	return "", -1, -1, ExampleConfigMode
}

// ParseReadArguments parses arguments for the "create_pipes" server mode.
func ParseCreatePipesArguments(args []string) (
	confName string, first, last int, mode int,
) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Wrong number of arguments for " +
			"'create_pipes' mode.\n\n")
		Usage()
	} else if _, err := os.Stat(args[0]); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ConfigName, '%s' does not exist\n\n", args[0])
		Usage()
	}

	return args[0], -1, -1, CreatePipesMode
}

// ParseReadArguments parses arguments for the "delete_pipes" server mode.
func ParseDeletePipesArguments(args []string) (
	confName string, first, last int, mode int,
) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Wrong number of arguments for " +
			"'delete_pipes' mode.\n\n")
		Usage()
	} else if _, err := os.Stat(args[0]); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ConfigName, '%s' does not exist\n\n", args[0])
		Usage()
	}
	return args[0], -1, -1, DeletePipesMode
}

//Usage prints the expected usage and exits with an error code.
func Usage() {
	fmt.Fprintf(os.Stderr, `Expected usage:
./guppy_server read <ConfigName> <FirstPipe> <LastPipe>
./guppy_server write <ConfigName> <FirstPipe> <LastPipe>
./guppy_server create_pipes <ConfigName>
./guppy_server delete_pipes <ConfigName>
./guppy_server example_config

- "read" mode will read compressed data and write it in uncompressed form to a
  series of pipes which can then be read by another process.
- "write" mode will read uncompressed data from a series of pipes and writes it
  in compressed form to disk.
- "create_pipes" creates the pipes that will be used by the server.
- "delete_pipes" deletes these pipes.
- "example_config" creates an example configuration file

- <ConfigName> is the name of the config file being use for the server
- <FirstPipe> and <LastPipe> are the index of the first and last pipes used
  by this particular instance of the server. This will allow different machines
  to load balance running the server.
`)
	os.Exit(1)
}

//////////////////
// Server Modes //
//////////////////

// ExampleCOnfig prints and exmaple configuration file to stdout.
func ExampleConfig() {
	fmt.Println(`[guppy_server]

###########################################
## Variables needed by the create_pipes, ##
## delete_pipes, read, and write modes   ##
###########################################

# Blocks gives the number of blocks (files) per snapshot.
Blocks = 512

# PipeDirectory is the directory where pipes are created.
PipeDirectory = path/to/pipes/

#############################
## Variables needed by the ##
## read and write modes    ##
#############################

# Format tells the server what format to use when writing data to the pipes.
# currently the only supported format is "rockstar". See the online
# documentation for a description of this format.
Format = rockstar

# Snapshots is a string representing the snapshots used by the input. In most
# cases, this string will be MinSnapshot..MaxSnapshot and will enumerate all the
# snapshots in the range [MinSnapshot, MaxSnapshot].
#
# If you have a more complicated set of snapshots that you want to look at (say,
# some of your snapshots are corrupted, or you only want to compress some
# snapshots), you can remove individual snapshots or sequences of snapshots
# with "-", and you can add snapshots in the same way with "+". The example
# below would be for a simulation with snapshots ranging from 0 to 100, but
# which has a corrupted snapshot 63 that needs to be removed.
Snapshots = 0..100 - 63

# GuppyFiles is a format string representing the location of guppy files. It is
# similar to the scheme used by Rockstar's config files, but it allows you to
# specify how a value is printed. Input files will generally be at some 
# extended path with integers at various locations which represent snapshots
# and blocks. You can specify where these numbers are by placing them in braces 
# with the form {integer_format:variable_name}. "variable_name" should be either
# "snapshot"  or "block", and "integer_format" should be a valid C-style printf
# verb (e.g.  %03d will convert 97 to 097).
GuppyFiles = path/to/sim/snapdir_{%03d:snapshot}/snapshot_{%03d:snapshot}.{%d:block}.gup

# Number of threads to use during execution. If set to -1, one thread will be
# used for each core on the node.
Threads = -1

#########################
## Variables needed by ##
## the write mode      ##
#########################

# Not yet implemented
`)
	os.Exit(0)
}

// Create pipes creates the appropriate number of pipes in PipeDirectory.
func CreatePipes(conf *Config) {
	for block := 0; block < int(conf.Blocks); block++ {
		pipeName := path.Join(conf.PipeDirectory, fmt.Sprintf("pipe.%d", block))

		if _, err := os.Stat(pipeName); !os.IsNotExist(err) {
			os.Remove(pipeName)
		}

		err := syscall.Mkfifo(pipeName, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while calling 'mkfifo %s': %s\n",
				pipeName, err.Error())
			os.Exit(1)
		}
	}
}

// Create pipes deletes the pipes in PipeDirectory.
func DeletePipes(conf *Config) {
	for block := 0; block < int(conf.Blocks); block++ {
		pipeName := path.Join(conf.PipeDirectory, fmt.Sprintf("pipe.%d", block))
		err := os.Remove(pipeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while calling 'rm %s': %s\n",
				pipeName, err.Error())
			os.Exit(1)
		}
	}
}

// Read reads guppy files from disk and writes uncompressed data to pipes.
func Read(conf *Config, first, last int) {
	snaps := GetSnaps(conf)
	
	workers, jobs := int(conf.Threads), last - first + 1
	guppy.InitWorkers(workers)
	//workers = jobs

	bufs := make([][]guppy.RockstarParticle, workers)
	
	var format int
	switch conf.Format {
	case "rockstar": format = RockstarFormat
	default:
		fmt.Fprintf(os.Stderr, "Unrecognized 'Format' variable, '%s'.\n",
			conf.Format)
		os.Exit(1)
	}
	
	for _, snap := range snaps {
		// Use a queue to assign threads to different pipes.
		thread.WorkerQueue(workers, jobs, func(worker, job int) {
			fmt.Fprintf(os.Stderr, "worker %d job %d\n", worker, job)
			switch format {
			case RockstarFormat:
				GuppyToPipeRockstar(conf, worker, job+first, snap, bufs)
			default:
				panic("Internal error")
			}
		})
	}
}

// GetSnaps return the target snapshots.
func GetSnaps(conf *Config) []int {
	snapSeq := conf.Snapshots
	snaps, err := format.ExpandSequenceFormat(snapSeq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot parse the 'Snapshots' variable, " +
			"'%s': %s", snapSeq, err.Error())
		os.Exit(1)
	}
	return snaps
}

// Write reads uncompressed particles from pipes and compresses them into
// .gup files.
func Write(conf *Config, first, last int) {
	panic("NYI")
}

/////////////////////////////////////
// Rockstar-format reading/writing //
/////////////////////////////////////

// GuppyToPipeRockstar converts a guppy file into a data stream through a
// pipe formatted as rockstar particles.
func GuppyToPipeRockstar(
	conf *Config, worker, block, snap int,
	buf [][]guppy.RockstarParticle, 
) {
	vars := map[string]int{ "snapshot": snap, "block": block }	
	guppyName, err := format.ExpandFormatString(conf.GuppyFiles, vars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse 'GuppyFiles': %s", err.Error())
		os.Exit(1)
	} else if len(guppyName) != 1 {
		fmt.Fprintf(os.Stderr, "Could not parse 'GuppyFiles': '%s' results " +
			"in %d names per snapshot-block pair.", err.Error())
		os.Exit(1)
	}

	hd := guppy.ReadHeader(guppyName[0])
	
	if len(buf[worker]) == 0 {
		buf[worker] = make([]guppy.RockstarParticle, hd.N)
	} else if len(buf[worker]) != int(hd.N) {
		fmt.Fprintf(os.Stderr, "%s contains %d particles, but the previous " +
			"the same block in the previously read snapshot had %d " +
			"particles. Each guppy block must be the same size across all " +
			"snapshots. You probably got in this situation because someone " +
			"manually ran guppy on different snapshots with different block " +
			"sizes. In that case, you'll need to manually read the different " +
			"snapshots separately, as well.\n", guppyName, hd.N,
			len(buf[worker]))
		os.Exit(1)
	}

	fmt.Printf("Reading %s with worker %d, block %d, snap %d\n",
		guppyName[0], worker, block, snap)
	guppy.ReadVar(guppyName[0], "[RockstarParticle]", -2, buf[worker])

	WriteRockstarToPipe(conf, hd, buf[worker], worker, block, snap)
}

// WriteRockstarToPipe handles the actual disk writes that make up
// GuppyToPipeRockstar.
func WriteRockstarToPipe(
	conf *Config, hd *guppy.Header, buf []guppy.RockstarParticle,
	worker, block, snap int,
) {
	pipeName := path.Join(conf.PipeDirectory, fmt.Sprintf("pipe.%d", block))
	fmt.Fprintf(os.Stderr, "Trying to open pipe, %s.\n", pipeName)
	pipe, err := os.OpenFile(pipeName, os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open pipe '%s': %s",
			pipeName, err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Opened pipe, %s.\n", pipeName)

	defer pipe.Close()

	order := GetSystemOrder()
	rsHeader := GuppyToRockstarHeader(hd)

	fmt.Printf("Writing header with worker %d, block %d, snap %d\n",
		worker, block, snap)
	err = binary.Write(pipe, order, rsHeader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Worker %d could not write the rockstar " +
			"header of block %d, snap %d to its pipe. %s", worker, block, snap,
			err.Error())
		os.Exit(1)
	}

	sliceHd := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sliceHd.Len *= 32
	sliceHd.Cap *= 32	

	fmt.Printf("Writing particle with worker %d, block %d, snap %d\n",	
		worker, block, snap)
	b := *(*[]byte)(unsafe.Pointer(&sliceHd))
	_, err = pipe.Write(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Worker %d could not write the rockstar " +
			"particles in block %d, snap %d to its pipe. %s.\n", worker, block,
			snap, err.Error())
		os.Exit(1)
	}
	
	sliceHd.Len /= 32
	sliceHd.Cap /= 32

	fmt.Printf("Finished writing with worker %d, block %d, snap %d\n",	
		worker, block, snap)
}

// GetSystemOrder returns the byte ordering of the current system.
func GetSystemOrder() binary.ByteOrder {
    buf := [2]byte{ }
    *(*uint16)(unsafe.Pointer(&buf[0])) = uint16(1)
	if buf == [2]byte{ 0, 1 } { return binary.BigEndian }
	return binary.LittleEndian
}

// RockstarHeader is the first block of data written to pipes in the Rockstar
// format.
type RockstarHeader struct {
	Version, Format uint64
	N, NTot int64
    Span [3]int64
    Z, OmegaM, OmegaL, H100, L, Mass float64
}

// GuppyToRockstarHeader converts a Guppy header to a Rockstar header.
func GuppyToRockstarHeader(hd *guppy.Header) *RockstarHeader {
	return &RockstarHeader{
		Version, RockstarFormatCode, hd.N, hd.NTot, hd.Span,
		hd.Z, hd.OmegaM, hd.OmegaL, hd.H100, hd.L, hd.Mass,
	}
}
