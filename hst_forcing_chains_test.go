package sudoku

import (
	"testing"
)

func TestForcingChains(t *testing.T) {

	//Steps to test this:
	//* Make a getSteps (all/one) util method for a grid and technique. it runs
	//and returns either a one-len array or a full array of all options it found.
	//* Use getSteps in humanSolveTechniqueHelper
	//* Configure humanSolveTechniqueHelper to have a checkAllSteps option, that if
	//true will pull all steps from getSteps and check them.
	//* create a stepsToCheck list of steps we can pass in to humanSolveTEchniqueHelper,
	//so we can calculate it once and pass that in to the helper.
	//* In the forcing chain helper, calculate the steps once, then
	//pass them in each time in a list of ~10 calls to solveTechniqueTEstHelper that we know are valid here.

	//TODO: test description
	options := solveTechniqueTestHelperOptions{
		targetCells: []cellRef{{0, 1}},
		targetNums:  IntSlice([]int{7}),
		debugPrint:  true,
	}
	humanSolveTechniqueTestHelper(t, "forcingchain_test1.sdk", "Forcing Chain", options)
}
