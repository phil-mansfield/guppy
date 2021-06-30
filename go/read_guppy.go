/*package read_guppy provides several functions for reading .gup files.*/
package read_guppy

import (
	"fmt"
	"sync"
)

var (
	workers []*worker
	mutexes []*sync.Mutex
)

// Header contains header information about a given .gup file.
type Header struct {
	// OriginalHeader is the original header of the one of the simulation 
	OriginalHeader []byte
	// Names gives the names of all the variables stored in the file.
	// Types give the types of these variables. "u32"/"u64" give 32-bit and
	// 64-bit integers, respectively, anf "f32"/"f64" give 32-bit and
	// 64-bit floats, respectively.
	Names, Types []string
	// N and Ntot give the number of particles in the file and in the
	// total simulation, respectively.
	N, Ntot int
	// Z, OmegaM, H100, L, and Mass give the redshift, Omega_m,
	// H0 / (100 km/s/Mpc), box width in comoving Mpc/h, and particle
	// mass in Msun/h.
	Z, OmegaM, H100, L, Mass float64

}

// worker contains various buffers which prevent excess heap allocations
// when reading the file.
type worker struct {

}

// newWorker creates a blank worker object that can be used for reading.
func newWorker() *worker {
	
}

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
	panic("Current version of Guppy files doesn't contain Header information.")
	return &Header{ }
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
func ReadVar(fileName, name string, workerID int, interface{} buf) {
	worker := getWorker(workerID)

	finishWorker(workerID)
}

// InitWorkers
func InitWorkers(nWorkers int) {
	workers = make([]*worker, nWorkers)
	mutexes = make([]*sync.Mutex, nWorkers)

	for i := 0; i < nWorkers; i++ {
		workers[i] = newWorker()
		mutexes[i] = &sync.Mutex{ }
	}
}