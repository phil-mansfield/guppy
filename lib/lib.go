/*package lib contains various functions needed by guppy and mpi_guppy. The
functions in this particular package mainly just utility functions that might
be useful for other programs manually piping output from Guppy. Almost all
of the heavy lifting is done by lib/'s subpackages.
*/
package lib

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	
	"reflect"
	"unsafe"

	"github.com/phil-mansfield/guppy/lib/config"
	"github.com/phil-mansfield/guppy/lib/format"
	"github.com/phil-mansfield/guppy/lib/snapio"
	"github.com/phil-mansfield/guppy/lib/compress"
)

var (
	// Version is the version of the software. This can potentially be used
	// to differentiate between breaking changes to the input/output format.
	Version uint64 = 0x1
	RockstarFormatCode uint64 = 0xffffffff00000001

	SupportedCompressionMethods = []string{ "LagrangianDelta" }
	SupportedIDOrders = []string{ "ZUnigridPlusOne" }
)

// RockstarParticle is a particle with the structure expected by the
// Rockstar halo finder.
type RockstarParticle struct {
	ID uint64 
	X, V [3]float32
}

type PipeHeader struct {
    Version, Format uint64
    N, NTot int64
    Span, Origin, TotalSpan [3]int64
    Z, OmegaM, OmegaL, H100, L, Mass float64
}

func WriteAsBytes(f io.Writer, buf interface{}) error {
	sysOrder := SystemByteOrder()
	switch x := buf.(type) {
	case []uint32: return binary.Write(f, sysOrder, x)
	case []uint64: return binary.Write(f, sysOrder, x)
	case []float32: return binary.Write(f, sysOrder, x)
	case []float64: return binary.Write(f, sysOrder, x)
	case [][3]float32:
		// Go uses the reflect package to write non-primitive data through
		// the binary package. This is slow and makes tons of heap allocations.
		// So you need to be sneaky and "cast" to a primitive array.
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= 3
        hd.Cap *= 3

        f32x := *(*[]float32)(unsafe.Pointer(&hd))
        err := binary.Write(f, sysOrder, f32x)

        hd.Len /= 3
        hd.Cap /= 3

		return err
		
	case [][3]float64:
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= 3
        hd.Cap *= 3

        f64x := *(*[]float64)(unsafe.Pointer(&hd))
        err := binary.Write(f, sysOrder, f64x)

        hd.Len /= 3
        hd.Cap /= 3

		return err
		
	case []RockstarParticle:
		particleSize := int(unsafe.Sizeof(RockstarParticle{ }))
		
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= particleSize
        hd.Cap *= particleSize

		// RockstarParticle fields have inhomogenous sizes, so we need to
		// convert to bytes.
        bx := *(*[]byte)(unsafe.Pointer(&hd))
        _, err := f.Write(bx)

        hd.Len /= particleSize
        hd.Cap /= particleSize

		return err
	}
	
	panic("Internal error: unrecognized type of interal buffer.")
}

func ReadAsBytes(f io.Reader, buf interface{}) error {
	sysOrder := SystemByteOrder()
	switch x := buf.(type) {
	case []uint32: return binary.Read(f, sysOrder, x)
	case []uint64: return binary.Read(f, sysOrder, x)
	case []float32: return binary.Read(f, sysOrder, x)
	case []float64: return binary.Read(f, sysOrder, x)
	case [][3]float32:
		// Go uses the reflect package to write non-primitive data through
		// the binary package. This is slow and makes tons of heap allocations.
		// So you need to be sneaky and "cast" to a primitive array.
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= 3
        hd.Cap *= 3

        f32x := *(*[]float32)(unsafe.Pointer(&hd))
        err := binary.Read(f, sysOrder, f32x)

        hd.Len /= 3
        hd.Cap /= 3

		return err
		
	case [][3]float64:
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= 3
        hd.Cap *= 3

        f64x := *(*[]float64)(unsafe.Pointer(&hd))
        err := binary.Read(f, sysOrder, f64x)

        hd.Len /= 3
        hd.Cap /= 3

		return err
		
	case []RockstarParticle:
		particleSize := int(unsafe.Sizeof(RockstarParticle{ }))
		
		hd := *(*reflect.SliceHeader)(unsafe.Pointer(&x))
        hd.Len *= particleSize
        hd.Cap *= particleSize

		// RockstarParticle fields have inhomogenous sizes, so we need to
		// convert to bytes.
        bx := *(*[]byte)(unsafe.Pointer(&hd))
        _, err := io.ReadFull(f, bx)

        hd.Len /= particleSize
        hd.Cap /= particleSize

		return err
	}
	
	panic("Internal error: unrecognized type of interal buffer.")
}

