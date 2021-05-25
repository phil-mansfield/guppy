package compress

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/phil-mansfield/guppy/lib/eq"
	"github.com/phil-mansfield/guppy/lib/particles"
)

func TestBuffer(t *testing.T) {
	tests := []int{ 0, 10, 0, 10, 20, 30, 30, 30, 10, 10000, 0}
	buf := NewBuffer(0)
	prevLen := 0
	// Different element sizes grow at different rates when appending.
	prevCapMap := map[string]int {
		"b": 0, "u32": 0, "u64": 0, "u64_2": 0, "f32": 0, "f64": 0,
	}
	
	for i := range tests {		
		buf.Resize(tests[i])

		test := func(name string, l, c int) {
			prevCap := prevCapMap[name]
			if l != tests[i] {
				t.Errorf("%d) buf.%s resized to %d, but had len %d.",
					i, name, tests[i], l)
			} else if l <= prevLen && c > prevCap {
				t.Errorf("%d) buf.%s didn't need to change cap size when len went form %d to %d , but increased from %d to %d",
					i, name, prevLen, l, prevCap, c)
			}
			prevCapMap[name] = c
		}

		test("b", len(buf.b), cap(buf.b))
		test("u32", len(buf.u32), cap(buf.u32))
		test("u64", len(buf.u64), cap(buf.u64))
		test("q", len(buf.q), cap(buf.q))
		test("f32", len(buf.f32), cap(buf.f32))
		test("f64", len(buf.f64), cap(buf.f64))

		prevLen = len(buf.u32)
	}
}

func TestQuantize(t *testing.T) {
	name := "meow"
	
	tests := []struct{
		f particles.Field
		delta float64
	} {
		{particles.NewUint32(name, []uint32{ }), 0.0},
		{particles.NewUint64(name, []uint64{ }), 0.0},
		{particles.NewFloat32(name, []float32{ }), 0.0},
		{particles.NewFloat64(name, []float64{ }), 0.0},

		{particles.NewUint32(name, []uint32{0, 1, 2, 3, 4, 5}), 0.0},
		{particles.NewUint64(name, []uint64{0, 0, 0, 0,0, 100000,100000}), 0.0},
		{particles.NewFloat32(name,
			[]float32{1, 1.5, 2, 0, 4, 5.5, 6}), 1e-3},
		{particles.NewFloat64(name,
			[]float64{1, 1.5, 2, 0, 4, 5.5, 6}), 1e-3},

	}

	buf := NewBuffer(0)
	
	for i := range tests {
		buf.Resize(tests[i].f.Len())
		Quantize(tests[i].f, tests[i].delta, buf.q)

		var flag TypeFlag
		switch tests[i].f.Data().(type) {
		case []uint32: flag = Uint32Flag
		case []uint64: flag = Uint64Flag
		case []float32: flag = Float32Flag
		case []float64: flag = Float64Flag
		default: panic("'Impossible' type configuration")
		}
		
		f := Dequantize(name, buf.q, tests[i].delta, flag, buf)

		switch x := f.Data().(type) {
		case []uint32:
			y, ok := tests[i].f.Data().([]uint32)
			if !ok {
				t.Errorf("%d) output Field has type []uint32", i)
			} else if !eq.Uint32s(x, y) {
				t.Errorf("%d) Expected output %d, got %d.", i, y, x)
			}
		case []uint64:
			y, ok := tests[i].f.Data().([]uint64)
			if !ok {
				t.Errorf("%d) output Field has type []uint64", i)
			} else if !eq.Uint64s(x, y) {
				t.Errorf("%d) Expected output %d, got %d.", i, y, x)
			}
		case []float32:
			y, ok := tests[i].f.Data().([]float32)
			if !ok {
				t.Errorf("%d) output Field has type []float32", i)
			} else if !eq.Float32sEps(x, y, float32(tests[i].delta)) {
				t.Errorf("%d) Expected output %f, got %f.", i, y, x)
			}
		case []float64:
			y, ok := tests[i].f.Data().([]float64)
			if !ok {
				t.Errorf("%d) output Field has type []float64", i)
			} else if !eq.Float64sEps(x, y, tests[i].delta) {
				t.Errorf("%d) Expected output %f, got %f.", i, y, x)
			}

		default:
			t.Errorf("%d) Unknown type for output, %v", i, f.Data())
		}
	}
}

