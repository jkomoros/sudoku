package sudoku

import (
	"testing"
)

func TestPointingPairCol(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells:  []CellRef{{3, 1}, {4, 1}, {5, 1}, {6, 1}, {7, 1}, {8, 1}},
		pointerCells: []CellRef{{0, 1}, {2, 1}},
		targetSame:   _GROUP_COL,
		targetGroup:  1,
		targetNums:   IntSlice([]int{7}),
		description:  "7 is only possible in column 1 of block 0, which means it can't be in any other cell in that column not in that block",
	}
	humanSolveTechniqueTestHelper(t, "pointingpaircol1.sdk", "Pointing Pair Col", options)
	techniqueVariantsTestHelper(t, "Pointing Pair Col")

}

func TestPointingPairRow(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells:  []CellRef{{1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}},
		pointerCells: []CellRef{{1, 0}, {1, 2}},
		targetSame:   _GROUP_ROW,
		targetGroup:  1,
		targetNums:   IntSlice([]int{7}),
		description:  "7 is only possible in row 1 of block 0, which means it can't be in any other cell in that row not in that block",
	}
	humanSolveTechniqueTestHelper(t, "pointingpairrow1.sdk", "Pointing Pair Row", options)
	techniqueVariantsTestHelper(t, "Pointing Pair Row")

}
