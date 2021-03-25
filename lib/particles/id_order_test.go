package particles

import (
	"testing"
)

func TestZMajorUnigridLevels(t *testing.T) {
	order := NewZMajorUnigrid(10)
	if levels := order.Levels(); levels != 1 {
		t.Errorf("Expected order.Levels() = 1, got %d", levels)
	} else if span := order.LevelSpan(0); span != [3]int{ 10, 10, 10 } {
		t.Errorf("Expected order.LevelSpan(0) = [10 10 10], got %d", span)
	}  else if org := order.LevelOrigin(0); org != [3]int{ 0, 0, 0 } {
		t.Errorf("Expected order.LevelOrigin(0) = [10 10 10], got %d", org)
	}
}

func TestZMajorUnigridIndex(t *testing.T) {
	n := 10
	order := NewZMajorUnigrid(n)
	tests := []struct{
		idx [3]int
		id uint64
	} {
		{[3]int{0, 0, 0}, 0},
		{[3]int{9, 9, 9}, 999},
		{[3]int{1, 1, 1}, 111},
		{[3]int{3, 2, 1}, 321},
	}

	for i := range tests {
		id := order.IndexToID(tests[i].idx, 0)
		idx, level := order.IDToIndex(tests[i].id)
		if level != 0 {
			t.Errorf("%d) Expected id %d to have level %d, got %d",
				i, tests[i].id, 0, level)
		} else if id != tests[i].id {
			t.Errorf("%d) Expected index %d to have id %d, got %d.",
				i, tests[i].idx, tests[i].id, id)
		} else if idx != tests[i].idx {
			t.Errorf("%d) Expected id %d to have index %d, got %d.",
				i, tests[i].id, tests[i].idx, id)
		}
	}
}
