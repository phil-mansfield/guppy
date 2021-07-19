/*package read_guppy provides several functions for reading .gup files.*/
package read_guppy

import (
	"github.com/phil-mansfield/guppy/lib/compress"

	"fmt"
	"sync"
)

var (
	workers []*worker
	mutexes []*sync.Mutex
)

// Header contains header information about a given .gup file.
type Header struct {
	// OriginalHeader is the original header of the one of the original
	// simulation files.
	OriginalHeader []byte
	// Names gives the names of all the variables stored in the file.
	// Types give the types of these variables. "u32"/"u64" give 32-bit and
	// 64-bit unsigned integers, respectively, and "f32"/"f64" give 32-bit
	// and 64-bit floats, respectively.
	Names, Types []string
	// N and NTot give the number of particles in the file and in the
	// total simulation, respectively.
	N, NTot int64
	// Span gives the dimensions of the slab of particles in the file.
	// Span[0], Span[1], and Span[2] are the x-, y-, and z- dimensions.
	Span [3]int64
	// Z, OmegaM, H100, L, and Mass give the redshift, Omega_m,
	// H0 / (100 km/s/Mpc), box width in comoving Mpc/h, and particle
	// mass in Msun/h, respectively.
	Z, OmegaM, H100, L, Mass float64

}

// worker contains various buffers which prevent excess heap allocations
// when reading the file.
type worker struct {
	buf *compress.Buffer
	midBuf []byte

}

// newWorker creates a blank worker object that can be used for reading.
func newWorker() *worker {
	return &worker{ compress.NewBuffer(0), []byte{ } }
}

// getWorker retrieves the buffer space associated with the given
// worker ID and handles mutex ownership. If -1 is passed to this function,
// a new worker is allocated.
func getWorker(workerID int) *worker {
	if workerID == - 1 {
		return newWorker()
	} else if workerID < -1 || workerID >= len(workers) {
		panic(fmt.Sprintf("Cannot use worker %d for nWorkers = %d",
			workerID, len(workers)))
	} else {
		mutexes[workerID].Lock()
		return workers[workerID] 

	}
}

func finishWorker(workerID int) {
	if workerID != -1 {
		mutexes[workerID].Unlock()
	}
}

// ReadHeader returns the header of a given file.
func ReadHeader(fileName string) *Header {
	worker := getWorker(-1)
	rd, err := compress.NewReader(fileName, worker.buf, worker.midBuf)
	if err != nil {
		panic(fmt.Sprintf("Guppy encountered an error while opening and " + 
			"initializing the file: %s", err.Error()))
	}
	defer rd.Close()

	rhd := &rd.Header
	return &Header{
		rhd.OriginalHeader,
		rhd.Names, rhd.Types,
		rhd.N, rhd.NTot, rhd.Span,
		rhd.Z, rhd.OmegaM, rhd.H100, rhd.L, rhd.Mass,
	}
}

// ReadVar reads a variable with a given name from a givne file. If you
// and to use one of the pre-allocated workers, you should give the integer
// ID of that workers (i.e. in the range [0, nWorkers). ReadVar uses
// mutexes to make sure that same worker isn't being used simultaneously,
// so feel free to throw a zillion threads at the same worker. If you
// don't care about heap space, just set workerID to -1. The last argument
// is a buffer with length Header.N where the variable will be written to.
//
// For vector quantities, you can either load each component one by one
// (e.g. "x[0]", "x[1]", etc.) and supply a []float32 or []float64 buffer,
// or you can get the full vector (e.g. "x") and supply a [][3]float32 or
// [][3]float64.
//
// The variable "id" is implicitly contained in every .gup file and can be
// read into a []uint64 array.
func ReadVar(fileName, name string, workerID int, buf interface{}) {
	// Allocated underlying buffers.
	worker := getWorker(workerID)
	rd, err := compress.NewReader(fileName, worker.buf, worker.midBuf)
	if err != nil {
		panic(fmt.Sprintf("Guppy encountered an error while opening and " + 
			"initializing the file: %s", err.Error()))
	}

	defer rd.Close()

	// Handle generic variables.
	switch x := buf.(type) {
	case [][3]float32: readVec32(rd, name, x)
	case [][3]float64: readVec64(rd, name, x)
	case []float32: readFloat32(rd, name, x)
	case []float64: readFloat64(rd, name, x)
	case []uint32: readUint32(rd, name, x)
	case []uint64: readUint64(rd, name, x)
	default:
		panic("The buffer passed to ReadVar is not [][3]float32, " + 
			"[][3]float64, []float32, []float64, []uint32, or []uint64.")
	}

	worker.midBuf = rd.ReuseMidBuf()

	finishWorker(workerID)
}

