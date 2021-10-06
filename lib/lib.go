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
	
	"reflect"
	"unsafe"

	"github.com/phil-mansfield/guppy/lib/config"
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
	X, V [3]float32
	ID uint64 
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
# anything to your gadget configuration files, the example  gives the correct
# names/ordering. You're free to change these names to whatever you want, but
# (1) make sure to make the same changes to the Vars varaible, (2) don't use
# any braces in your variable names, since guppy's internal naming conventions
# use braces to separate user-create named from guppy-created annotations, and
# (3) the 'id' and 'x' variable names are handled specially.
# GadgetVars = x, v, id

# GadgetTypes gives the types of each Gadget block, using the same naming
# conventions as the Vars variable. If you haven't done anything to your Gadget
# configuration, the example gives the right types. The most common change from
# the default is using 64-bit IDs instead of 32-bit IDs. If you have >2 billion
# particles, you're definitely using 64-bit IDs.
# GadgetTypes = v32, v32, u32
`
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

	vars.String(&cfg.FileType, "FileType", "")
	vars.Strings(&cfg.GadgetVars, "GadgetVars", []string{"x", "v", "id"})
	vars.Strings(&cfg.GadgetTypes, "GadgetTypes",
		[]string{"v32", "v32", "u32"})
	vars.String(&cfg.IDOrder, "IDOrder", "ZUnigridPlusOne")

	err := config.ReadConfig(configName, vars)
	if err != nil { return nil, err }

	return cfg, nil
}

type WriteConfig struct {
	CompressionMethod string
	Vars, Types []string
	Accuracies []float64
	
	Input, Output string
	Snaps []string
	OutputGridWidth int64
	CreateMissingDirectories bool

	FileType string
	GadgetVars, GadgetTypes []string
	IDOrder string
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

	// We don't check Input, Output, and Snaps here. That gets done when those
	// variables are expanded.
	
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

	if !containsString(SupportedIDOrders, cfg.IDOrder) {
		return fmt.Errorf("The IDOrder variable was set to %s, " +
			"but only supported orderings are: %s", cfg.IDOrder,
			SupportedIDOrders)
	}

	return nil
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