func SystemByteOrder() binary.ByteOrder {
	// See https://stackoverflow.com/questions/51332658/any-better-way-to-check-endianness-in-go/51332762
	b := [2]byte{ }
	*(*uint16)(unsafe.Pointer(&b[0])) = uint16(0x0001)
	if b[0] == 0 {
		return binary.BigEndian
	} else {
		return binary.LittleEndian
	}
}

func ExampleWriteConfig() string {
	return `[write]

#######################
# Compression Options #
#######################

# CompressionMethod is the method used to compress particles. Right now only
# "LagrangianDelta" is supported.
CompressionMethod = LagrangianDelta

# Vars specfies the variables that should be added to the files, using the same
# naming scheme that your input files used.
Vars = x, v, id

# Types specifies the types of these variables. u32/u64 mean 32- and 64-bit
# unsigned ints, f32/f64 mean 32- and 64-bit floats/doubles, and v32/v64 mean
# 3-vectors of 32- and 64-bit floats/doubles
Types = v32, v32, u32

# Accuracies tells guppy how accurately these variables should be stored. All
# of these are done in guppy's code units: comoving Mpc/h for positions and
# comoving km/s for velocities. (Other variables aren't supported yet.) For
# integers, you must always set the accuracy to 0.
#
# I /strongly suggest/ skimming the code paper, Mansfield & Abel (2021),
# before choosing your accuracy levels, as we ran many tests on the impact of
# different accuracy levels on halo properties and on the comrpession ratios
# that different accuracies can achieve.
Accuracies = 0.001, 1, 0

###########################
# Input/Output parameters #
###########################

# Input gives the location of the input files. You can specify some variable
# that gets used in the file names by putting it in braces:
# {print-format,variable-description}. The string before the comma is a
# formatting instruction like C's printf of Python's % operator use. The string
# after the variable is either the name of the variable (e.g. "snapshot") or
# an inclusive range of numbers (e.g. 0..511). The example string below
# would describe files that looked like /path/to/input/snapdir_015/snap_015.31,
# with the first two numbers being the snapshot and the last one being the
# index of the file in the snapshot.
Input = /path/to/input/snapdir_{%03d,snaphot}/snap_{%03d,snapshot}.{%d,0..511}

# Output gives the location of the output files and is formatted identically
# to Input. You should add an "output" variable somewhere to the file name
# that contains the index of the output file within the snapshot. It's usually
# a good idea to use the same naming scheme as your normal files, except with
# '.gup' appended to the end of the files. The example string below would
# describe files that looked like /path/to/output/snapdir_015/snap_015.31.gup.
Output = /path/to/output/snapdir_{%03d,snaphot}/snap_{%03d,snapshot}.{%d,output}.gup

# Snaps lists the snapshots that you want to run guppy on. This is a
# comma-separated list of either numbers or (inclusive) particle ranges written
# as start..end. If you have corrupted snapshots, you can remove them via
# "subtraction". The example below would run Guppy on snapshots 0 to 100,
# except for the corrupted snapshot 63, and then also on snapshot 200.
Snaps = 0..100 - 63, 200

# OutputGridWdith is the width of the output grid of files in each dimension.
# particles will be evenly distributed among OutputGridWidth^3 files.
OutputGridWidth = 4

# CreateMissingDirectories tells guppy to create any directories it needs that
# don't already exist when generating output files. By default it will assume
# you had a typo in the Output variable and crash.
# CreateMissingDirectories = false

# ByteOrder specifies what byte ordering multi-byte types use in your input
# files. This can be set to whatever ordering your current machine uses via
# SystemOrder, or can be set manually to BigEndian or LittleEndian. If you
# don't know, assume LittleEndian: this is the most common ordering and guppy
# can usually catch when you're wrong and tell you.
# ByteOrder = SystemOrder

#####################
# File Type Options #
#####################

# FileType tells guppy what type the input files have. Currently the only
# supported types are Gadget-2 and LGadget-2.
FileType = LGadget-2

# IDOrder tells guppy how to map IDs onto Lagrangian space. Currently, the
# only supported ordering is ZUnigridPlusOne, the overwhelmingly most common
# ordering. In this ordering, the first particle has ID 1, the particle above
# it the z-direction in the ICs is 2, and so on.
# IDOrder = ZUnigridPlusOne

# GadgetVars gives the names of the different data blocks in your Gadget file.
# You don't need this variable if you aren't using Gadget. If you haven't done
# anything to your gadget configuration files, the example gives the correct
# names/ordering. guppy will check that some commonly-used field names have the
# right types.
# x - v32
# v - v32
# id - u32 or u64
# phi - f32
# acc - v32
# dt - f32
# GadgetVars = x, v, id

# GadgetTypes gives the types of each Gadget block, using the same naming
# conventions as the Vars variable. If you haven't done anything to your Gadget
# configuration, the example gives the right types. The most common change from
# the default is using 64-bit IDs instead of 32-bit IDs. If you have >=4
# billion particles, you're definitely using 64-bit IDs.
# GadgetTypes = v32, v32, u32

#######################
# Performance Options #
#######################

# Threads sets the number of threads per node used by guppy to compress the
# files. Most users will want to keep this to the default value of -1, which
# will use one thread per core. I'd only suggest changing this if you're on a
# shared machine without queueing and don't want to interfere with other jobs.
# Threads = -1
`
}

