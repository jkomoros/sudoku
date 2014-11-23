package sudoku

import (
	"testing"
)

func TestXWingRow(t *testing.T) {
	//TODO: test descrption
	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{1, 4}, {1, 7}, {8, 4}},
		pointerCells: []cellRef{{0, 4}, {0, 7}, {7, 4}, {7, 7}},
		targetNums:   IntSlice([]int{9}),
	}
	humanSolveTechniqueTestHelper(t, "xwingtest.sdk", "XWing Row", options)

}
