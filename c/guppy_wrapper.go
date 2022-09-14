package main

// #include "read_guppy.h"
import "C"
import (
	read_guppy "github.com/phil-mansfield/guppy/go"
	"reflect"
	"unsafe"
	"fmt"
)

const maxArraySize = 1<<14

//export ReadHeader
func ReadHeader(fileName *C.char) *C.Guppy_Header {
	// Turns out that allocating nested arrays in C memory from Go is a
	// little complicated, ha ha...

	goFileName := C.GoString(fileName)
	goHd := read_guppy.ReadHeader(goFileName)

	var pointer *C.char
	pointerSize := (C.ulong)(unsafe.Sizeof((*C.char)(pointer)))

	headerSize := (C.ulong)(unsafe.Sizeof(C.Guppy_Header{ }))
	cHd := (*C.Guppy_Header)(C.malloc(headerSize))

	// Handle the Names and Types parameters.
	nVars := (C.ulong)(len(goHd.Names))
	cHd.NVars = (C.int64_t)(nVars)
	cHd.Names = (**C.char)(C.malloc(nVars*pointerSize))
	cHd.Types = (**C.char)(C.malloc(nVars*pointerSize))
	cHd.Szies = (*C.int64_t)(C.malloc(nVars*pointerSize))

	n := len(goHd.Names)
	// Need to convert the C pointer to a Go slice. The idea here
	// is to convert to a fixed-size Go /array/, then convert that
	// to a slice.
	cNames := (*[maxArraySize]*C.char)(unsafe.Pointer(cHd.Names))[:n:n]
	cTypes := (*[maxArraySize]*C.char)(unsafe.Pointer(cHd.Types))[:n:n]

	for i := range goHd.Names {
		cNames[i] = C.CString(goHd.Names[i])
		cTypes[i] = C.CString(goHd.Types[i])
		cHd.Sizes[i] = (C.int64_t)(goHd.Sizes[i])
	}

	//Handle all the (much simpler!) header properties
	cHd.OriginalHeader = (*C.char)(C.CBytes(goHd.OriginalHeader))
	cHd.OriginalHeaderLength = (C.int64_t)(len(goHd.OriginalHeader))
	cHd.N = (C.int64_t)(goHd.N)
	cHd.NTot = (C.int64_t)(goHd.NTot)
	cHd.Span = [3]C.int64_t{
		C.int64_t(goHd.Span[0]), C.int64_t(goHd.Span[1]),
		C.int64_t(goHd.Span[2]),
	}
	cHd.Z = (C.double)(goHd.Z)
	cHd.OmegaM = (C.double)(goHd.OmegaM)
	cHd.OmegaL = (C.double)(goHd.OmegaL)
	cHd.H100 = (C.double)(goHd.H100)
	cHd.L = (C.double)(goHd.L)
	cHd.Mass = (C.double)(goHd.Mass)

	return cHd
}

//export ReadVar
func ReadVar(fileName, varName *C.char, workerID C.int, out unsafe.Pointer) {
	goFileName, goVarName := C.GoString(fileName), C.GoString(varName)
	hd := read_guppy.ReadHeader(goFileName)

	typeString := getTypeString(hd, goVarName)
	buf := createBuffer(out, int(hd.N), typeString)

	read_guppy.ReadVar(goFileName, goVarName, int(workerID), buf)
}

func getTypeString(hd *read_guppy.Header, varName string) string {
	if varName == "[RockstarParticle]" { return "rockstar" }

	for i := range hd.Names {
		if hd.Names[i] == varName {
			return hd.Types[i]
		}
		if varName + "[0]" == hd.Names[i] {
			switch hd.Types[i] {
			case "f32": return "v32"
			case "f64": return "v64"
			default:
				panic(fmt.Sprintf("Impossible file configuration: the " + 
					"variable '%s' has type '%s'.", hd.Names[i], hd.Types[i]))
			}
		}
	}
	panic(fmt.Sprintf("The file does not have a variable named " + 
		"'%s'. It only has the variables %s", varName, hd.Names))
}

func createBuffer(ptr unsafe.Pointer, n int, typeString string) interface{} {
	var buf interface{}

	switch typeString {
	case "f32": 
		// Here we get an empty slice and reassign its header fields to 
		// point to the data behind ptr.
		slice := []float32{ } 
		hd := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
		hd.Data, hd.Len, hd.Cap = uintptr(ptr), n, n
		buf = slice
	case "f64":
		slice := []float64{ } 
		hd := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
		hd.Data, hd.Len, hd.Cap = uintptr(ptr), n, n
		buf = slice
	case "u32":
		slice := []uint32{ } 
		hd := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
		hd.Data, hd.Len, hd.Cap = uintptr(ptr), n, n
		buf = slice
	case "u64":
		slice := []uint64{ } 
		hd := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
		hd.Data, hd.Len, hd.Cap = uintptr(ptr), n, n
		buf = slice
	case "v32":
		slice := [][3]float32{ } 
		hd := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
		hd.Data, hd.Len, hd.Cap = uintptr(ptr), n, n
		buf = slice
	case "v64":
		slice := [][3]float64{ } 
		hd := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
		hd.Data, hd.Len, hd.Cap = uintptr(ptr), n, n
		buf = slice
	case "rockstar":
		slice := []read_guppy.RockstarParticle{ } 
		hd := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
		hd.Data, hd.Len, hd.Cap = uintptr(ptr), n, n
		buf = slice
	default:
		panic(fmt.Sprintf("Unrecognized type string: '%s'", typeString))
	} 

	return buf
}

//export InitWorkers
func InitWorkers(n int) {
	read_guppy.InitWorkers(int(n))
}

func main() { }