type WriteConfig struct {
	CompressionMethod string
	Vars, Types []string
	Accuracies []float64
	
	Input, Output string
	Snaps []string
	OutputGridWidth int64
	CreateMissingDirectories bool
	ByteOrder string
	
	FileType string
	GadgetVars, GadgetTypes []string
	IDOrder string

	Threads int64
}

func ParseWriteConfig(configName string) (*WriteConfig, error) {
	cfg := &WriteConfig{ }
	vars := config.NewConfigVars("write")
	
	vars.String(&cfg.CompressionMethod, "CompressionMethod", "")
	vars.Strings(&cfg.Vars, "Vars", []string{})
	vars.Strings(&cfg.Types, "Types", []string{})
	vars.Floats(&cfg.Accuracies, "Accuracies", []float64{})
	
	vars.String(&cfg.Input, "Input", "")
	vars.String(&cfg.Output, "Output", "")
	vars.Strings(&cfg.Snaps, "Snaps", []string{})
	vars.Int(&cfg.OutputGridWidth, "OutputGridWidth", -1)
	vars.Bool(&cfg.CreateMissingDirectories, "CreateMissingDirectories", false)
	vars.String(&cfg.ByteOrder, "ByteOrder", "SystemOrder")
	
	vars.String(&cfg.FileType, "FileType", "")
	vars.Strings(&cfg.GadgetVars, "GadgetVars", []string{"x", "v", "id"})
	vars.Strings(&cfg.GadgetTypes, "GadgetTypes",
		[]string{"v32", "v32", "u32"})
	vars.String(&cfg.IDOrder, "IDOrder", "ZUnigridPlusOne")
	vars.Int(&cfg.Threads, "Threads", -1)
	
	err := config.ReadConfig(configName, vars)
	if err != nil { return nil, err }

	return cfg, nil
}

