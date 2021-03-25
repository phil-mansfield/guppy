/*package particles contains functions for manipulating particles with generic
fields and ID-orderings.*/
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
	_ Field = &Uint64{ }
	_ Field = &Float32{ }
	_ Field = &Float64{ }
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

// Uint64 implements the Field interface for []uint32 data. See the Field
// interface for documentation of this struct's methods.
type Uint64 struct {
	name string
	data []uint64
}

// NewUint64 creates a field with a given name assoicated with a given array.
func NewUint64(name string, x []uint64) *Uint64 {
	return &Uint64{ name, x }
}

func (x *Uint64) Len() int { return len(x.data) }
func (x *Uint64) Data() interface{} { return x.data }

func (x *Uint64) CreateDestination(p Particles, n int) {
	p[x.name] = NewUint64(x.name, make([]uint64, n))
}

func (x *Uint64) Transfer(dest Particles, from, to []int) error {	
	destField, ok := dest[x.name]
	if !ok {
		return fmt.Errorf("Destination Particles object does not contain the field '%s'.", x.name)
	}
		
	destData, ok := destField.Data().([]uint64)
	if !ok {
		return fmt.Errorf("Field '%s' in destination Particles object does not have []uint64 type, as expected.", x.name)
	}

	if len(from) != len(to) {
		return fmt.Errorf("'from' index array has length %d, but 'to' has length %d.", len(from), len(to))
	}
	
	for i := range from {
		destData[to[i]] = x.data[from[i]]
	}

	return nil
}

// Float32 implements the Field interface for []float32 data. See the Field
// interface for documentation of this struct's methods.
type Float32 struct {
	name string
	data []float32
}

// NewFloat32 creates a field with a given name assoicated with a given array.
func NewFloat32(name string, x []float32) *Float32 {
	return &Float32{ name, x }
}

func (x *Float32) Len() int { return len(x.data) }
func (x *Float32) Data() interface{} { return x.data }

func (x *Float32) CreateDestination(p Particles, n int) {
	p[x.name] = NewFloat32(x.name, make([]float32, n))
}

func (x *Float32) Transfer(dest Particles, from, to []int) error {	
	destField, ok := dest[x.name]
	if !ok {
		return fmt.Errorf("Destination Particles object does not contain the field '%s'.", x.name)
	}
		
	destData, ok := destField.Data().([]float32)
	if !ok {
		return fmt.Errorf("Field '%s' in destination Particles object does not have []float32 type, as expected.", x.name)
	}

	if len(from) != len(to) {
		return fmt.Errorf("'from' index array has length %d, but 'to' has length %d.", len(from), len(to))
	}
	
	for i := range from {
		destData[to[i]] = x.data[from[i]]
	}

	return nil
}

// Float64 implements the Field interface for []float32 data. See the Field
// interface for documentation of this struct's methods.
type Float64 struct {
	name string
	data []float64
}

// NewFloat64 creates a field with a given name assoicated with a given array.
func NewFloat64(name string, x []float64) *Float64 {
	return &Float64{ name, x }
}

func (x *Float64) Len() int { return len(x.data) }
func (x *Float64) Data() interface{} { return x.data }

func (x *Float64) CreateDestination(p Particles, n int) {
	p[x.name] = NewFloat64(x.name, make([]float64, n))
}

func (x *Float64) Transfer(dest Particles, from, to []int) error {	
	destField, ok := dest[x.name]
	if !ok {
		return fmt.Errorf("Destination Particles object does not contain the field '%s'.", x.name)
	}
		
	destData, ok := destField.Data().([]float64)
	if !ok {
		return fmt.Errorf("Field '%s' in destination Particles object does not have []float64 type, as expected.", x.name)
	}

	if len(from) != len(to) {
		return fmt.Errorf("'from' index array has length %d, but 'to' has length %d.", len(from), len(to))
	}
	
	for i := range from {
		destData[to[i]] = x.data[from[i]]
	}

	return nil
}

// Vec32 implements the Field interface for [][3]float32 data. See the Field
// interface for documentation of this struct's methods.
type Vec32 struct {
	dimNames [3]string
	data [][3]float32
}

// NewVec32 creates a field with a given name assoicated with a given array.
func NewVec32(name string, x [][3]float32) *Vec32 {
	dimNames := [3]string{ }
	for dim := range dimNames {
		dimNames[dim] = fmt.Sprintf("%s[%d]", name, dim)
	}
	return &Vec32{ dimNames, x }
}

func (x *Vec32) Len() int { return len(x.data) }
func (x *Vec32) Data() interface{} { return x.data }

func (x *Vec32) CreateDestination(p Particles, n int) {
	for dim := range x.dimNames {
		p[x.dimNames[dim]] = NewFloat32(x.dimNames[dim], make([]float32, n))
	}
}

func (x *Vec32) Transfer(dest Particles, from, to []int) error {
	if len(from) != len(to) {
		return fmt.Errorf("'from' index array has length %d, but 'to' has length %d.", len(from), len(to))
	}
	
	for dim := range x.dimNames {
		destField, ok := dest[x.dimNames[dim]]
		if !ok {
			return fmt.Errorf("Destination Particles object does not contain the field '%s'.", x.dimNames[dim])
		}
		
		destData, ok := destField.Data().([]float32)
		if !ok {
			return fmt.Errorf("Field '%s' in destination Particles object does not have []float32 type, as expected.", x.dimNames[dim])
		}


		for i := range from {
			destData[to[i]] = x.data[from[i]][dim]
		}
	}

	return nil
}

// Vec64 implements the Field interface for [][3]float64 data. See the Field
// interface for documentation of this struct's methods.
type Vec64 struct {
	dimNames [3]string
	data [][3]float64
}

// NewVec64 creates a field with a given name assoicated with a given array.
func NewVec64(name string, x [][3]float64) *Vec64 {
	dimNames := [3]string{ }
	for dim := range dimNames {
		dimNames[dim] = fmt.Sprintf("%s[%d]", name, dim)
	}
	return &Vec64{ dimNames, x }
}

func (x *Vec64) Len() int { return len(x.data) }
func (x *Vec64) Data() interface{} { return x.data }

func (x *Vec64) CreateDestination(p Particles, n int) {
	for dim := range x.dimNames {
		p[x.dimNames[dim]] = NewFloat64(x.dimNames[dim], make([]float64, n))
	}
}

func (x *Vec64) Transfer(dest Particles, from, to []int) error {
	if len(from) != len(to) {
		return fmt.Errorf("'from' index array has length %d, but 'to' has length %d.", len(from), len(to))
	}
	
	for dim := range x.dimNames {
		destField, ok := dest[x.dimNames[dim]]
		if !ok {
			return fmt.Errorf("Destination Particles object does not contain the field '%s'.", x.dimNames[dim])
		}
		
		destData, ok := destField.Data().([]float64)
		if !ok {
			return fmt.Errorf("Field '%s' in destination Particles object does not have []float64 type, as expected.", x.dimNames[dim])
		}


		for i := range from {
			destData[to[i]] = x.data[from[i]][dim]
		}
	}

	return nil
}
