/*package eq is a simple package for telling whether two arrays are equal to
one another.*/
package eq

// Generic returns true if two arrays are the same type and have the same values
// and false otherwise. Only []byte, []string, []uint32, []uint64, []float32,
// []float64, [][3]float32, [][3]float64.
func Generic(x, y interface{}) bool {
	switch xx := x.(type) {
	case []byte:
		yy, ok := y.([]byte)
		if !ok { return false }
		return Bytes(xx, yy)
	case []int:
		yy, ok := y.([]int)
		if !ok { return false }
		return Ints(xx, yy)
	case []string:
		yy, ok := y.([]string)
		if !ok { return false }
		return Strings(xx, yy)
	case []float32:
		yy, ok := y.([]float32)
		if !ok { return false }
		return Float32s(xx, yy)
	case []float64:
		yy, ok := y.([]float64)
		if !ok { return false }
		return Float64s(xx, yy)
	case []uint32:
		yy, ok := y.([]uint32)
		if !ok { return false }
		return Uint32s(xx, yy)
	case []uint64:
		yy, ok := y.([]uint64)
		if !ok { return false }
		return Uint64s(xx, yy)
	case [][3]float32:
		yy, ok := y.([][3]float32)
		if !ok { return false }
		return Vec32s(xx, yy)
	case [][3]float64:
		yy, ok := y.([][3]float64)
		if !ok { return false }
		return Vec64s(xx, yy)
	default:
		return false
	}
	return false
}

// Strings returns true if two []string arrays are the same and false otherwise.
func Strings(x, y []string) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

// Bytes returns true if two []byte arrays are the same and false otherwise.
func Bytes(x, y []byte) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

// Ints returns true if two []int arrays are the same and false otherwise.
func Ints(x, y []int) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}


// Uint32s returns true if two []uint32 arrays are the same and false otherwise.
func Uint32s(x, y []uint32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

// Uint64s returns true if two []uint64 arrays are the same and false otherwise.
func Uint64s(x, y []uint64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

// Float32s returns true if two []float32 arrays are the same and false
// otherwise.
func Float32s(x, y []float32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

// Float64s returns true if two []float64 arrays are the same and false
// otherwise.
func Float64s(x, y []float64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

// Vec32s returns true if two [][3]float32 arrays are the same and false
// otherwise.
func Vec32s(x, y [][3]float32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

// Vec64s returns true if two [][3]float64 arrays are the same and false
// otherwise.
func Vec64s(x, y [][3]float64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

// Float32sEps returns true if the two []float32 arrays are within eps of one
// another and false otherwise.
func Float32sEps(x, y []float32, eps float32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] + eps < y[i] || x[i] - eps > y[i] {
			return false
		}
	}
	return true
}

// Float64sEps returns true if the two []float64 arrays are within eps of one
// another and false otherwise.
func Float64sEps(x, y []float64, eps float64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] + eps < y[i] || x[i] - eps > y[i] {
			return false
		}
	}
	return true
}