func TestLagrangianDelta(t *testing.T) {
	order := binary.LittleEndian
	
	lastTestData := make([]float64, 32*16*8)
	for i := range lastTestData {
		lastTestData[i] = rand.Float64()
	}

	tests := []struct{
		span [3]int
		dir int
		delta float64
		data interface{}
	} {
		{ [3]int{2, 2, 2}, 0, 0, []uint32{0, 1, 2, 4, 4, 5, 6, 0} },
		{ [3]int{1, 1, 8}, 0, 0, []uint32{1, 2, 3, 4, 5, 6, 7, 0} },
		{ [3]int{1, 8, 1}, 0, 0, []uint32{0, 1, 2, 4, 4, 5, 6, 0} },
		{ [3]int{8, 1, 1}, 0, 0, []uint32{0, 1, 2, 4, 4, 5, 6, 0} },
		{ [3]int{2, 2, 2}, 0, 0, []uint64{0, 1, 2, 4, 4, 5, 6, 0} },
		{ [3]int{2, 2, 2}, 0, 1e-4, []float32{0, 1, 2, 4, 4, 5, 6, 0} },
		{ [3]int{2, 2, 2}, 0, 1e-4, []float64{0, 1, 2, 4, 4, 5, 6, 0} },
		{ [3]int{32, 16, 8}, 2, 1e-4, lastTestData},
	}

	buf := NewBuffer(0)
	for i := range tests {

		m := NewLagrangianDelta(
			order, tests[i].span, tests[i].dir, tests[i].delta,
		)
		f, err := particles.NewGenericField("meow", tests[i].data)
		if err != nil { t.Errorf(err.Error()) }
		wr := bytes.NewBuffer(make([]byte, 0, 0))

		err = m.WriteInfo(wr)
		if err != nil {
			t.Errorf("%d) Got error '%s' on WriteInfo", i, err.Error())
			continue
		}

		err = m.Compress(f, buf, wr)
		if err != nil {
			t.Errorf("%d) Got error '%s' on Compress", i, err.Error())
			continue
		}

		rd := bytes.NewReader(wr.Bytes())
		mOut := &LagrangianDelta{ }

		err = mOut.ReadInfo(order, rd)
		if err != nil {
			t.Errorf("%d) Got error '%s' on ReadInfo", i, err.Error())
			continue
		}

		fOut, err := mOut.Decompress(buf, rd)
		if err != nil {
			t.Errorf("%d) Got error '%s' on Demcompress", i, err.Error())
			continue
		}

		if mOut.order != order {
			t.Errorf("%d) Expected order = %d, got %d.", i, order, mOut.order)
			continue
		} else if mOut.span != tests[i].span {
			t.Errorf("%d) Expected span = %d, got %d.",
				i, tests[i].span, mOut.span)
			continue
		} else if mOut.dir != tests[i].dir {
			t.Errorf("%d) Expected dir = %d, got %d.",
				i, tests[i].dir, mOut.dir)
			continue
		} else if mOut.delta != tests[i].delta {
			t.Errorf("%d) Expected delta = %g, got %g.",
				i, tests[i].delta, mOut.delta)
			continue
		}

		if fOut.Name() != "meow" {
			t.Errorf("%d) Expected field name '%s', got '%s'.",
				i, "meow", fOut.Name())
			continue
		}
		
		x := fOut.Data()
		dataEqual := false
		switch d := tests[i].data.(type) {
		case []uint32: dataEqual = eq.Generic(d, x)
		case []uint64: dataEqual = eq.Generic(d, x)
		case []float32:
			x32, ok := x.([]float32)
			if !ok {
				t.Errorf("%d) Decompressed array %v is not []float32.", i, x)
			}
			dataEqual = eq.Float32sEps(d, x32, float32(tests[i].delta))
		case []float64:
			x64, ok := x.([]float64)
			if !ok {
				t.Errorf("%d) Decompressed array %v is not []float64.", i, x)
			}
			dataEqual = eq.Float64sEps(d, x64, tests[i].delta)
		}
		if !dataEqual {
			t.Errorf("%d) Compressed the array %v, but it decompressed to %v.",
				i, tests[i].data, x)
		}
	}
}

func TestSplitArray(t *testing.T) {
	tests := []struct{
		x []uint64
		lens []int
		splits [][]uint64
	} {
		{
			[]uint64{},
			[]int{},
			[][]uint64{},
		},
		{
			[]uint64{},
			[]int{},
			[][]uint64{{}},
		},
		{
			[]uint64{1},
			[]int{1},
			[][]uint64{{1}},
		},
		{
			[]uint64{1},
			[]int{0, 1, 0},
			[][]uint64{{}, {1}, {}},
		},
		{
			[]uint64{1, 1, 1, 1, 2, 2, 3, 3, 3},
			[]int{4, 2, 3},
			[][]uint64{{1, 1, 1, 1}, {2, 2}, {3, 3, 3}},
		},
	}

	for i := range tests {
		splits := make([][]uint64, len(tests[i].lens))
		splitArray(tests[i].x, tests[i].lens, splits)

		for j := range splits {
			if !eq.Uint64s(splits[j], tests[i].splits[j]) {
				t.Errorf("%d) Expected spliArray(%d, %d) = %d, got %d.",
					i, tests[i].x, tests[i].lens, tests[i].splits, splits)
				 continue
			}
		}
	}
}

