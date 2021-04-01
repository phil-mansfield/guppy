package compress

import (
	"math"
)

var (
	xorshiftMaxUint = float64(math.MaxUint32)
)

// RNG is an xorshift random number generator. It is the same as gotetra's
// xorshiftGenerator. It is not thread safe.
type RNG struct {
	w, x, y, z uint32
}

// Init initializes RNG with a given seed.
func NewRNG(seed uint64) *RNG {
	return &RNG{ uint32(seed), 123456789, 362436069, 521288629 }
}

// Uniform generates a single random number in the range [0, 1)
func (gen *RNG) Uniform() float64 {
	t := gen.x ^ (gen.x << 11)
	gen.x, gen.y, gen.z = gen.y, gen.z, gen.w
	gen.w = gen.w ^ (gen.w >> 19) ^ (t ^ (t >> 8))
	res := float64(math.MaxUint32 - gen.w) / xorshiftMaxUint
	if res == 1.0 { return gen.Uniform() }
	return res
}

// UniformSeqeunce generates one random number in the range [0, 1) for each
// element of the array target and writes them to that array.
func (gen *RNG) UniformSequence(target []float64) {
	for i := 0; i < len(target); i++ {
		t := gen.x ^ (gen.x << 11)
		gen.x, gen.y, gen.z = gen.y, gen.z, gen.w
		gen.w = gen.w ^ (gen.w >> 19) ^ (t ^ (t >> 8))
		target[i] = float64(math.MaxUint32 - gen.w) / xorshiftMaxUint
		if target[i] == 1.0 { i-- }
	}
}