func CheckWriteConfig(cfg *WriteConfig) error {
	// CompressionMethod
	if cfg.CompressionMethod == "" {
		return fmt.Errorf("The CompressionMethod variable was not set.")
	} else if !containsString(SupportedCompressionMethods,
		cfg.CompressionMethod) {
		return fmt.Errorf("The CompressionMethod variable was set to %s, " +
			"but only supported methods are: %s", cfg.CompressionMethod,
			SupportedCompressionMethods)
	}
	
	// Vars, Types, and Accuracies
	if len(cfg.Vars) == 0 {
		return fmt.Errorf("The Vars variable was not set.")
	} else if len(cfg.Types) == 0 {
		return fmt.Errorf("The Types variable was not set.")
	} else if len(cfg.Accuracies) == 0 {
		return fmt.Errorf("The Accuracies variable was not set.")
	} else if !sameLength([]int{len(cfg.Vars), len(cfg.Types),
		len(cfg.Accuracies)}) {
		return fmt.Errorf("Vars, Types, and Accuracies should all have the" +
			"same length, but they have lengths %d, %d, and %d.",
			len(cfg.Vars), len(cfg.Types), len(cfg.Accuracies))
	}

	if ok, i := validTypes(cfg.Types); !ok {
		return fmt.Errorf("The Types variable, %s, has '%s' at index %d, " +
			"but the only supported types are u32 u63, f32, f64, v32, and " +
			"v64.", cfg.Types, cfg.Types[i], i)
	} else if err := checkAccuracies(cfg.Vars, cfg.Types,
		cfg.Accuracies); err != nil {
		return err
	}

	// Input, Output, and Snaps
	if err := checkInputOutputSnaps(cfg); err != nil { return err }
	
	// OutputGridWidth
	if cfg.OutputGridWidth == -1 {
		return fmt.Errorf("The OutputGridWidth variable was not set")
	} else if cfg.OutputGridWidth <= 0 {
		return fmt.Errorf("The OutptuGridWidth variable must be positive, " +
			"but was set to %d", cfg.OutputGridWidth)
	} else if cfg.OutputGridWidth >= 100 {
		return fmt.Errorf("OutputGridWidth is set to %d. This means that %d " +
			"files are about to be created! guppy can't handle creating this" +
			"many files at once (and your file system may not be able to " +
			"either).", cfg.OutputGridWidth,
			cfg.OutputGridWidth*cfg.OutputGridWidth*cfg.OutputGridWidth)
	}

	// FileType
	switch cfg.FileType {
	case "Gadget-2", "LGadget-2":
		if len(cfg.GadgetVars) == 0 {
			return fmt.Errorf("The GadgetVars variable was not set, even " +
				"though the FileType is %s", cfg.FileType)
		} else if len(cfg.GadgetTypes) == 0 {
			return fmt.Errorf("The GadgetTypes variable was not set, even " +
				"though the FileType is %s", cfg.FileType)
		} else if len(cfg.GadgetVars) != len(cfg.GadgetTypes) {
			return fmt.Errorf("The GadgetVars variable has length %d, but " +
				"the GadgetTypes variable has length %d.",
				len(cfg.GadgetVars), len(cfg.GadgetTypes))
		} else if err := checkVarTypeMatch(cfg.Vars, cfg.Types, cfg.GadgetVars,
			cfg.GadgetTypes, "GadgetVars", "GadgetTypes"); err != nil {
			return fmt.Errorf(err.Error())
		}
	default:
		return fmt.Errorf("The variable FileType is set to %s, but the only " +
			"supported files types are currently Gadget-2 and LGadget-2",
			cfg.FileType)
	}

	// IDOrder
	if !containsString(SupportedIDOrders, cfg.IDOrder) {
		return fmt.Errorf("The IDOrder variable was set to %s, " +
			"but only supported orderings are: %s", cfg.IDOrder,
			SupportedIDOrders)
	}

	// Threads
	if cfg.Threads < -1 || cfg.Threads == 0 {
		return fmt.Errorf("The Threads variable was set to %d, but the " +
			"only valid values are -1 or a positive integer.", cfg.Threads)
	}

	// ByteOrder
	switch cfg.ByteOrder {
	case "SystemOrder", "BigEndian", "LittleEndian":
	default:
		return fmt.Errorf("The ByteOrder variable was set to %s, but the " +
			"only valid values are SystemOrder, LittleEndian, and BigEndian",
			cfg.ByteOrder)
	}
	
	return nil
}

