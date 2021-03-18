/*package snapio contains functions for reading snapshot files. Adding support
for a new file format requires writing a function to read those snapshot files,
and writing a struct that implements the Header interface.
*/
package snapio

type FileType int64
const (
	Gadget2 FileType = iota
	LGadget2
)

// File is a generic interface around
type File interface {
	FileType() FileType
	ReadHeader() Header
	Read(name string, buf Buffer)
}

type Header interface {
	ToBytes() []byte
}
