package lib

/* This file contains functions for managing particles and their fields. */

import (
	"github.com/phil-mansfield/guppy/lib/snapio"
)

// Particles represents the particles in a simulation or chunk of a simulation.
// It maps the name of each field (e.g. ID, x0, x1, v0, etc.) to a Field
// interface that represents the data associated with that field.
type Particles map[string]Field

// Field is a generic interface for the data associated with a field. It is
// implemented by the different supported primatives (float64, uint32, etc.)
type Field interface {
}

// CollectParticles collects particles from all the files into a single set of
// arrays, ordered by particle ID. It also returns the header of the first file.
// It is not used in mpi_guppy.
func CollectParticles(args *Args, snap int) (snapio.Header, Particles) {
	panic("NYI")
}

// SplitBuffer creates a buffer large enough to store the output of all calls
// to Split.
func SplitBuffer(args *Args, buf Particles) Particles {
	panic("NYI")
}

// Split splits off the particles associated with file i from a full Particles
// array and writes it to a given buffer.
func (part Particles) Split(args *Args, i int, buf Particles) {
	panic("NYI")
}
