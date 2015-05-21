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

	type loopOptions struct {
		targetCells  []cellRef
		targetNums   IntSlice
		pointerCells []cellRef
		pointerNums  IntSlice
		description  string
	}

	tests := []loopOptions{
		{
			targetCells:  []cellRef{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{1, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
			description:  "cell (1,0) only has two options, 1 and 2, and if you put either one in and see the chain of implications it leads to, both ones end up with 7 in cell (0,1), so we can just fill that number in",
		},
		{
			targetCells:  []cellRef{{1, 0}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []cellRef{{0, 6}},
			pointerNums:  IntSlice([]int{3, 7}),
			//Explicitly don't test description after the first one.
		},
	}

	if len(tests) != len(steps) {
		t.Error("We didn't have enough tests for all of the steps that forcing chains returned. Got", len(tests), "expected", len(steps))
	}

	for _, test := range tests {

		options.targetCells = test.targetCells
		options.targetNums = test.targetNums
		options.pointerCells = test.pointerCells
		options.pointerNums = test.pointerNums
		options.description = test.description

		humanSolveTechniqueTestHelper(t, "forcingchain_test1.sdk", "Forcing Chain", options)
	}

	//TODO: test all other valid steps that could be found at this grid state for this technique.

}
