package sudoku

import (
	"testing"
)

func TestForcingChains(t *testing.T) {

	//Steps to test this:
	//* In the forcing chain helper, calculate the steps once, then
	//pass them in each time in a list of ~10 calls to solveTechniqueTEstHelper that we know are valid here.
	//* VERIFY MANUALLY that each step that is returned is actually a valid application of forcingchains.

	/*
		//TODO: test description
		options := solveTechniqueTestHelperOptions{
			targetCells: []cellRef{{0, 1}},
			targetNums:  IntSlice([]int{7}),
			checkAllSteps: true,
			debugPrint:  true,
		}
		humanSolveTechniqueTestHelper(t, "forcingchain_test1.sdk", "Forcing Chain", options)
	*/
}