func expandSnaps(snaps []string) ([]int, error) {
	out := []int{ }
	for i := range snaps {
		s, err := format.ExpandSequenceFormat(snaps[i])
		if err != nil {
			return nil, fmt.Errorf("The Snaps field at index %d, %s, could " +
				"not be parsed. %s", i, snaps[i], err.Error())
			
		}
		out = append(out, s...)
	}
	return out, nil
}

func checkInputOutputSnaps(cfg *WriteConfig) error {
	// The basic idea here is to check that a 
	
	snaps, err := expandSnaps(cfg.Snaps)
	if err != nil { return err }

	end := len(snaps) - 1
	testSnaps := []int{snaps[0], snaps[end / 2], snaps[end]}

	for _, testSnap := range testSnaps {
		inputs, err := format.ExpandFormatString(cfg.Input,
			map[string]int{"snapshot": testSnap})
		if err != nil {
			fmt.Errorf("The Input variable, %s, could not be parsed. %s",
				cfg.Input, err.Error())
		}
		
		end := len(inputs) - 1
		testInputs := []string{inputs[0], inputs[end/2], inputs[end]}

		for _, testInput := range testInputs {
			if !fileAccessible(testInput) {
				return fmt.Errorf("guppy tested a few example files that " +
					"would be generated by the Input variable, %s. The " +
					"file %s does not exist or isn't readable.", cfg.Input,
					testInput)
			}
		}

		outputMap := map[string]int{"snapshot": testSnap, "output": 0}
		outputs, err :=  format.ExpandFormatString(cfg.Output, outputMap)
		if err != nil {
			return fmt.Errorf("The Output variable, %s, could not be " +
				"parsed. %s", cfg.Output, err.Error())
		} else if len(outputs) != 1 {
			return fmt.Errorf("The Output variable should only generate " +
				"one filename for a given ouput index and snapshot, but " +
				"%d were generated. This is probably because you added " +
				"an unneeded {%%d,start..end} variable.", len(outputs))
		} else if !cfg.CreateMissingDirectories &&
			!parentDirWritable(outputs[0]) {
			return fmt.Errorf("The Output variable, %s, generated files in " +
				"the directory %s, but this directory does not exist is " +
				"not writable, or is a non-directory file. If the directory " +
				"does not exist and you would like guppy to create it for " +
				"you, rerun guppy with CreateMissingDirectories = true.")
		}
	}

	return nil
}

