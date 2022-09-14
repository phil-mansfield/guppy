package snapio


import (
	"encoding/binary"
	"fmt"
	"os"
	"bytes"
	"sort"
	//"math"
)

const (
	gadget2HeaderSize = 256
)

// abstractGadget2 is a base type which implements reading from files.
// LGadget2 and vanilla cosmological Gadget2 simulations  have identical data
// layouts, but different header formats, so the two types embed the abstract
// type and implement headers themselves.
type abstractGadget2 struct {
	fileName string
	names, types []string
	order binary.ByteOrder
	hd *Gadget2Header
}

func (f *abstractGadget2) Read(name string, buf *Buffer) error {
	// Do error handling on the existence of the file. This should already have
	// happened, so if this is triggering it's technically a low-priority
	// internal error.
	file, err := os.Open(f.fileName)
	if err != nil {
		return fmt.Errorf("The file %s does not exist or cannot be " + 
			"accessed.", f.fileName)
	}
	defer file.Close()

	// Find the block's offset.
	offset := int64(8 + gadget2HeaderSize)

	var i int
	for i = 0; i < len(f.names); i++ {
		if f.names[i] == name { break }
		offset += blockSize(f.types[i], f.hd.n) + 8
	}

	finalBlockSize := blockSize(f.types[i], f.hd.n)

	// Check that the Fortran block header is right.
	_, err = file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("Internal error: %s. It's likely that the " + 
			"provided Gadget-2 block types are incorrect, although Guppy " + 
			"did not detect this somehow.", err.Error())
	}

	hdSize := uint32(0)
	err = binary.Read(file, f.order, &hdSize)
	if err != nil {
		return fmt.Errorf("Internal error: %s. It's likely that the " + 
			"provided Gadget-2 block types are incorrect, although Guppy " + 
			"did not detect this somehow.", err.Error())
	}

	if hdSize % uint32(f.hd.n) != 0 {
		return fmt.Errorf("The header uint32 of the the '%s' block in the " + 
			"file %s is garbage: %d when it should be %d. This likely means " + 
			"that at least one of the earlier blocks shouldn't be there, " +
			"has the wrong type, or is missing. The supplied blocks are " + 
			"%s, with types %s.",
			f.names[i], f.fileName, finalBlockSize, hdSize, f.names, f.types)
	} else if hdSize != uint32(finalBlockSize) {
		frac := float64(hdSize) / float64(f.hd.n)
		return fmt.Errorf("The block '%s' in file should have %d bytes due " + 
			"to its type, '%s', but actually has %d bytes. This is likely " + 
			"due to using the incorrect type for this block. Note that the " + 
			"two sizes are off by a factor of %g.", f.names[i],
			f.hd.n, f.types[i], hdSize, frac,
		)
	}

	// After all that error detection, reading is very easy. (Isn't I/O fun?)
	err = buf.read(file, f.names[i], f.hd.n)

	/*
	if err == nil && name == "v" {
		vIntr, err := buf.Get("v")
		if err != nil { panic(fmt.Sprintf("Internal error: %s", err.Error())) }
		v, ok := vIntr.([][3]float32)
		if !ok { panic("Internal type consistency error.") }

		rootA := float32(math.Sqrt(float64(1/(1 +f.hd.Z()))))
		for i := range v {
			for dim := 0; dim < 3; dim++ {
				v[i][dim] *= rootA
			}
		}	
	}
*/

	return err
}

func blockSize(typ string, n int) int64 {
	wordSize := -1
	switch typ {
	case "u32", "f32": wordSize = 4
	case "u64", "f64": wordSize = 8
	case "v32": wordSize = 12
	case "v64": wordSize = 24
	}
	if wordSize == -1 {
		panic(fmt.Sprintf("Internal error: unrecognized type string, '%s'",
			typ))
	}
	return int64(wordSize)*int64(n)
}

// LGadget2 is an implementation of the File interface for LGadget-2 files.
// These files have different header fields than a standard Gadget-2 file and
// always have uniform particle masses, but store data identically. See the
// File interface for a description of the methods.
type LGadget2 struct {
	abstractGadget2
}

