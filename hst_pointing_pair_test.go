package sudoku

import (
	"testing"
)

func TestPointingPairCol(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{3, 1}, {4, 1}, {5, 1}, {6, 1}, {7, 1}, {8, 1}},
		pointerCells: []cellRef{{0, 1}, {2, 1}},
		targetSame:   GROUP_COL,
		targetGroup:  1,
		targetNums:   IntSlice([]int{7}),
		description:  "7 is only possible in column 1 of block 0, which means it can't be in any other cell in that column not in that block",
	}
	humanSolveTechniqueTestHelper(t, "pointingpaircol1.sdk", "Pointing Pair Col", options)

}

func TestPointingPairRow(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}},
		pointerCells: []cellRef{{1, 0}, {1, 2}},
		targetSame:   GROUP_ROW,
		targetGroup:  1,
		targetNums:   IntSlice([]int{7}),
		description:  "7 is only possible in row 1 of block 0, which means it can't be in any other cell in that row not in that block",
	}
	humanSolveTechniqueTestHelper(t, "pointingpairrow1.sdk", "Pointing Pair Row", options)

}