func readVec32(
	rd *compress.Reader, name string, buf [][3]float32,
) {
	expTypeName := "f32"
	for dim := 0; dim < 3; dim++ {
		typeName := checkName(&rd.Header, name)
		if typeName != expTypeName {
			panic(fmt.Sprintf("Field '%s' has type '%s', but the supplied " + 
				"buffer is type '%s'. The file's Header struct contains " + 
				"information on the types of fields.", name,
				typeName, expTypeName))
		}

		field, err := rd.ReadField(name)
		if err != nil {
			panic(fmt.Sprintf("Guppy encountered error while reading the " + 
				"field '%s': %s", name, err.Error()))
		}
		x, ok := field.Data().([]float32)
		if !ok {
			panic(fmt.Sprintf("Internal type error: Field '%s' has type " + 
				"'%s', but this is not the type returned by ReadField().",
				name, expTypeName))
		}

		if len(x) != len(buf) {
			panic(fmt.Sprintf("Length of the buffer supplied for field " + 
				"'%s' is %d, but the field has %d elements.", name,
				len(buf), len(x)))
		}

		for i := range x { buf[i][dim] = x[i] }
	}
}

func readVec64(
	rd *compress.Reader, name string, buf [][3]float64,
) {
	expTypeName := "f64"
	for dim := 0; dim < 3; dim++ {
		typeName := checkName(&rd.Header, name)
		if typeName != expTypeName {
			panic(fmt.Sprintf("Field '%s' has type '%s', but the supplied " + 
				"buffer is type '%s'. The file's Header struct contains " + 
				"information on the types of fields.", name,
				typeName, expTypeName))
		}

		field, err := rd.ReadField(name)
		if err != nil {
			panic(fmt.Sprintf("Guppy encountered error while reading the " + 
				"field '%s': %s", name, err.Error()))
		}
		x, ok := field.Data().([]float64)
		if !ok {
			panic(fmt.Sprintf("Internal type error: Field '%s' has type " + 
				"'%s', but this is not the type returned by ReadField().",
				name, expTypeName))
		}

		if len(x) != len(buf) {
			panic(fmt.Sprintf("Length of the buffer supplied for field " + 
				"'%s' is %d, but the field has %d elements.", name,
				len(buf), len(x)))
		}

		for i := range x { buf[i][dim] = x[i] }
	}
}

func readFloat32(
	rd *compress.Reader, name string, buf []float32,
) {
	expTypeName := "f32"
	typeName := checkName(&rd.Header, name)
	if typeName != expTypeName {
		panic(fmt.Sprintf("Field '%s' has type '%s', but the supplied buffer " + 
			"is type '%s'. The file's Header struct contains information on " + 
			"the types of fields.", name, typeName, expTypeName))
	}

	field, err := rd.ReadField(name)
	if err != nil {
		panic(fmt.Sprintf("Guppy encountered error while reading the field " + 
			"'%s': %s", name, err.Error()))
	}
	x, ok := field.Data().([]float32)
	if !ok {
		panic(fmt.Sprintf("Internal type error: Field '%s' has type '%s', " + 
			"but this is not the type returned by ReadField().",
			name, expTypeName))
	}

	if len(x) != len(buf) {
		panic(fmt.Sprintf("Length of the buffer supplied for field '%s' " + 
			"is %d, but the field has %d elements.", name, len(buf), len(x)))
	}

	for i := range x { buf[i] = x[i] }
}

