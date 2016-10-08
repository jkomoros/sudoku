package sudoku

import (
	"testing"
)

func TestGuessTechnique(t *testing.T) {

	techniqueVariantsTestHelper(t, "Guess")

	//TODO: this test doesn't exercise whether pointerNums is the right value. The test case to do so would be too hard.
	options := solveTechniqueTestHelperOptions{
		matchMode:   solveTechniqueMatchModeAny,
		targetCells: []CellReference{{0, 0}, {1, 0}},
		targetNums:  IntSlice([]int{3, 7}),
		descriptions: []string{
			"we have no other moves to make, so we randomly pick a cell with the smallest number of possibilities, (0,0), and pick one of its possibilities",
			"we have no other moves to make, so we randomly pick a cell with the smallest number of possibilities, (1,0), and pick one of its possibilities"},
	}
	humanSolveTechniqueTestHelper(t, "guesstestgrid.sdk", "Guess", options)
}
