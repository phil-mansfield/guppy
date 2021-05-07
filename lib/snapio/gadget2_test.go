package snapio

import (
	"encoding/binary"
	"testing"
	"unsafe"
	"fmt"
)

var (
	fileName = "../../large_test_data/snapshot_023.39"
	names = []string{"x", "v", "id"}
	types = []string{"v32", "v32", "u64"}
	order = binary.LittleEndian
)


func TestGadget2HeaderSize(t *testing.T) {
	if size := unsafe.Sizeof(rawGadget2Header{ }); size != 256 {
		t.Errorf("rawGadget2Header{} has size %d, not 256", size)
	}
	if size := unsafe.Sizeof(rawLGadget2Header{ }); size != 256 {
		t.Errorf("rawLGadget2Header{} has size %d, not 256", size)
	}
}
func TestGadget2CosmologicalFailure(t *testing.T) {
	fileNames := []string{"file_that_doesn't_exist.dat", "test_files",
		"test_files/tiny_file.txt"}

	for _, fileName := range fileNames {
		_, err := NewGadget2Cosmological(fileName, names, types, order)
		if err == nil { 
			t.Errorf("Expected read of %s to fail, but succeeded.", fileName)
		}
	}

	tests := []struct{
		names, types []string
	} {
		{ []string{}, []string{} },
		{ []string{"x"}, []string{"x32"} },
		{ []string{"x"}, []string{"f32"} },
		{ []string{"v"}, []string{"f32"} },
		{ []string{"id"}, []string{"v32"} },
		{ []string{"x", "x", "id"}, []string{"v32", "v32", "u64"} },
		{ []string{"x", "v", "id"}, []string{"v32", "v32", "u32"} },
		{ []string{"x", "v", "id", "phi"},
			[]string{"v32", "v32", "u32", "f32"} },
	}

	for i := range tests {
		_, err := NewGadget2Cosmological(
			fileName, tests[i].names, tests[i].types, order,
		)
		if err == nil {
			t.Errorf("Expected read with names = %s and types = %s to fail, " + 
				"but succeeded", tests[i].names, tests[i].types)
		} else {
			fmt.Println(err.Error())
		}
	}
}

//func TestReadGadget2Header(t *testing.T) {
//	f, err := NewGadget2(TestFile, names, types, order)
//	if err != nil {
//		t.Fatalf("Expected valid read, got error message %s.")
//	}
//}