func readFloat64(
	rd *compress.Reader, name string, buf []float64,
) {
	expTypeName := "f64"
	typeName := checkName(&rd.Header, name)
	if typeName != expTypeName {
		panic(fmt.Sprintf("Field '%s' has type '%s', but the supplied buffer " + 
			"is type '%s'. The file's Header struct contains information on " + 
			"the types of fields.", name, typeName, expTypeName))
	}

	field, err := rd.ReadField(name)
	if err != nil {
		panic(fmt.Sprintf("Guppy encountered error while reading the field " + 
			"'%s': %s", name, err.Error()))
	}
	x, ok := field.Data().([]float64)
	if !ok {
		panic(fmt.Sprintf("Internal type error: Field '%s' has type '%s', " + 
			"but this is not the type returned by ReadField().",
			name, expTypeName))
	}

	if len(x) != len(buf) {
		panic(fmt.Sprintf("Length of the buffer supplied for field '%s' " + 
			"is %d, but the field has %d elements.", name, len(buf), len(x)))
	}

	for i := range x { buf[i] = x[i] }
}

func readUint32(
	rd *compress.Reader, name string, buf []uint32,	
) {
	expTypeName := "u32"
	typeName := checkName(&rd.Header, name)
	if typeName != expTypeName {
		panic(fmt.Sprintf("Field '%s' has type '%s', but the supplied buffer " + 
			"is type '%s'. The file's Header struct contains information on " + 
			"the types of fields.", name, typeName, expTypeName))
	}

	field, err := rd.ReadField(name)
	if err != nil {
		panic(fmt.Sprintf("Guppy encountered error while reading the field " + 
			"'%s': %s", name, err.Error()))
	}
	x, ok := field.Data().([]uint32)
	if !ok {
		panic(fmt.Sprintf("Internal type error: Field '%s' has type '%s', " + 
			"but this is not the type returned by ReadField().",
			name, expTypeName))
	}

	if len(x) != len(buf) {
		panic(fmt.Sprintf("Length of the buffer supplied for field '%s' " + 
			"is %d, but the field has %d elements.", name, len(buf), len(x)))
	}

	for i := range x { buf[i] = x[i] }
}

func readUint64(
	rd *compress.Reader, name string, buf []uint64,	
) {
	expTypeName := "u64"
	typeName := checkName(&rd.Header, name)
	if typeName != expTypeName {
		panic(fmt.Sprintf("Field '%s' has type '%s', but the supplied buffer " + 
			"is type '%s'. The file's Header struct contains information on " + 
			"the types of fields.", name, typeName, expTypeName))
	}

	field, err := rd.ReadField(name)
	if err != nil {
		panic(fmt.Sprintf("Guppy encountered error while reading the field " + 
			"'%s': %s", name, err.Error()))
	}
	x, ok := field.Data().([]uint64)
	if !ok {
		panic(fmt.Sprintf("Internal type error: Field '%s' has type '%s', " + 
			"but this is not the type returned by ReadField().",
			name, expTypeName))
	}

	if len(x) != len(buf) {
		panic(fmt.Sprintf("Length of the buffer supplied for field '%s' " + 
			"is %d, but the field has %d elements.", name, len(buf), len(x)))
	}

	for i := range x { buf[i] = x[i] }
}

func checkName(hd *compress.Header, name string) string {
	for i := range hd.Names {
		if hd.Names[i] == name { return hd.Types[i] }
	}

	panic(fmt.Sprintf("File does not contain any variable named '%d'. " + 
		"It only contains the variables %s.", name, hd.Names))
}

// InitWorkers allocates space for nWorkers workers which can be run
// simultaneously by different threads.
func InitWorkers(nWorkers int) {
	workers = make([]*worker, nWorkers)
	mutexes = make([]*sync.Mutex, nWorkers)

	for i := 0; i < nWorkers; i++ {
		workers[i] = newWorker()
		mutexes[i] = &sync.Mutex{ }
	}
}