// NewLGadget2 creates a new LGadget2 file with the given file name, byte order,
// field names, and types. The field names are only used internally to keep
// track fo variables and the the varaible names follow the common Guppy
// convention: "u32" and "u64" are ints, "f32" and "f64" are floats, and "v32"
// and "v64" are 3-vectors.
//
// To aid with error-catching, Guppy recognizes serveral common varaibles names
// and will crash if incorrect types are assigned to them:
// x - v32
// v - v32
// id - u32 or u64
// phi - f32
// acc - v32
// dt - f32
func NewLGadget2(
	fileName string, names, types []string, order binary.ByteOrder,
) (*LGadget2, error) {
	err := checkGadget2File(fileName)
	if err != nil { return nil, err }
	err = checkGadget2Types(names, types)
	if err != nil { return nil, err }

	f := &LGadget2{ abstractGadget2{ fileName, names, types, order, nil } }
	f.hd, err = f.readHeader()
	if err != nil { return nil, err }
	err = checkGadget2FileSize(fileName, f.hd.n, types)
	if err != nil { return nil, err }

	return f, nil
}

// NewLGadget2Cosmological creates a new cosmological Gadget2 file with the
// given file name, byte order, field names, and types. The field names are only
// used internally to keep track fo variables and the the varaible names follow
// the common Guppy convention: "u32" and "u64" are ints, "f32" and "f64" are
// floats, and "v32" and "v64" are 3-vectors.
//
// To aid with error-catching, Guppy recognizes serveral common varaibles names
// and will crash if incorrect types are assigned to them:
// x - v32
// v - v32
// id - u32 or u64
// phi - f32
// acc - v32
// dt - f32
func NewGadget2Cosmological(
	fileName string, names, types []string, order binary.ByteOrder,
) (*Gadget2Cosmological, error) {
	err := checkGadget2File(fileName)
	if err != nil { return nil, err }
	err = checkGadget2Types(names, types)
	if err != nil { return nil, err }

	f := &Gadget2Cosmological{
		abstractGadget2{ fileName, names, types, order, nil } }
	f.hd, err = f.readHeader()
	if err != nil { return nil, err }
	err = checkGadget2FileSize(fileName, f.hd.n, types)
	if err != nil { return nil, err }

	return f, nil
}

// checkGadget2File returns an error if the given file can't be opened or if
// it i s adirectory or too small to be a Gadget-2 file.
func checkGadget2File(fileName string) error {
	info, err := os.Stat(fileName)
	if err != nil {
		return fmt.Errorf("The file %s cannot be opened. The system error " + 
			"is: \"%s\"", fileName, err.Error())
	} else if info.IsDir() {
		return fmt.Errorf("The file %s is a directory, not a Gadget-2 file.",
			fileName)
	} 

	return nil
}

func checkGadget2FileSize(fileName string, n int, types []string) error {
	info, err := os.Stat(fileName)
	if err != nil {
		return fmt.Errorf("The file %s cannot be opened. The system error " + 
			"is: %s", fileName, err.Error())
	}

	size := int64(8 + gadget2HeaderSize)
	for i := range types {
		size += 8 + blockSize(types[i], n)
	}

	//maxSizeDiff := int64(n*4)
	//if size + maxSizeDiff < info.Size() || size - maxSizeDiff > info.Size() {
	if size> info.Size() {
		return fmt.Errorf("The provided Gadget-2 data types, %s, would lead " + 
			"to the %d-particle file, %s, having %d bytes, but it actually " + 
			"has %d bytes. You should check that the types are correct " + 
			"(particularly the id size) and that no fields are missing or " + 
			"incorrect. Gadget will often generate files with some junk " + 
			"data in them, but the size difference in this case is way " + 
			"too big.", 
			types, n, fileName, size, info.Size(),
		)
	}

	return nil
}

