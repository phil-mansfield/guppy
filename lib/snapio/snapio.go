/*package snapio contains functions for reading snapshot files. Adding support
for a new file format requires writing a function to read those snapshot files,
and writing a struct that implements the Header interface.
*/
package snapio

// Header is an abstraction over the the header data of various snapshot
// formats.
type Header interface {
	// ToBytes converts the Header to bytes. In most cases, this should just be
	// calling binary.Encode() on the struct.
	ToBytes() []byte
}
