package sudoku

import (
	"testing"
)

//TODO: test a few more puzzles to make sure I'm exercising it correctly.

func TestXYWing(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells: []cellRef{{3, 8}, {5, 3}},
		//TODO: figure out how to test that {3,6} (the pivot cell) comes first
		pointerCells: []cellRef{{3, 6}, {3, 3}, {5, 7}},
		targetNums:   IntSlice([]int{7}),
		//TODO: test description
	}
	humanSolveTechniqueTestHelper(t, "xywing_example.sdk", "XYWing", options)
	techniqueVariantsTestHelper(t, "XYWing")

}
