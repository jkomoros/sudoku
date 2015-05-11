package sudoku

import (
	"testing"
)

func TestForcingChains(t *testing.T) {
	//TODO: test description
	options := solveTechniqueTestHelperOptions{
		targetCells: []cellRef{{0, 1}},
		targetNums:  IntSlice([]int{7}),
		debugPrint:  true,
	}
	humanSolveTechniqueTestHelper(t, "forcingchain_test1.sdk", "Forcing Chain", options)
}
