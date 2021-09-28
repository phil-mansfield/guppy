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
		[][3]float32{{5.0, 5.0, 5.0}}, [][3]float64{{6.0, 6.0, 6.0}},
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
		types[i] = names[i][len(names[i])-3:]
	}
	
	g, err := NewEqualSplitUnigrid(hd, order, 2, names)
	if err != nil {
		t.Errorf("Initialization of EqualSplitUnigrid with names = %s failed with error '%s', but it should have succeeded", names, err.Error())
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

func TestEqualSplitUnigridBuffers(t *testing.T) {
	names := []string{
		"x_f32", "x_f64", "x_u32", "x_u64", "x_v32", "x_v64", "id",
	}
	values :=[]interface{}{
		[]float32{1.0}, []float64{2.0}, []uint32{3}, []uint64{4},
		[][3]float32{{5.0, 5.0, 5.0}}, [][3]float64{{6.0, 6.0, 6.0}},
		[]uint32{7},
	}
	binOrder := binary.LittleEndian
	order := NewZMajorUnigrid(10)

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

	g, err := NewEqualSplitUnigrid(hd, order, 2, names)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	p := g.Buffers()
	if len(p) != 8 {
		t.Errorf("Expected %d Particles{ } arrays, got %d", 8, len(p))
	}

	pNames := []string{
		"x_f32", "x_f64", "x_u32", "x_u64",
		"x_v32{0}", "x_v32{1}", "x_v32{2}",
		"x_v64{0}", "x_v64{1}", "x_v64{2}",
	}
	pTypes := []string{
		"f32", "f64", "u32", "u64",
		"f32", "f32", "f32",
		"f64", "f64", "f64",
	}

	
	for i := range p {		
		for j := range pTypes {
			name := pNames[j]
			field, ok := p[i][name]
			if !ok {
				t.Errorf("field '%s' missing from p{%d}", name, i)
				continue
			}
			
			if field.Len() != 125 {
				t.Errorf("field '%s' has length %d, but expected %d.",
					name, field.Len(), 125)
			}

			switch field.(type) {
			case *Float32:
				if pTypes[j] != "f32" {
					t.Errorf("Field '%s' has type '%s', but field is Float32",
						name, pTypes[j])
				}
			case *Float64:
				if pTypes[j] != "f64" {
					t.Errorf("Field '%s' has type '%s', but field is Float64",
						name, pTypes[j])
				}
			case *Uint32:
				if pTypes[j] != "u32" {
					t.Errorf("Field '%s' has type '%s', but field is Uint32",
						name, pTypes[j])
				}
			case *Uint64:
				if pTypes[j] != "u64" {
					t.Errorf("Field '%s' has type '%s', but field is Uint64",
						name, pTypes[j])
				}
			case *Vec32:
				if pTypes[j] != "v32" {
					t.Errorf("Field '%s' has type '%s', but field is Vec32",
						name, pTypes[j])
				}
			case *Vec64:
				if pTypes[j] != "v64" {
					t.Errorf("Field '%s' has type '%s', but field is Float32",
						name, pTypes[j])
				}
			default:
				t.Errorf("Field '%s' has type '%s', but field is unknown.",
						name, pTypes[j])
			}
		}
	}
}

func TestEqualSplitUnigrid(t *testing.T) {
	names := []string{
		"x_f32", "x_f64", "x_u32", "x_u64", "x_v32", "x_v64", "id",
	}
	values :=[]interface{}{
		[]float32{1.0}, []float64{2.0}, []uint32{3}, []uint64{4},
		[][3]float32{{5.0, 5.0, 5.0}}, [][3]float64{{6.0, 6.0, 6.0}},
		[]uint32{7},
	}
	binOrder := binary.LittleEndian
	order := NewZMajorUnigrid(10)

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

	g, err := NewEqualSplitUnigrid(hd, order, 2, names)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	tests := []struct{
		id []uint64
		from, to [][]int
		valid bool
	} {
		{
			[]uint64{ },
			[][]int{ {}, {}, {}, {}, {}, {}, {}, {} },
			[][]int{ {}, {}, {}, {}, {}, {}, {}, {} },
			true,
		},
		{
			[]uint64{ 0 },
			[][]int{ {0}, {}, {}, {}, {}, {}, {}, {} },
			[][]int{ {0}, {}, {}, {}, {}, {}, {}, {} },
			true,
		},
		{
			[]uint64{ 1 },
			[][]int{ {0}, {}, {}, {}, {}, {}, {}, {} },
			[][]int{ {25}, {}, {}, {}, {}, {}, {}, {} },
			true,
		},
		{
			[]uint64{ 100 },
			[][]int{ {0}, {}, {}, {}, {}, {}, {}, {} },
			[][]int{ {1}, {}, {}, {}, {}, {}, {}, {} },
			true,
		},
		{
			[]uint64{ 500 },
			[][]int{ {}, {0}, {}, {}, {}, {}, {}, {} },
			[][]int{ {}, {0}, {}, {}, {}, {}, {}, {} },
			true,
		},
		{
			[]uint64{600, 500, 501, 510},
			[][]int{ {}, {0, 1, 2, 3}, {}, {}, {}, {}, {}, {} },
			[][]int{ {}, {1, 0, 25, 5}, {}, {}, {}, {}, {}, {} },
			true,
		},
		{
			[]uint64{600, 500, 501, 510},
			[][]int{ {}, {0, 1, 2, 3}, {}, {}, {}, {}, {}, {} },
			[][]int{ {}, {1, 0, 25, 5}, {}, {}, {}, {}, {}, {} },
			true,
		},
		{
			[]uint64{0, 5},
			[][]int{ {0}, {}, {}, {}, {1}, {}, {}, {} },
			[][]int{ {0}, {}, {}, {}, {0}, {}, {}, {} },
			true,
		},
		{
			[]uint64{0, 5, 50, 500, 555, 556, 565, 655},
			[][]int{ {0}, {3}, {2}, {}, {1}, {}, {}, {4, 5, 6, 7} },
			[][]int{ {0}, {0}, {0}, {}, {0}, {}, {}, {0, 25, 5, 1} },
			true,
		},
	}

	for i := range tests {
		var err error
		from, to := startingIndexArray(8)
		from, to, err = g.Indices(tests[i].id, from, to)
		
		if tests[i].valid && err != nil {
			t.Errorf("%d) Expected valid Indices() call, but got error '%s'.",
				i, err.Error())
			return
		} else if !tests[i].valid && err == nil {
			t.Errorf("%d) Expected invalid Indices() call, but got no error.",
				i)
			return
		}
		
		if len(from) != 8 || len(to) != 8 {
			t.Errorf(
				"Length of from and to were %d and %d, but expected 8 and 8.",
				len(from), len(to),
			)
		}

		for j := range from {
			if !eq.Ints(from[j], tests[i].from[j]) {
				t.Errorf("%d) Expected from[%d] = %d, got %d.",
					i, j, tests[i].from[j], from[j])
			}
			if !eq.Ints(to[j], tests[i].to[j]) {
				t.Errorf("%d) Expected to[%d] = %d, got %d.",
					i, j, tests[i].to[j], to[j])
			}
		}
	}
}

// Create from and to arrays with different lengths and random junk in them.
func startingIndexArray(n int) (from, to [][]int) {
	from, to = make([][]int, n), make([][]int, n)
	k := 0
	for i := range from {
		for j := 0; j < i; j++ {
			from[i] = append(from[i], k)
			to[i] = append(to[i], k)
			k++
		}
	}

	return from, to
}

func TestSplit(t *testing.T) {
	data := []struct{
		id []uint32
		x []float32
	} {
		{
			[]uint32{ 4, 7, 2 },
			[]float32{ 1, 7, 2 },	
		},
		{
			[]uint32{ 0 },
			[]float32{ 0 },	
		},
		{
			[]uint32{ },
			[]float32{ },	
		},
		{
			[]uint32{ 6, 5, 1, 3 },
			[]float32{ 3, 5, 4, 6 },	
		},
	}

	names := []string{"x", "id"}
	nTot := 8
	order := binary.LittleEndian

	files := make([]snapio.File, len(data))
	for i := range files {
		var err error
		files[i], err = snapio.NewFakeFile(
			names, []interface{}{ data[i].x, data[i].id }, nTot, order,
		)
		if err != nil {
			t.Errorf("Error while creating FakeFile %d, '%s'", i, err.Error())
			return 
		}
	}

	hd, err := files[0].ReadHeader()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	idOrder := NewZMajorUnigrid(2)
	
	scheme, err := NewEqualSplitUnigrid(hd, idOrder, 2, []string{"x"})
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	p, err, errType := Split(scheme, files, []string{"x"})
	if err != nil {
		t.Errorf("%v: '%s'", errType, err.Error())
		return
	}

	for i := range p {
		field := p[i]["x"]
		x, ok := field.Data().([]float32)
		if !ok {
			t.Errorf("p[%d]['x'] is not a float32 array.", i)
			continue
		}
		
		if len(x) != 1 || !eq.Float32s(x, []float32{ float32(i) }) {
			t.Errorf("p[%d]['x'] = %.1f.", i, x)
		}
	}
}