func TestDeltaEncode(t *testing.T) {
	tests := []struct{
		offset uint64
		minDx int64
		x, out []uint64
	} {
		{0, 0, []uint64{}, []uint64{}},
		{0, 0, []uint64{10}, []uint64{10}},
		{10, 0, []uint64{10}, []uint64{0}},
		{0, 5, []uint64{10}, []uint64{5}},
		{0, 0, []uint64{1, 5, 5, 10, 16, 20}, []uint64{1, 4, 0, 5, 6, 4}},
		{0, -2, []uint64{1, 5, 5, 10, 16, 20}, []uint64{3, 6, 2, 7, 8, 6}},
		{10, -10, []uint64{5, 12, 10, 0}, []uint64{5, 17, 8, 0}},
	}

	for i := range tests {
		x := make([]uint64, len(tests[i].x))
		for j := range x { x[j] = tests[i].x[j] }

		DeltaEncode(tests[i].offset, tests[i].minDx, x, x)

		if !eq.Uint64s(tests[i].out, x) {
			t.Errorf("%d) Expected deltaEncode(offset=%d, minDx=%d, x=%d) to " + 
				"be %d, but got %d.", i, tests[i].offset, tests[i].minDx,
				tests[i].x, tests[i].out, x)
		}
	}
}

func TestDeltaDecode(t *testing.T) {
	tests := []struct{
		offset uint64
		minDx int64
		x, out []uint64
	} {
		{0, 0, []uint64{}, []uint64{}},
		{0, 0, []uint64{10}, []uint64{10}},
		{10, 0, []uint64{0}, []uint64{10}},
		{0, 5, []uint64{5}, []uint64{10}},
		{0, 0, []uint64{1, 4, 0, 5, 6, 4}, []uint64{1, 5, 5, 10, 16, 20}},
		{0, -2, []uint64{3, 6, 2, 7, 8, 6}, []uint64{1, 5, 5, 10, 16, 20}},
		{10, -10, []uint64{5, 17, 8, 0}, []uint64{5, 12, 10, 0}},
	}

	for i := range tests {
		x := make([]uint64, len(tests[i].x))
		for j := range x { x[j] = tests[i].x[j] }

		DeltaDecode(tests[i].offset, tests[i].minDx, x, x)

		if !eq.Uint64s(tests[i].out, x) {
			t.Errorf("%d) Expected deltaDecode(offset=%d, minDx=%d, x=%d) to " + 
				"be %d, but got %d.", i, tests[i].offset, tests[i].minDx,
				tests[i].x, tests[i].out, x)
		}
	}
}

func TestBlockToSlices(t *testing.T) {
	block0 := []uint64{
		0, 0, 0, 0, 0,
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
		
		 6,  7,  8,  9, 10,
		11, 12, 13, 14, 15,
		16, 17, 18, 19, 20,
		21, 22, 23, 24, 25,

		 6,  7,  8,  9, 10,
		11, 12, 13, 14, 15,
		16, 17, 18, 19, 20,
		21, 22, 23, 24, 25,
	}
	buf := make([]uint64, len(block0))
	span := [3]int{ 5, 4, 3 }
	slices := BlockToSlices(span, 0, block0, buf)

	for i := range slices {
		if i == 0 && len(slices[i]) != 5 {
			t.Errorf("Block 0: slice %d has length %d instead of 5.",
				i, len(slices[i]))
		} else if (i >= 1 && i < 6) && len(slices[i]) != 3 {
			t.Errorf("Block 0: slice %d has length %d instead of 5.",
				i, len(slices[i]))
		} else if i >= 6 && len(slices[i]) != 2 {
			t.Errorf("Block 0: slice %d has length %d instead of 5.",
				i, len(slices[i]))
		}

		for j := range slices[i] {
			if slices[i][j] != uint64(i) {
				t.Errorf("Block 0: slice %d = %d", i, slices[i])
			}
		}
	}

	block1 := []uint64{
		 0,  1,  1,  1,  1, 
		 4,  7, 10, 13, 16,
		 4,  7, 10, 13, 16,
		 4,  7, 10, 13, 16,

		 0,  2,  2,  2,  2, 
		 5,  8, 11, 14, 17,
		 5,  8, 11, 14, 17,
		 5,  8, 11, 14, 17,

		 0,  3,  3,  3,  3, 
		 6,  9, 12, 15, 18,
		 6,  9, 12, 15, 18,
		 6,  9, 12, 15, 18,
	}

	buf = make([]uint64, len(block1))
	span = [3]int{ 5, 4, 3 }
	slices = BlockToSlices(span, 2, block1, buf)

	for i := range slices {
		if i == 0 && len(slices[i]) != 3 {
			t.Errorf("Block 1: slice %d has length %d instead of 3.",
				i, len(slices[i]))
		} else if (i >= 1 && i < 4) && len(slices[i]) != 4 {
			t.Errorf("Block 1: slice %d has length %d instead of 4.",
				i, len(slices[i]))
		} else if i >= 4 && len(slices[i]) != 3 {
			t.Errorf("Block 1: slice %d has length %d instead of 3.",
				i, len(slices[i]))
		}

		for j := range slices[i] {
			if slices[i][j] != uint64(i) {
				t.Errorf("Block 1: slice %d = %d", i, slices[i])
			}
		}
	}
}