// checkGadget2Types returns nil if names and types secribe a valid set of 
// Gadget-2 names and types, respectively. Otherwise an error is returned.
func checkGadget2Types(names, types []string) error {
	if len(names) != len(types) {
		return fmt.Errorf("%d block names were given for Gadget-2 " + 
			"files, but %d block types were given.", len(names), len(types))
	} 

	if s, ok := containsDuplicates(names); ok {
		return fmt.Errorf("'%s' occurs multiple times in the " +
			"list of block names given for Gadget-2 files, %s.", s, names)
	}

	hasID := false

	for i := range types {
		// Check that the types are set to something valid.
		switch types[i] {
		case "u32", "u64", "f32", "f64", "v32", "v64":
		default:
			return fmt.Errorf("block %d in Gadget-2 files, '%s', was given " + 
				"type '%s', but the only valid types are 'u32', 'u64', " + 
				"'f32', 'f64', 'v32', 'v64'", i, names[i], types[i])
		}

		// Check that known blocks have valid types.
		n, t := names[i], types[i]
		if err := knownBlock(n, t, "x", []string{"v32"}); err != nil {
			return err
		} else if err := knownBlock(n, t, "v", []string{"v32"}); err != nil {
			return err
		} else if err := knownBlock(n, t, "id", []string{"u32", "u64"});
			err != nil {
			return err
		} else if err := knownBlock(n, t, "phi", []string{"f32"}); err != nil {
			return err
		} else if err := knownBlock(n, t, "acc", []string{"v32"}); err != nil {
			return err
		} else if err := knownBlock(n, t, "dt", []string{"f32"}); err != nil {
			return err
		}

		// Check that an id field was supplied.
		if names[i] == "id" { hasID = true }
	}

	if !hasID {
		return fmt.Errorf("Guppy requires an 'id' block, but no such block " + 
			"was specified.")
	}
	return nil
}

// knownBlock checks the name and type of a field. If name == targetName and
// typ is not in validTypes, an error is returned. Otherwise, nil is returned.
func knownBlock(name, typ, targetName string, validTypes []string) error {
	if name != targetName { return nil }
	for i := range validTypes {
		if typ == validTypes[i] { return nil }
	}
	if len(validTypes) == 1 {
			return fmt.Errorf("The block '%s' was given the type '%s', " + 
				"but '%s' blocks in Gadget-2 must have type '%s'",
				name, typ, name, validTypes[0])
	} 
	return fmt.Errorf("The block '%s' was given the type '%s', but '%s' " +
		"blocks in Gadget-2 must have types that are one of: %s",
		name, typ, name, validTypes)
}

// containsDuplicates tests whether any strings show up multiple times.
// If so, it returns on of those strings and returns true, otherwise it returns
// and empty stgring and false.
func containsDuplicates(s []string) (string, bool) {
	sSort := make([]string, len(s))
	for i := range s { sSort[i] = s[i] }
	sort.Strings(sSort)
	for i := 1; i < len(sSort); i++ {
		if sSort[i] == sSort[i - 1] {
			return sSort[i], true
		}
	}
	return "", false
}

// Gadget2Cosmological is an implementation of the File interface for standard
// Gadget-2 files. Gadget-2 files do not have a standard set of variables
// or order to those variables, so they must be specified at runtime. This
// file assumes that particle masses are uniform. See the File interface for a
// description of the methods.
type Gadget2Cosmological struct {
	abstractGadget2
}

// Gadget2Header implements the Header interface for either an LGadget-2 or
// Gadget-2 simulation. See the Header interface for a description of the
// methods.
type Gadget2Header struct {
	rawBytes []byte
	order binary.ByteOrder
	names, types []string
	n int
	nTot int64
	z, omegaM, omegaL, h100, l, mass float64
}

func (hd *Gadget2Header) ToBytes() []byte { return hd.rawBytes }
func (hd *Gadget2Header) ByteOrder() binary.ByteOrder { return hd.order }
func (hd *Gadget2Header) Names() []string { return hd.names }
func (hd *Gadget2Header) Types() []string { return hd.types }
func (hd *Gadget2Header) NTot() int64 { return hd.nTot }
func (hd *Gadget2Header) Z() float64 { return hd.z }
func (hd *Gadget2Header) OmegaM() float64 { return hd.omegaM }
func (hd *Gadget2Header) OmegaL() float64 { return hd.omegaL }
func (hd *Gadget2Header) H100() float64 { return hd.h100 }
func (hd *Gadget2Header) L() float64 { return hd.l }
func (hd *Gadget2Header) Mass() float64 { return hd.mass }

// rawLGadget2Header is a struct with the same fields as the raw header data of 
// an LGadget-2 file.
type rawLGadget2Header struct {
	NPart [6]uint32
	Mass [6]float64
	Time, Redshift float64
	FlagSFR, FlagFeedback uint32
	NPartTotal [6]uint32
	RawFlagCooling, NumFiles uint32
	BoxSize, Omega0, OmegaLambda, HubbleParam float64
	FlagStellarAge, HashTabSize uint32
	Empty [88]byte
}

