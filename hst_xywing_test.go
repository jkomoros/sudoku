package sudoku

import (
	"testing"
)

//TODO: test a few more puzzles to make sure I'm exercising it correctly.

func TestXYWing(t *testing.T) {

	//TODO: this code is substantially recreated in forcingChainsTest. Factour
	//out into a new helper?

	techniqueVariantsTestHelper(t, "XYWing")

	options := solveTechniqueTestHelperOptions{
		checkAllSteps: true,
	}

	grid, solver, steps := humanSolveTechniqueTestHelperStepGenerator(t,
		"xywing_example.sdk", "XYWing", options)

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
			targetCells: []cellRef{{3, 8}, {5, 3}},
			//TODO: figure out how to test that {3,6} (the pivot cell) comes first
			pointerCells: []cellRef{{3, 6}, {3, 3}, {5, 7}},
			targetNums:   IntSlice([]int{7}),
			pointerNums:  IntSlice{2, 5, 7},
			description:  "(3,6) can only be two values, and cells (5,7) and (3,3) have those two possibilities, plus one other, so if you put either of the main cell's two possibiltiies in, it forces the intersection of the other two cells to not have 7",
		},
		{
			targetCells:  []cellRef{{3, 3}},
			targetNums:   IntSlice{5},
			pointerCells: []cellRef{{5, 7}, {3, 6}, {5, 3}},
			pointerNums:  IntSlice{2, 7, 5},
			//Explicitly don't test description after the first one.
		},
	}

	for _, test := range tests {

		options.targetCells = test.targetCells
		options.targetNums = test.targetNums
		options.pointerCells = test.pointerCells
		options.pointerNums = test.pointerNums
		options.description = test.description

		humanSolveTechniqueTestHelper(t, "xywing_example.sdk", "XYWing", options)
	}

	if len(tests) != len(steps) {
		t.Error("We didn't have enough tests for all of the steps that xywing returned. Got", len(tests), "expected", len(steps))
	}

}
