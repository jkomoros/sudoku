package sudoku

import (
	"log"
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

	//Tester puzzle: http://www.komoroske.com/sudoku/index.php?puzzle=Q6Ur5iYGINSUFcyocqaY6G91DpttiqYzs

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
		{
			targetCells:  []cellRef{{5, 1}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []cellRef{{0, 6}},
			pointerNums:  IntSlice([]int{3, 7}),
		},
		{
			targetCells:  []cellRef{{8, 3}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{7, 8}},
			pointerNums:  IntSlice([]int{2, 7}),
		},
		//TODO: update this comment
		//Skipping 0,1 / 7 / 5,7 / 1,3 because I think implications force it wrong.
		//The reason is because there is an inconsistency down that branch... just one implication
		//step beyond when it finds a match. Hmmmm... interesting test case.
		//Of course, the number being filled _IS_ right... is that a coincidence? I wondder
		//if other ones in this set that I'm considering valid would ahve the same problem...
		//... I think they would, right? By definition one of hte branches leads to invalidity. This
		//technique is about finding a universal before you find that invalidity.
		{
			targetCells:  []cellRef{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{5, 7}},
			pointerNums:  IntSlice([]int{1, 3}),
		},
		{
			targetCells:  []cellRef{{8, 3}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{5, 7}},
			pointerNums:  IntSlice([]int{1, 3}),
		},
		{
			targetCells:  []cellRef{{1, 8}},
			targetNums:   IntSlice([]int{4}),
			pointerCells: []cellRef{{4, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
		},
		//This next one's particularly long implication chain
		{
			targetCells:  []cellRef{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{4, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
		},
		//Another particularly long one
		{
			targetCells:  []cellRef{{1, 0}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []cellRef{{0, 1}},
			pointerNums:  IntSlice([]int{2, 7}),
		},
		{
			targetCells:  []cellRef{{5, 1}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []cellRef{{0, 1}},
			pointerNums:  IntSlice([]int{2, 7}),
		},

		//All of these are missing... what?
		//Oh, we fail as soon as we notice they don't all match.
		//We haven't seen this set again... flakey?
		//Next step: do the manual check for a 'normal' run to see which is missing
		// 0,6 /3 / 1,0 / 1,2
		// 0,1 / 7 / 5,1 / 1,2
		// 0,6 / 3 / 5,1 / 1,2
		//8,3 / 7 / 8,7 / 1,2
		// 0,1 /7 / 5,4 / 2,3
		// 8,3 / 7 / 5,4 / 2,3

		{
			targetCells:  []cellRef{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{1, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
		},
		//Another particularly long one
		{
			targetCells:  []cellRef{{5, 1}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []cellRef{{0, 6}},
			pointerNums:  IntSlice([]int{3, 7}),
		},
		//Another particularly long one
		{
			targetCells:  []cellRef{{1, 0}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []cellRef{{0, 6}},
			pointerNums:  IntSlice([]int{3, 7}),
		},
		//Another particularly long one
		{
			targetCells:  []cellRef{{5, 1}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []cellRef{{0, 6}},
			pointerNums:  IntSlice([]int{3, 7}),
		},
		//Another particularly long one
		{
			targetCells:  []cellRef{{8, 3}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{7, 8}},
			pointerNums:  IntSlice([]int{2, 7}),
		},
		{
			targetCells:  []cellRef{{1, 8}},
			targetNums:   IntSlice([]int{4}),
			pointerCells: []cellRef{{1, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
		},
	}

	for _, step := range steps {
		log.Println(step)
	}

	for _, test := range tests {

		options.targetCells = test.targetCells
		options.targetNums = test.targetNums
		options.pointerCells = test.pointerCells
		options.pointerNums = test.pointerNums
		options.description = test.description

		humanSolveTechniqueTestHelper(t, "forcingchain_test1.sdk", "Forcing Chain", options)
	}

	if len(tests) != len(steps) {
		t.Error("We didn't have enough tests for all of the steps that forcing chains returned. Got", len(tests), "expected", len(steps))
	}

	//TODO: test all other valid steps that could be found at this grid state for this technique.

}