// rawGadget2Header is a struct with the same fields as the raw header data of 
// an LGadget-2 file.
type rawGadget2Header struct {
	NPart [6]uint32
	Mass [6]float64
	Time, Redshift float64
	FlagSFR, FlagFeedback uint32
	Nall [6]uint32
	FlagCooling, NumFiles uint32
	BoxSize, Omega0, OmegaLambda, HubbleParam float64
	FlagStellarAge, FlagMetals uint32
	NallHW [6]uint32
	FlagEntroypICs uint32
	Empty [60]byte
}

// readRawGadgetHeader is a generic Gadget-2 header-reading function that
// handles all the messy error-handling and file I/O. It take the name of the
// file, the byte order and a pointer to the header struct being read as
// arguments.
func readRawGadgetHeader(
	fileName string, order binary.ByteOrder, rawHd interface{},
) error {
	file, err := os.Open(fileName)
	if err != nil { return err }
	defer file.Close()

	nHeader, nFooter := uint32(0), uint32(0)

	// Read the header block and check that it's the right size.
	err = binary.Read(file, order, &nHeader)
	if err != nil { return err }
	if nHeader != gadget2HeaderSize {
		return fmt.Errorf("%s is not a valid Gadget-2 file: the first " +
			"integer would lead to a header with %d bytes instead of %d.",
			fileName, nHeader, gadget2HeaderSize)
	}

	// Read the data block.
	err = binary.Read(file, order, rawHd)
	if err != nil { return err }

	// Confirm that everything's okay with the footer.
	err = binary.Read(file, order, &nFooter)
	if err != nil { return err }
	if nHeader != nFooter {
		return fmt.Errorf("%s is not a valid Gadget-2 file: the header, %d, " + 
			"and footer, %d, of the first data block don't match.",
			fileName, nHeader, nFooter,
		)
	}

	return nil
}

func (f *LGadget2) readHeader() (hd *Gadget2Header, err error) {
	rawHd := &rawLGadget2Header{ }
	err = readRawGadgetHeader(f.fileName, f.order, rawHd)
	if err != nil { return nil, err }

	n := int(rawHd.NPart[1])
	nTot := int64(uint64(rawHd.NPartTotal[1]) +
		uint64(rawHd.NPartTotal[0])<<32)

	buf := &bytes.Buffer{ }
	err = binary.Write(buf, f.order, rawHd)
	if err != nil { return nil, err }

	hd = &Gadget2Header{
		rawBytes: buf.Bytes(), order: f.order,
		names: f.names, types: f.types,
		n: n, nTot: nTot,
		z: rawHd.Redshift, omegaM: rawHd.Omega0,
		omegaL: rawHd.OmegaLambda,
		h100: rawHd.HubbleParam, l: rawHd.BoxSize,
		mass: rawHd.Mass[1] * 1e10, 
	}

	return hd, nil
}

func (f *Gadget2Cosmological) readHeader() (hd *Gadget2Header, err error) {
	rawHd := &rawGadget2Header{ }
	err = readRawGadgetHeader(f.fileName, f.order, rawHd)
	if err != nil { return nil, err }

	n := int(rawHd.NPart[1])
	nTot := int64(uint64(rawHd.Nall[1]) + uint64(rawHd.NallHW[0])<<32)

	buf := &bytes.Buffer{ }
	err = binary.Write(buf, f.order, rawHd)
	if err != nil { return nil, err }

	hd = &Gadget2Header{
		rawBytes: buf.Bytes(), order: f.order,
		names: f.names, types: f.types,
		n: n, nTot: nTot,
		z: rawHd.Redshift, omegaM: rawHd.Omega0,
		omegaL: rawHd.OmegaLambda,
		h100: rawHd.HubbleParam, l: rawHd.BoxSize,
		mass: rawHd.Mass[1] * 1e10, 
	}

	return hd, nil
}

func (f *Gadget2Cosmological) ReadHeader() (Header, error) { return f.hd, nil }
func (f *LGadget2) ReadHeader() (Header, error) { return f.hd, nil }

// Type checking
var (
	_ Header = &Gadget2Header{ }
	_ File = &LGadget2{ }
	_ File = &Gadget2Cosmological{ } 
)
