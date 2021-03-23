/*package particles contains functions for */
package particles

/* This file contains functions for managing particles and their fields. */

import (
	"fmt"
)

// Particles represents the particles in a simulation or chunk of a simulation.
// It maps the name of each field (e.g. 'id', 'x', 'phi', etc.) to a Field.
type Particles map[string]Field

// Field is a generic interface around
type Field interface {
	// Len returns the length of the underlying array.
	Len() int
	// Data returns the underlying array as an interface{}.
	Data() interface{}
	// Transfer transfers data from the Field to the appropriately named field
	// in dest. Particles are transfer from the indices 'from' to the indices
	// 'to'. These indices are passed as arrays to amortize the cost of error
	// handling and type conversion.
	Transfer(dest Particles, from, to []int) error
	// CreateDestination creates output fields in p with the specified size
	// that have the correct names and types.
	CreateDestination(p Particles, n int)
}

// Type assertions
var (
	_ Field = &Uint32{ }
)

// Uint32 implements the Field interface for []uint32 data. See the Field
// interface for documentation of this struct's methods.
type Uint32 struct {
	name string
	data []uint32
}

// NewUint32 creates a field with a given name assoicated with a given array.
func NewUint32(name string, x []uint32) *Uint32 {
	return &Uint32{ name, x }
}

func (x *Uint32) Len() int { return len(x.data) }
func (x *Uint32) Data() interface{} { return x.data }

func (x *Uint32) CreateDestination(p Particles, n int) {
	p[x.name] = NewUint32(x.name, make([]uint32, n))
}

func (x *Uint32) Transfer(dest Particles, from, to []int) error {	
	destField, ok := dest[x.name]
	if !ok {
		return fmt.Errorf("Destination Particles object does not contain the field '%s'.", x.name)
	}
		
	destData, ok := destField.Data().([]uint32)
	if !ok {
		return fmt.Errorf("Field '%s' in destination Particles object does not have []uint32 type, as expected.", x.name)
	}

	if len(from) != len(to) {
		return fmt.Errorf("'from' index array has length %d, but 'to' has length %d.", len(from), len(to))
	}
	
	for i := range from {
		destData[to[i]] = x.data[from[i]]
	}

	return nil
}
