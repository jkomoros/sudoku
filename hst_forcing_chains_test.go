package sudoku

import (
	"testing"
)

func TestForcingChains(t *testing.T) {

	//Steps to test this:
	//* In the forcing chain helper, calculate the steps once, then
	//pass them in each time in a list of ~10 calls to solveTechniqueTEstHelper that we know are valid here.
	//* VERIFY MANUALLY that each step that is returned is actually a valid application of forcingchains.

	options := solveTechniqueTestHelperOptions{
		checkAllSteps: true,
	}

	grid, solver, steps := humanSolveTechniqueTestHelperStepGenerator(t,
		"forcingchain_test1.sdk", "Forcing Chain", options)

	options.stepsToCheck.grid = grid
	options.stepsToCheck.solver = solver
	options.stepsToCheck.steps = steps

	//OK, now we'll walk through all of the options in a loop and make sure they all show
	//up in the solve steps.

	options.targetCells = []cellRef{{0, 1}}
	options.targetNums = IntSlice([]int{7})

	//TODO: test description

	humanSolveTechniqueTestHelper(t, "forcingchain_test1.sdk", "Forcing Chain", options)

	//TODO: test all other valid steps that could be found at this grid state for this technique.

}
