package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	
	"github.com/phil-mansfield/guppy/lib"
)

// Note: there is almost no reason why you would want to handle reading in
// this way rather than calling the library functions in go/guppy. This mode
// of reading guppy files exists almost solely for the benefit of C programs
// that rely on fork() calls and therefore nuke the Go runtime. I'm supplying
// this here so future developers know how this works.

func main() {
	// Build a command to guppy that gets you the data you want.
	cmd := exec.Command("../guppy", "read",
		"-file", "../large_test_data/L125_sheet000_snap_100.gadget2.dat.gup",
		"-vars", "{RockstarParticle},x{0},v,id")

	
	// Create a pipe the guppy command. This is like calling popen.
	pipe, err := cmd.StdoutPipe()
	if err != nil { panic(err.Error()) }

	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil { panic(err.Error()) }
	
	// Guppy writes in system order so our C-based friends don't have to
	// do more work.
	sysOrder := lib.SystemByteOrder()
	
	// Read the header first
	hd := &lib.PipeHeader{ }

	err = binary.Read(pipe, sysOrder, hd)
	if err != nil { panic(err.Error()) }
	
	part := make([]lib.RockstarParticle, hd.N)
	x0 := make([]float32, hd.N)
	v := make([][3]float32, hd.N)
	id := make([]uint64, hd.N)
	
	// Read the data next. You could do this with calls to binary.Read(), but
	// Go uses reflection to read composite type and arrays of arrays. So you
	// should really use the function I wrote that takes care of this for you.
	err = lib.ReadAsBytes(pipe, part)
	if err != nil { panic(err.Error) }
	err = lib.ReadAsBytes(pipe, x0)
	if err != nil { panic(err.Error) }
	err = lib.ReadAsBytes(pipe, v)
	if err != nil { panic(err.Error) }
	err = lib.ReadAsBytes(pipe, id)
	if err != nil { panic(err.Error) }

	for i := 0; i < 8; i++ {
		fmt.Printf("%9x %.4f %.4f\n", part[i].ID, part[i].X, part[i].V)
	}
}

