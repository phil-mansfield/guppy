package particles

import (
	"encoding/binary"
	"testing"

	"github.com/phil-mansfield/guppy/lib/eq"
	"github.com/phil-mansfield/guppy/lib/snapio"
)

func TestRound(t *testing.T) {
	tests := []struct{
		x float64
		i int
	} {
		{0.0, 0}, {1.0, 1}, {1.1, 1}, {1.9, 2}, {-1.1, -1}, {-1.9, -2},
		{-1.5, -1}, {1.5, 2},
	}

	for j := range tests {
		i := round(tests[j].x)
		if i != tests[j].i {
			t.Errorf("%d) Expected round(%.2f) = %d, got %d.",
				j, tests[j].x, tests[j].i, i)
		}
	}
}

func TestNewEqualSplitUnigrid(t *testing.T) {
	names := []string{
		"x_f32", "x_f64", "x_u32", "x_u64", "x_v32", "x_v64", "id",
	}
	values :=[]interface{}{
		[]float32{1.0}, []float64{2.0}, []uint32{3}, []uint64{4},
		[][3]float32{{5.0, 5.0, 5.0}}, [][3]float32{{6.0, 6.0, 6.0}},
		[]uint32{7},
	}
	binOrder := binary.LittleEndian
	order := NewZMajorUnigrid(10)
	
	
	fBadNTot, err := snapio.NewFakeFile(names, values, 50, binOrder)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	hdBadNTot, err := fBadNTot.ReadHeader()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	fMismatchNTot, err := snapio.NewFakeFile(names, values, 64, binOrder)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	hdMismatchNTot, err := fMismatchNTot.ReadHeader()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	f, err := snapio.NewFakeFile(names, values, 1000, binOrder)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	hd, err := f.ReadHeader()
	if err != nil {
		t.Errorf(err.Error())
		return
	}


	
	_, err = NewEqualSplitUnigrid(hdBadNTot, order, 2, names)
	if err == nil {
		t.Errorf("Expected Header with %d particles to be invalid Unigrd SplitSheme, but got no error.", hdBadNTot.NTot())
	}
	_, err = NewEqualSplitUnigrid(hdMismatchNTot, order, 2, names)
	if err == nil {
		t.Errorf("Expected Header with %d particles total and IDOrder with %d particles would lead to error, but got none.", 64, 1000)
	}
	_, err = NewEqualSplitUnigrid(hd, order, 3, names)
	if err == nil {
		t.Errorf("Expected EqualSplitUnigrid split with nAll = 10 and nCube = 3 to fail, but got no error.")
	}
	_, err = NewEqualSplitUnigrid(hd, order, 2, []string{"meow"})
	if err == nil {
		t.Errorf("Expected Unigrid with invalid variable name to fail, but got no error.")
	}

	types := make([]string, len(names) - 1)
	for i := range types {
		types[i] = names[i][len(types)-3:]
	}
	
	g, err := NewEqualSplitUnigrid(hd, order, 2, names)
	if err != nil {
		t.Errorf("Initialization of EqualSplitUnigrid with names = %s failed",
			names)
	} else if g.order != order {
		t.Errorf("Unigrid's IDOrder changed form input order.")
	} else if g.nAll != 10 {
		t.Errorf("nAll = %d instead of %d.", g.nAll, 10)
	} else if g.nSub != 5 {
		t.Errorf("nSub = %d instead of %d.", g.nSub, 2)
	} else if g.nCube != 2 {
		t.Errorf("nCube = %d instead of %d", g.nCube, 5)
	} else if !eq.Strings(g.names, names[:len(names) -1]) {
		t.Errorf("Expected names = %s, got %s.", names[:len(names)-1], g.names)
	} else if !eq.Strings(g.types, types) {
		t.Errorf("Expected types = %s, got %s.", types, g.types)
	}
}
