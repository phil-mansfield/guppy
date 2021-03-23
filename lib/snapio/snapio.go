/*package snapio contains functions for reading snapshot files. It centers
around three main types. The first is a Buffer type which files read data into,
the second is a File interface that abstracts over the details of different
files, and the last is a Header interface that abstracts around header
information.
*/
package snapio

import (
	"encoding/binary"
)

// File is a generic interface around different file types.
type File interface {
	// ReadHeader reads the file's header and abstracts it behind the Header
	// interface.
	ReadHeader() (Header, error)
	// Read reads a given variable into a Buffer.
	Read(name string, buf *Buffer) error
}

// Header is a generic interface around the headers of different file tpyes.
type Header interface {
	// ToBytes converts the content of the original header to bytes. This should
	// be preserved exactly, so that there is no header information lost.
	ToBytes() []byte
	// ByteOrder returns the order of bytes in the file.
	ByteOrder() binary.ByteOrder

	// Names returns the names of the fields stored in the file, in the order
	// they will be stored in the .gup file.
	Names() []string
	// Types returns strings describing the types of the file's fields.
	Types() []string
	
	// NTot returns the total number of particles in the simulation.
	NTot() int
	// Z returns the redshift of the snapshot.
	Z() float64
	// OmegaM returns Omega_m(z=0).
	OmegaM() float64
	// H100 returns H0 / (100 km/s/Mpc).
	H100() float64
	// L returns the width of the simulation box in comoving Mpc/h.
	L() float64
}
