/*package read_guppy provides several functions for reading .gup files.*/
package read_guppy

// Header contains information about the simulation and compression process.
type Header struct {
}

// File represents a .gup file. Data can be read from this file.
type File struct {
	Header
}

// Open creates a File struct corresponding to the file located at fname.
func Open(fname string) *File {
	panic("NYI")
}