func TestSlicesToBlock(t *testing.T) {
	span := [3]int{ 5, 4, 3 }
	block := make([]uint64, span[0]*span[1]*span[2])
	buf := make([]uint64, len(block))
	result := make([]uint64, len(block))
	for i := range block { block[i] = uint64(i) }

	for firstDim := 0; firstDim < 3; firstDim++ {
		for i := range result { result[i] = 0 }

		slices := BlockToSlices(span, firstDim, block, buf)
		SlicesToBlock(span, firstDim, slices, result)

		if !eq.Uint64s(result, block) {
			t.Errorf("Output block is %d, but input block was %d",
				result, block)
		}
	}
}

func TestChooseFirstDim(t *testing.T) {
	tests := []struct{
		f particles.Field
		firstDim int
	} {
		{particles.NewUint64("x", []uint64{}), 0},
		{particles.NewUint64("y", []uint64{0, 0, 0}), 0},
		{particles.NewFloat64("z", []float64{}), 0},
		{particles.NewUint64("x[0]", []uint64{}), 0},
		{particles.NewUint64("x[1]", []uint64{}), 1},
		{particles.NewUint64("x[2]", []uint64{}), 2},
	}

	for i := range tests {
		firstDim := ChooseFirstDim(tests[i].f.Name())
		if firstDim != tests[i].firstDim {
			t.Errorf("%d) Expected ChooseFirstDim('%s') = %d, got %d",
				i, tests[i].f.Name(), tests[i].firstDim, firstDim)
		}
	}
}

func TestSliceOffsets(t *testing.T) {
	block0 := []uint64{
		30, 31, 32, 33, 34,
		 1,  2,  3,  4,  5,
		 1,  2,  3,  4,  5,
		 1,  2,  3,  4,  5,
		
		 6,  7,  8,  9, 10,
		11, 12, 13, 14, 15,
		16, 17, 18, 19, 20,
		21, 22, 23, 24, 25,

		 6,  7,  8,  9, 10,
		11, 12, 13, 14, 15,
		16, 17, 18, 19, 20,
		21, 22, 23, 24, 25,
	}

	buf := make([]uint64, len(block0))
	span := [3]int{ 5, 4, 3 }
	slices := BlockToSlices(span, 0, block0, buf)

	offsets := []uint64{
		30,
		30, 31, 32, 33, 34,
		30, 31, 32, 33, 34,
		 1,  2,  3,  4,  5,
		 1,  2,  3,  4,  5,
		 1,  2,  3,  4,  5,
	}

	testOffsets := SliceOffsets(slices)

	if !eq.Uint64s(offsets, testOffsets) {
		t.Errorf("Expected offsets = %d, got %d", offsets, testOffsets)
	}

	block1 := []uint64{
		 30,  1, 11, 21, 31, 
		  4,  7, 10, 13, 16,
		  4,  7, 10, 13, 16,
		  4,  7, 10, 13, 16,

		 31,  2, 12, 22, 32, 
		  5,  8, 11, 14, 17,
		  5,  8, 11, 14, 17,
		  5,  8, 11, 14, 17,

		 32,  3, 13, 23, 33, 
		  6,  9, 12, 15, 18,
		  6,  9, 12, 15, 18,
		  6,  9, 12, 15, 18,
	}

	buf = make([]uint64, len(block1))
	slices = BlockToSlices(span, 2, block1, buf)

	offsets = []uint64{
		30,
		30, 31, 32,
		30, 31, 32,
		 1,  2,  3,
		11, 12, 13,
		21, 22, 23,
		31, 32, 33,
	}

	testOffsets = SliceOffsets(slices)

	if !eq.Uint64s(offsets, testOffsets) {
		t.Errorf("Expected offsets = %d, got %d", offsets, testOffsets)
	}
}