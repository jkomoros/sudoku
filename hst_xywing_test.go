package sudoku

import (
	"testing"
)

//TODO: test a few more puzzles to make sure I'm exercising it correctly.

func TestXYWing(t *testing.T) {

	techniqueVariantsTestHelper(t, "XYWing", "XYWing", "XYWing (Same Block)")

	tests := []multipleValidStepLoopOptions{
		{
			targetCells: []cellRef{{3, 8}, {5, 3}},
			//TODO: figure out how to test that {3,6} (the pivot cell) comes first
			pointerCells: []cellRef{{3, 6}, {3, 3}, {5, 7}},
			targetNums:   IntSlice([]int{7}),
			pointerNums:  IntSlice{2, 5, 7},
			description:  "(3,6) can only be two values, and cells (5,7) and (3,3) have those two possibilities, plus one other, so if you put either of the main cell's two possibiltiies in, it forces the intersection of the other two cells to not have 7",
		},
		{
			targetCells: []cellRef{{3, 8}},
			//TODO: figure out how to test that {3,6} (the pivot cell) comes first
			pointerCells: []cellRef{{3, 6}, {3, 3}, {5, 7}},
			targetNums:   IntSlice([]int{7}),
			pointerNums:  IntSlice{2, 5, 7},
			//Explicitly don't test description after the first one.
			variantName: "XYWing (Same Block)",
		},
		{
			targetCells: []cellRef{{5, 3}},
			//TODO: figure out how to test that {3,6} (the pivot cell) comes first
			pointerCells: []cellRef{{3, 6}, {3, 3}, {5, 7}},
			targetNums:   IntSlice([]int{7}),
			pointerNums:  IntSlice{2, 5, 7},
			variantName:  "XYWing (Same Block)",
		},
		{
			targetCells:  []cellRef{{3, 3}},
			targetNums:   IntSlice{5},
			pointerCells: []cellRef{{5, 7}, {3, 6}, {5, 3}},
			pointerNums:  IntSlice{2, 7, 5},
			variantName:  "XYWing (Same Block)",
		},
	}

	multipleValidStepsTestHelper(t, "xywing_example.sdk", "XYWing", tests)

}
