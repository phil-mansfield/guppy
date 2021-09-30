package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"encoding/binary"
	
	read_guppy "github.com/phil-mansfield/guppy/go"
	"github.com/phil-mansfield/guppy/lib"
)

func main() {
	if len(os.Args) <= 1 { ModeError() }

	mode := os.Args[1]
	flags := os.Args[2:]
	
	switch mode {
	case "read": Read(flags)
	case "write": Write(flags)
	case "write_server": WriteServer(flags)
	default:
		ModeError()
	}
}

func Read(flags []string) {
	set := flag.NewFlagSet("read", flag.ContinueOnError)
	filePtr := set.String("file", "",
		"The name of the guppy file to read.")
	varStringPtr := set.String("vars", "",
		"The name of the variables that will be read. Should be a comma-separated list.")
	err := set.Parse(flags)
	file, varString := *filePtr, *varStringPtr
	vars := SplitCommaList(varString)
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	hd, err := ReadHeader(file, vars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	
	err = PipeDataToStdout(file, hd, vars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

func SplitCommaList(vars string) []string {
	tokens := strings.Split(vars, ",")
	for i := range tokens {
		tokens[i] = strings.Trim(tokens[i], " ")
	}
	
	out := []string{ }
	for i := range tokens {
		if len(tokens[i]) > 0 { out = append(out, tokens[i]) }
	}
	
	return out
}

func ReadHeader(file string, vars []string) (
	hd *read_guppy.Header, err error) {
	if file == "" {
		return nil, fmt.Errorf("Must set the 'file' flag to run guppy in " +
			"read mode. Call 'guppy read --help' for flag descriptions.")
	} else if len(vars) == 0 {
		return nil, fmt.Errorf("Must set the 'vars' flag to run guppy in " +
			"read mode. Call 'guppy read --help' for flag descriptions.")
	}
	
	if info, err := os.Stat(file); err != nil {
		return nil, fmt.Errorf("Guppy could not open %s: " +
			"%s", file, err.Error())
	} else if info.IsDir() {
		return nil, fmt.Errorf("Guppy could not open %s: " +
			"it's a directory.", file)
	}

	// Need to go through these hoops because ReadHeader panics on error. This
	// is the right call if you're calling those functions from C, but is
	// annoying if you want to handle it cleanly.
	err = nil
	defer func() {
		if panicData := recover(); panicData != nil {
			panicMsg, ok := panicData.(string)
			if !ok {
				panic(fmt.Sprintf("Internal error: read_guppy panicked " +
					"with something other than a string: %v", panicData))
			}
			err = fmt.Errorf(panicMsg)
		}
	}()

	hd = read_guppy.ReadHeader(file)

	for i, v := range vars {
		if !IsValidVar(v, hd) {
			return nil, fmt.Errorf("The %s requested variable, '%s', is not " +
				"in the guppy file %s. This file can only read the " +
				"variables %s, as well as derived variables, like " +
				"'{RockstarParticle}', if the prerequisite variables exist " +
				"in the file.", OrderString(i+1), v,
				file, ArrayToCommaList(hd.Names))
		}
	}

	return hd, nil
}

func IsValidVar(v string, hd *read_guppy.Header) bool {
	hasV, hasX := false, false
	for i := range hd.Names {
		if hd.Names[i] == "x{0}" { hasX = true }
		if hd.Names[i] == "v{0}" { hasV = true }
	}

	if hasV && hasX && v == "{RockstarParticle}" { return true }

	for i := range hd.Names {
		if v == hd.Names[i] {
			return true
		}
		j := strings.Index(hd.Names[i], "{0}")
		if j >= 0 && hd.Names[i][:j] == v {
			return true
		}
	}

	return false
}

func OrderString(n int) string {
	switch n % 10 {
	case 1: return fmt.Sprintf("%dst", n)
	case 2: return fmt.Sprintf("%dnd", n)
	case 3: return fmt.Sprintf("%drd", n)
	}
	return fmt.Sprintf("%dth", n)
}

func ArrayToCommaList(x []string) string {
	return strings.Join(x, ", ")
}

func PipeDataToStdout(
	file string, hd *read_guppy.Header, vars []string,
) (err error) {
	// As state above: read_guppy.go panics instead of returning errors to
	// make C-users' lives easier. We need to convert those back into
	// errors with this defer + recover().
	defer func() {
		if panicData := recover(); panicData != nil {
			panicMsg, ok := panicData.(string)
			if !ok {
				panic(fmt.Sprintf("Internal error: read_guppy panicked " +
					"with something other than a string: %v", panicData))
			}
			err = fmt.Errorf(panicMsg)
		}
	}()


	read_guppy.InitWorkers(1)
	err = WriteHeader(hd, os.Stdout)
	if err != nil {
		return fmt.Errorf("Could not write header: %s", err.Error())
	}

	for i := range vars {
		buf := AllocateBuffer(vars[i], hd)
		read_guppy.ReadVar(file, vars[i], 0, buf)
		err = lib.WriteAsBytes(os.Stdout, buf)
		if err != nil {
			return fmt.Errorf("Could not write %s: %s", vars[i], err.Error())
		}
	}

	return nil
}

func AllocateBuffer(v string, hd *read_guppy.Header) interface{} {
	typeString := ""	
	for i := range hd.Names {
		if v == hd.Names[i] {
			typeString = hd.Types[i]
			break
		}
		j := strings.Index(hd.Names[i], "{0}")
		if j >= 0 && v == hd.Names[i][:j] {
			typeString = "v" + hd.Types[i][1:]
			break
		}
	}
	if v == "{RockstarParticle}" {
		typeString = "{RockstarParticle}"
	}

	switch typeString {
	case "u32": return make([]uint32, hd.N)
	case "u64": return make([]uint64, hd.N)
	case "f32": return make([]float32, hd.N)
	case "f64": return make([]float64, hd.N)
	case "v32": return make([][3]float32, hd.N)
	case "v64": return make([][3]float64, hd.N)
	case "{RockstarParticle}": return make([]lib.RockstarParticle, hd.N)
	}

	panic(fmt.Sprintf("Internal error: the variable '%s' passed correctness " +
		"checks, but wasn't assigned a type string.", v))
}

type OutputHeader struct {
    Version, Format uint64
    N, NTot int64
    Span, Origin, TotalSpan [3]int64
    Z, OmegaM, OmegaL, H100, L, Mass float64
}

func WriteHeader(hd *read_guppy.Header, f *os.File) error {
    ohd := &lib.PipeHeader{
        lib.Version, lib.RockstarFormatCode, hd.N, hd.NTot,
		hd.Span, hd.Offset, hd.TotalSpan,
        hd.Z, hd.OmegaM, hd.OmegaL, hd.H100, hd.L, hd.Mass,
    }
	return binary.Write(f, lib.SystemByteOrder(), ohd)
}


func Write(flags []string) {
	panic("NYI")
}

func WriteServer(flags []string) {
	panic("NYI")
}

func ModeError() {
	fmt.Fprintf(os.Stderr,
`guppy requires at least a valid argument telling it what mode to run.
Valid modes are:
            read - reads particles from a file and writes them to stdout.
    write_server - manages writer processes and helps them transfer particles
                   between themselves.
           write - writes particles to disk that it's either read from files on
                   disk, or that have been written to its stdin.
Run "./guppy <mode_name> --help to print help information about what flags a
particular mode takes.%s`, "\n")
	os.Exit(1)
}
