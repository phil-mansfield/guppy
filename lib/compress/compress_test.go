package compress

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/phil-mansfield/guppy/lib/eq"
	"github.com/phil-mansfield/guppy/lib/particles"
)

func TestBuffer(t *testing.T) {
	tests := []int{ 0, 10, 0, 10, 20, 30, 30, 30, 10, 10000, 0}
	buf := NewBuffer()
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
		test("u64_2", len(buf.u64_2), cap(buf.u64_2))
		test("f32", len(buf.f32), cap(buf.f32))
		test("f64", len(buf.f64), cap(buf.f64))

		prevLen = len(buf.u32)
	}
}

func TestLagrangianDelta(t *testing.T) {
	order := binary.LittleEndian
	
	tests := []struct{
		span [3]int
		dir int
		delta float64
		data interface{}
	} {
		{ [3]int{0, 0, 0}, 0, 0, []uint32{} },
		{ [3]int{0, 0, 0}, 0, 0, []uint64{} },
		{ [3]int{0, 0, 0}, 0, 0, []float32{} },
		{ [3]int{0, 0, 0}, 0, 0, []float64{} },
	}

	buf := NewBuffer()
	for i := range tests {
		m := NewLagrangianDelta(
			order, tests[i].span, tests[i].dir, tests[i].delta,
		)
		f, err := particles.NewGenericField("meow", tests[i].data)
		if err != nil { t.Errorf(err.Error()) }
		wr := &bytes.Buffer{ }

		err = m.WriteInfo(wr)
		if err != nil {
			t.Errorf("%d) Got error '%s' on WriteInfo", i, err.Error())
		}

		err = m.Compress(f, buf, wr)
		if err != nil {
			t.Errorf("%d) Got error '%s' on Compress", i, err.Error())
		}

		rd := bytes.NewReader(wr.Bytes())
		mOut := &LagrangianDelta{ }

		err = mOut.ReadInfo(order, rd)
		if err != nil {
			t.Errorf("%d) Got error '%s' on ReadInfo", i, err.Error())
		}

		fOut, err := mOut.Decompress(buf, rd)
		if err != nil {
			t.Errorf("%d) Got error '%s' on Demcompress", i, err.Error())
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
		
		x := f.Data()
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