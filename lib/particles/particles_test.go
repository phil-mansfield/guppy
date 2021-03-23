package particles

import (
	"testing"

	"github.com/phil-mansfield/guppy/lib/eq"
)

func TestUint32(t *testing.T) {
	out := []uint32{42, 0, 23, 0, 16, 0, 15, 0, 8, 0, 4, 0}
	data := []uint32{4, 8, 15, 16, 23, 42}
	from := []int{ 5, 4, 3, 2, 1, 0 }
	to := []int{ 0, 2, 4, 6, 8, 10 }
	name := "test_value"
		
	x := NewUint32(name, data)
	
	if x.Len() != len(data) {
		t.Errorf("Expected x.Len() = %d, got %d.", len(data), x.Len())
		return
	} else if !eq.Generic(data, x.Data()) {
		t.Errorf("Expected x.Data() = %v, got %v.", data, x.Data())
		return
	}

	p := Particles{ }
	
	x.CreateDestination(p, len(out))
	if _, ok := p[name]; !ok {
		t.Errorf("Expected Particles to gain '%s' field, but it wasn't added.",
			name)
		return
	}

	x.Transfer(p, from, to)
	if !eq.Generic(out, p[name].Data()) {
		t.Errorf("Expected p['%s'] = %v, got %v.", name, data, p[name].Data())
	}
}

func TestUint64(t *testing.T) {
	out := []uint64{42, 0, 23, 0, 16, 0, 15, 0, 8, 0, 4, 0}
	data := []uint64{4, 8, 15, 16, 23, 42}
	from := []int{ 5, 4, 3, 2, 1, 0 }
	to := []int{ 0, 2, 4, 6, 8, 10 }
	name := "test_value"
		
	x := NewUint64(name, data)
	
	if x.Len() != len(data) {
		t.Errorf("Expected x.Len() = %d, got %d.", len(data), x.Len())
		return
	} else if !eq.Generic(data, x.Data()) {
		t.Errorf("Expected x.Data() = %v, got %v.", data, x.Data())
		return
	}

	p := Particles{ }
	
	x.CreateDestination(p, len(out))
	if _, ok := p[name]; !ok {
		t.Errorf("Expected Particles to gain '%s' field, but it wasn't added.",
			name)
		return
	}

	x.Transfer(p, from, to)
	if !eq.Generic(out, p[name].Data()) {
		t.Errorf("Expected p['%s'] = %v, got %v.", name, data, p[name].Data())
	}
}

func TestFloat32(t *testing.T) {
	out := []float32{42, 0, 23, 0, 16, 0, 15, 0, 8, 0, 4, 0}
	data := []float32{4, 8, 15, 16, 23, 42}
	from := []int{ 5, 4, 3, 2, 1, 0 }
	to := []int{ 0, 2, 4, 6, 8, 10 }
	name := "test_value"
		
	x := NewFloat32(name, data)
	
	if x.Len() != len(data) {
		t.Errorf("Expected x.Len() = %d, got %d.", len(data), x.Len())
		return
	} else if !eq.Generic(data, x.Data()) {
		t.Errorf("Expected x.Data() = %v, got %v.", data, x.Data())
		return
	}

	p := Particles{ }
	
	x.CreateDestination(p, len(out))
	if _, ok := p[name]; !ok {
		t.Errorf("Expected Particles to gain '%s' field, but it wasn't added.",
			name)
		return
	}

	x.Transfer(p, from, to)
	if !eq.Generic(out, p[name].Data()) {
		t.Errorf("Expected p['%s'] = %v, got %v.", name, data, p[name].Data())
	}
}

func TestFloat64(t *testing.T) {
	out := []float64{42, 0, 23, 0, 16, 0, 15, 0, 8, 0, 4, 0}
	data := []float64{4, 8, 15, 16, 23, 42}
	from := []int{ 5, 4, 3, 2, 1, 0 }
	to := []int{ 0, 2, 4, 6, 8, 10 }
	name := "test_value"
		
	x := NewFloat64(name, data)
	
	if x.Len() != len(data) {
		t.Errorf("Expected x.Len() = %d, got %d.", len(data), x.Len())
		return
	} else if !eq.Generic(data, x.Data()) {
		t.Errorf("Expected x.Data() = %v, got %v.", data, x.Data())
		return
	}

	p := Particles{ }
	
	x.CreateDestination(p, len(out))
	if _, ok := p[name]; !ok {
		t.Errorf("Expected Particles to gain '%s' field, but it wasn't added.",
			name)
		return
	}

	x.Transfer(p, from, to)
	if !eq.Generic(out, p[name].Data()) {
		t.Errorf("Expected p['%s'] = %v, got %v.", name, data, p[name].Data())
	}
}

func TestVec32(t *testing.T) {
	out := [][3]float32{{42}, {0}, {23}, {0}, {16}, {0}, {15},
		{0}, {8}, {0}, {4}, {0}}
	data := [][3]float32{{4}, {8}, {15}, {16}, {23}, {42}}
	from := []int{ 5, 4, 3, 2, 1, 0 }
	to := []int{ 0, 2, 4, 6, 8, 10 }
	name := "test_value"
		
	x := NewVec32(name, data)
	
	if x.Len() != len(data) {
		t.Errorf("Expected x.Len() = %d, got %d.", len(data), x.Len())
		return
	} else if !eq.Generic(data, x.Data()) {
		t.Errorf("Expected x.Data() = %v, got %v.", data, x.Data())
		return
	}

	p := Particles{ }
	
	x.CreateDestination(p, len(out))
	if _, ok := p[name]; !ok {
		t.Errorf("Expected Particles to gain '%s' field, but it wasn't added.",
			name)
		return
	}

	x.Transfer(p, from, to)
	if !eq.Generic(out, p[name].Data()) {
		t.Errorf("Expected p['%s'] = %v, got %v.", name, data, p[name].Data())
	}
}

func TestVec64(t *testing.T) {
	out := [][3]float64{{42}, {0}, {23}, {0}, {16}, {0}, {15},
		{0}, {8}, {0}, {4}, {0}}
	data := [][3]float64{{4}, {8}, {15}, {16}, {23}, {42}}
	from := []int{ 5, 4, 3, 2, 1, 0 }
	to := []int{ 0, 2, 4, 6, 8, 10 }
	name := "test_value"
		
	x := NewVec64(name, data)
	
	if x.Len() != len(data) {
		t.Errorf("Expected x.Len() = %d, got %d.", len(data), x.Len())
		return
	} else if !eq.Generic(data, x.Data()) {
		t.Errorf("Expected x.Data() = %v, got %v.", data, x.Data())
		return
	}

	p := Particles{ }
	
	x.CreateDestination(p, len(out))
	if _, ok := p[name]; !ok {
		t.Errorf("Expected Particles to gain '%s' field, but it wasn't added.",
			name)
		return
	}

	x.Transfer(p, from, to)
	if !eq.Generic(out, p[name].Data()) {
		t.Errorf("Expected p['%s'] = %v, got %v.", name, data, p[name].Data())
	}
}
