package sudoku

import (
	"testing"
)

//TODO: test a few more puzzles to make sure I'm exercising it correctly.

func TestXWingRow(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{1, 4}, {1, 7}, {8, 4}},
		pointerCells: []cellRef{{0, 4}, {0, 7}, {7, 4}, {7, 7}},
		targetNums:   IntSlice([]int{9}),
		description:  "in rows 0 and 7, 9 is only possible in columns 4 and 7, and 9 must be in one of those cells per rows, so it can't be in any other cells in those columns",
	}
	humanSolveTechniqueTestHelper(t, "xwingtest.sdk", "XWing Row", options)
	techniqueVariantsTestHelper(t, "XWing Row")

}

func TestXWingCol(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		transpose:    true,
		targetCells:  []cellRef{{4, 1}, {7, 1}, {4, 8}},
		pointerCells: []cellRef{{4, 0}, {7, 0}, {4, 7}, {7, 7}},
		targetNums:   IntSlice([]int{9}),
		description:  "in columns 0 and 7, 9 is only possible in rows 4 and 7, and 9 must be in one of those cells per columns, so it can't be in any other cells in those rows",
	}
	humanSolveTechniqueTestHelper(t, "xwingtest.sdk", "XWing Col", options)
	techniqueVariantsTestHelper(t, "XWing Col")

}
