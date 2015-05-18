package sudoku

import (
	"testing"
)

func TestForcingChains(t *testing.T) {

	//Steps to test this:
	//* create a stepsToCheck list of steps we can pass in to humanSolveTEchniqueHelper,
	//so we can calculate it once and pass that in to the helper.
	//* In the forcing chain helper, calculate the steps once, then
	//pass them in each time in a list of ~10 calls to solveTechniqueTEstHelper that we know are valid here.
	//* VERIFY MANUALLY that each step that is returned is actually a valid application of forcingchains.

	/*
		//TODO: test description
		options := solveTechniqueTestHelperOptions{
			targetCells: []cellRef{{0, 1}},
			targetNums:  IntSlice([]int{7}),
			debugPrint:  true,
		}
		humanSolveTechniqueTestHelper(t, "forcingchain_test1.sdk", "Forcing Chain", options)
	*/
}
