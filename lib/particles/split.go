package particles

import (
	"fmt"
	"math"
	
	"github.com/phil-mansfield/guppy/lib/snapio"
)

// Split splits the particles in a series of Files into distinct Particles
// maps, which will be compressed individually into different output files.
// It's basically just a glue function for File and SplitScheme. The "id" field
// will not be copied over, even if it is supplied.
//
// Split can encounter both external and internal errors and returns a boolean
// specifying which type of error it is returning.
func Split(
	scheme SplitScheme, files []snapio.File, vars []string,
) (p []Particles, err error, externalErr bool) {
	if len(files) == 0 {
		return nil, fmt.Errorf("Zero files were specified."), false
	}

	// Set up shared header and buffer.
	hd, err := files[0].ReadHeader()
	if err != nil { return nil, err, true }
	buf, err := snapio.NewBuffer(hd)
	if err != nil { return nil, err, false }
	
	// Create output buffers. One for each file.
	out := scheme.Buffers()

	// Arrays storing the indices particles will be moved from in the original
	// arrays and the indices they will be moved to in the arrays in out.
	from, to := make([][]int, len(out)), make([][]int, len(out))
	id := []uint64{ }


	for _, file := range files {
		// Load variables into the buffer.
		buf.Reset()
		for _, v := range vars {
			err := file.Read(v, buf)
			if err != nil { return nil, err, true }
		}

		// Convert the IDs to a standardized format without doing unneeded
		// heap allocations.
		idGeneric, err := buf.Get("id")
		if err != nil { return nil, err, false }
		id, err = standardizeIDs(idGeneric, id)
		if err != nil { return nil, err, false }

		// Compute transfer indices.
		from, to, err := scheme.Indices(id, from, to)
		if err != nil { return nil, err, false }
		
		// Copy from the buffer to a Particles map.
		for _, name := range vars {
			// Don't copy IDs, even if the user specifies it.
			if name == "id" { continue }
			
			x, err := buf.Get(name)
			if err != nil { return nil, err, false }
			field, err := NewGenericField(name, x)
			if err != nil { return nil, err, false }

			for p := range out {
				err = field.Transfer(out[p], from[p], to[p])
				if err != nil { return nil, err, false }
			}
		}
	}

	return out, nil, false
}

// StandardizeIDs standardizes an array of []uint32 or []uint64 IDs to []uint64.
// If the input is []uint64, the array is just returned. If the input is
// []uint32, the output is written to a buffer. That buffer will be trimmed
// or expanded to the correct size.
func standardizeIDs(id interface{}, buf []uint64) ([]uint64, error) {
	switch x := id.(type) {
	case []uint32:
		if n := cap(buf) - len(x); n > 0 {
			buf = append(buf, make([]uint64, n)...)
			
		}
		buf = buf[:len(x)]

		for i := range x {
			buf[i] = uint64(x[i])
		}
		
		return buf, nil
	case []uint64:
		return x, nil
	default:
		return nil, fmt.Errorf("'id' not set to 'u32' or 'u64'")
	}
}

// resizeInts resizes an int buffer to have the specified length.
func resizeInts(x []int, n int) []int {
	if n := cap(x) - n; n > 0 {
		x = append(x, make([]int, n)...)
	}
	return x[:n]
}

// SplitScheme is a strategy for splitting up a simulation's particles into
// regions that are contiguous in Lagrangian space.
type SplitScheme interface {
	// Buffers returns the Paritcles maps for each file that the simulation
	// will be broken into. These Particles maps will be initialized with the
	// correct size and names, so this information should be passed to the
	// SplitScheme instructor.
	Buffers() []Particles
	// Indices writes the indices that the given IDs should be written to.
	// from[i][j] is the index into id of a particle that should be written to
	// file i and to[i][j] is the index of that particle in the Particles
	// arrays.
	Indices(id []uint64, from, to [][]int) (fromOut, toOut [][]int, err error)
}

// UniformSplitUnigrid is a SplitScheme which splits a uniform-density grid into
// equal-sized sub-cubes. See the SplitScheme documentation for method
// descriptions.
type UniformSplitUnigrid struct {
	nAll int // Number of particles on one side for the full simulation.
	nSub int // Number of particles on one side for a sub-cube
	nCube int // Number of sub-cubes on one side.
	names, types []string
	order IDOrder
	
}

