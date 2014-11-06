package sudoku

import (
	"testing"
)

//TODO: test guess technique here
func TestGuessTechnique(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		matchMode:   solveTechniqueMatchModeAny,
		targetCells: []cellRef{{0, 0}, {1, 0}},
		targetNums:  IntSlice([]int{3, 7}),
		descriptions: []string{
			"we have no other moves to make, so we randomly pick a cell with the smallest number of possibilities, (0,0), and pick one of its possibilities",
			"we have no other moves to make, so we randomly pick a cell with the smallest number of possibilities, (0,1), and pick one of its possibilities"},
	}
	humanSolveTechniqueTestHelper(t, "guesstestgrid.sdk", "Guess", options)
}