func fileAccessible(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func parentDirWritable(file string) bool {
	dir := path.Base(file)
	_, err := os.Stat(dir)
	if err != nil { return false }

	f, err := os.CreateTemp(dir, "tmp*")
	if err != nil { return false }
	os.Remove(f.Name())

	return true
}

func checkVarTypeMatch(smallVars, smallTypes,
	bigVars, bigTypes []string, bigVarsName, bigTypesName string) error {
	for i := range smallVars {
		j := containsStringIndex(bigVars, smallVars[i])
		if j == -1 {
			return fmt.Errorf("The variable at index %d in Vars, %s, is not " +
				"contained in %s, %s.", i, smallVars[i], bigVarsName, bigVars)
		} else if smallTypes[i] != smallTypes[j] {
			return fmt.Errorf("The type of the variable %s in Vars is %s, " +
				"but in %s it is %s.", smallVars[i], smallTypes[i],
				bigTypesName, bigTypes[j])
		}
	}
	return nil
}

func checkAccuracies(v, t []string, acc []float64) error {
	for i := range t {
		switch t[i] {
		case "u32", "u64":
			if acc[i] != 0 {
				return fmt.Errorf("The variable at index %d, %s, has Type " +
					"%s. This means that the corresponding Accuracy should " +
					"be 0, but instead it's %g.", i, v[i], t[i], acc[i])
			}
		case "f32", "f64", "v32", "v64":
			if acc[i] <= 0 {
				return fmt.Errorf("The variable at index %d, %s, has Type " +
					"%s. This means that the corresponding Accuracy should " +
					"be positive, but instead it's %g.", i, v[i], t[i], acc[i])
			}
		}
	}
	return nil
}

func validTypes(t []string) (ok bool, i int) {
	for i := range t {
		switch t[i] {
		case "f32", "f64", "u32", "u64", "v32", "v64":
		default:
			return false, i
		}
	}
	return true, -1
}

func sameLength(x []int) bool {
	if len(x) <= 1 { return true }
	x0 := x[0]
	for i := range x {
		if x[i] != x0 { return false }
	}
	return true
}

func containsString(list []string, x string) bool {
	return containsStringIndex(list, x) >= 0
}

func containsStringIndex(list []string, x string) int {
	for i := range list {
		if list[i] == x { return i }
	}
	return -1
}

func InputBuffers(hd snapio.Header, workers int) []*snapio.Buffer {
	out := make([]*snapio.Buffer, workers)
	for i := range out {
		var err error
		out[i], err = snapio.NewBuffer(hd)
		if err != nil { panic(fmt.Sprintf("Internal error: %s", err.Error())) }
	}
	return out
}

func byteOrder(cfg *WriteConfig) binary.ByteOrder {
	switch cfg.ByteOrder {
	case "LittleEndian":
		return binary.LittleEndian
	case "BigEndian":
		return binary.BigEndian
	case "SystemOrder":
		return SystemByteOrder()
	}
	panic(fmt.Sprintf("Internal error: unrecognized ByteOrder %s",
		cfg.ByteOrder))
}


type OutputBuffer struct {
	Buffer *compress.Buffer
	B []byte
	Writer *compress.Writer
}

func OutputBuffers(cfg *WriteConfig, workers int) []*OutputBuffer {
	out := make([]*OutputBuffer, workers)
	for i := range out {
		out[i] = &OutputBuffer{ compress.NewBuffer(0), []byte{ }, nil }
	}
	return out
}

func ExpandFileNames(cfg *WriteConfig) (
	snaps []int, inputs, outputs [][]string,
) {
	snaps, err := expandSnaps(cfg.Snaps)
	if err != nil { panic(fmt.Sprintf("Internal error: %s", err.Error())) }

	gw := cfg.OutputGridWidth
	nOutputs := int(gw*gw*gw)
	
	inputs, outputs = [][]string{}, [][]string{}
	for _, snap := range snaps {
		inputMap := map[string]int{ "snap": snap }
		snapInputs, err := format.ExpandFormatString(cfg.Input, inputMap)
		if err != nil { panic(fmt.Sprintf("Internal error: %s", err.Error())) }

		inputs = append(inputs, snapInputs)

		snapOutputs := make([]string, nOutputs)
		for i := 0; i < nOutputs; i++ {
			outputMap := map[string]int{ "snap": snap, "output": i }

			iSnapOutputs, err := format.ExpandFormatString(
				cfg.Output, outputMap)
			if err != nil {
				panic(fmt.Sprintf("Internal error: %s", err.Error()))
			}
			snapOutputs[i] = iSnapOutputs[0]
		}

		outputs = append(outputs, snapOutputs)
	}

	return snaps, inputs, outputs
}

func GetSnapioHeader(cfg *WriteConfig, file string) (snapio.Header, error) {
	var(
		err error
		f snapio.File
	)

	order := byteOrder(cfg)

	switch cfg.FileType {
	case "Gadget-2":
		f, err = snapio.NewGadget2Cosmological(file, cfg.GadgetVars,
			cfg.GadgetTypes, order)
	case "LGadget-2":
		f, err = snapio.NewLGadget2(file, cfg.GadgetVars,
			cfg.GadgetTypes, order)
	}

	if err != nil {
		return nil, fmt.Errorf("Cannot read %s: %s", file, err.Error())
	}

	hd, err := f.ReadHeader()
	if err != nil {
		return nil, fmt.Errorf("Cannot read %s: %s", file, err.Error())
	}
	return hd, nil
}

func RandomFileName(files [][]string) string {
	i := rand.Intn(len(files))
	j := rand.Intn(len(files[i]))
	return files[i][j]
}