// NewUniformSplitUnigrid splits a simulation up into subgrids with nCube
// cubes on one side. vars gives the names of the fields to transfer over.
func NewUniformSplitUnigrid(
	hd snapio.Header, order IDOrder, nCube int, names []string,
) (*UniformSplitUnigrid, error) {
	nTot := hd.NTot()
	nAll := round(math.Pow(float64(nTot), 1.0/3))
	if nAll*nAll*nAll != nTot {
		return nil, fmt.Errorf("The total number of particles in the simulation is %d, but uniform grids must be perfect cubes.", nTot)
	}

	if nAll % nCube != 0 {
		return nil, fmt.Errorf("The number of sub-cubes, %d^3, doesn't evenly divide the number of particles, %d^3", nCube, nAll)
	}
	nSub := nAll / nCube

	fullNames, fullTypes := hd.Names(), hd.Types()
	types := make([]string, len(names))
	
NamesLoop:
	for i := range names {
		for j := range fullNames {
			if names[i] == fullNames[j] {
				types[i] = fullTypes[j]
				continue NamesLoop
			}
			return nil, fmt.Errorf("Could not read the variable '%s', not variable was mapped to '%s'", names[i], names[i])
		}
	}
	
	return &UniformSplitUnigrid{ nAll, nSub, nCube, names, types, order }, nil
	
}

// round rounds a float to the nearest integer.
func round(x float64) int {
	low, high := math.Floor(x), math.Ceil(x)
	if x - low < high - x {
		return int(low)
	} else {
		return int(high)
	}
}

func (g *UniformSplitUnigrid) Buffers() []Particles {
	// Create blank Fields objects for each variable.
	srcFields := Particles{ }
	for i := range g.names {
		var f Field
		switch g.types[i] {
		case "u32": f = NewUint32(g.names[i], []uint32{ })
		case "u64": f = NewUint64(g.names[i], []uint64{ })
		case "f32": f = NewFloat32(g.names[i], []float32{ })
		case "f64": f = NewFloat64(g.names[i], []float64{ })
		case "v32": f = NewVec32(g.names[i], [][3]float32{ })
		case "v64": f = NewVec64(g.names[i], [][3]float64{ })
		default: panic("'Impossible' type configuration.")
		}
		srcFields[g.names[i]] = f
	}

	// use Fields.CreateDestination() to create output fields.
	n := g.nSub*g.nSub*g.nSub
	p := make([]Particles, g.nCube*g.nCube*g.nCube)
	for i := range p {
		p[i] = Particles{ }
		for _, f := range srcFields { f.CreateDestination(p[i], n) }
	}

	return p
}

func (g *UniformSplitUnigrid) Indices(
	id []uint64, from, to [][]int,
) (fromOut, toOut [][]int, err error) {
	// Refresh 
	for i := range fromOut {
		from[i] = from[i][:0]
		to[i] = to[i][:0]
	}

	// Index vectors.
	iCube, iSub := [3]int{ }, [3]int{ }
	for i, x := range id {
		vec, level := g.order.IDToIndex(x)
		if level != 0 {
			return nil, nil, fmt.Errorf("SplitScheme is UniformSplitUnigrid, but ID %d, %x, has level %d instead of 0",
				i, x, level)
		}
		
		for k := 0; k < 3; k++ {
			if vec[k] < 0 || vec[k] >= g.nAll {
				return nil, nil, fmt.Errorf("Simulation has %d^3 particles, but ID %d, %x, is converted index %d.",
					g.nAll, i, x, vec)
			}
			iCube[k] = vec[k] / g.nSub
			iSub[k] = vec[k] - iCube[k]*g.nSub

			jCube := iCube[0] + iCube[1]*g.nCube + iCube[2]*g.nCube*g.nCube
			jSub := iSub[0] + iSub[1]*g.nSub + iSub[2]*g.nSub*g.nSub

			from[jCube] = append(from[jCube], i)
			to[jCube] = append(to[jCube], jSub)
		}
	}

	return from, to, nil
}
