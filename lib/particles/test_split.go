package particles

import (
	"testing"
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
