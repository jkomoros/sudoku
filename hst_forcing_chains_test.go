package sudoku

import (
	"testing"
)

func TestForcingChains(t *testing.T) {

	techniqueVariantsTestHelper(t, "Forcing Chain")

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
		targetCells         []cellRef
		targetNums          IntSlice
		pointerCells        []cellRef
		pointerNums         IntSlice
		description         string
		numImplicationSteps int
	}

	//TODO: the fact that every time we make a relatively small change to the forcing chain algo
	//we have to manually swizzle the test cases around reveals that the exact behavior of forcing
	//chains is fundamentally arbitrary. That makes me nervous. We should probably add a bunch more
	//tests.

	//Tester puzzle: http://www.komoroske.com/sudoku/index.php?puzzle=Q6Ur5iYGINSUFcyocqaY6G91DpttiqYzs

	tests := []loopOptions{
		{
			targetCells:         []cellRef{{0, 1}},
			targetNums:          IntSlice([]int{7}),
			pointerCells:        []cellRef{{1, 0}},
			pointerNums:         IntSlice([]int{1, 2}),
			description:         "cell (1,0) only has two options, 1 and 2, and if you put either one in and see the chain of implications it leads to, both ones end up with 7 in cell (0,1), so we can just fill that number in",
			numImplicationSteps: 6,
		},
		{
			targetCells:  []cellRef{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{5, 1}},
			pointerNums:  IntSlice([]int{1, 2}),
			//Explicitly don't test description after the first one.
			numImplicationSteps: 4,
		},
		//Another particularly long one
		{
			targetCells:         []cellRef{{1, 0}},
			targetNums:          IntSlice([]int{1}),
			pointerCells:        []cellRef{{0, 1}},
			pointerNums:         IntSlice([]int{2, 7}),
			numImplicationSteps: 5,
		},
		{
			targetCells:         []cellRef{{1, 0}},
			targetNums:          IntSlice([]int{1}),
			pointerCells:        []cellRef{{0, 6}},
			pointerNums:         IntSlice([]int{3, 7}),
			numImplicationSteps: 5,
		},
		{
			targetCells:         []cellRef{{1, 8}},
			targetNums:          IntSlice([]int{4}),
			pointerCells:        []cellRef{{1, 0}},
			pointerNums:         IntSlice([]int{1, 2}),
			numImplicationSteps: 6,
		},
		{
			targetCells:         []cellRef{{1, 8}},
			targetNums:          IntSlice([]int{4}),
			pointerCells:        []cellRef{{4, 0}},
			pointerNums:         IntSlice([]int{1, 2}),
			numImplicationSteps: 6,
		},
		{
			targetCells:         []cellRef{{1, 8}},
			targetNums:          IntSlice([]int{4}),
			pointerCells:        []cellRef{{5, 1}},
			pointerNums:         IntSlice([]int{1, 2}),
			numImplicationSteps: 6,
		},
		{
			targetCells:         []cellRef{{1, 8}},
			targetNums:          IntSlice([]int{4}),
			pointerCells:        []cellRef{{5, 7}},
			pointerNums:         IntSlice([]int{1, 3}),
			numImplicationSteps: 6,
		},
		{
			targetCells:         []cellRef{{4, 0}},
			targetNums:          IntSlice([]int{2}),
			pointerCells:        []cellRef{{0, 1}},
			pointerNums:         IntSlice([]int{2, 7}),
			numImplicationSteps: 5,
		},
		{
			targetCells:         []cellRef{{4, 0}},
			targetNums:          IntSlice([]int{2}),
			pointerCells:        []cellRef{{0, 6}},
			pointerNums:         IntSlice([]int{3, 7}),
			numImplicationSteps: 5,
		},
		{
			targetCells:         []cellRef{{4, 5}},
			targetNums:          IntSlice([]int{7}),
			pointerCells:        []cellRef{{5, 4}},
			pointerNums:         IntSlice([]int{2, 3}),
			numImplicationSteps: 6,
		},
		{
			targetCells:         []cellRef{{5, 1}},
			targetNums:          IntSlice([]int{1}),
			pointerCells:        []cellRef{{0, 1}},
			pointerNums:         IntSlice([]int{2, 7}),
			numImplicationSteps: 5,
		},
		{
			targetCells:         []cellRef{{8, 3}},
			targetNums:          IntSlice([]int{7}),
			pointerCells:        []cellRef{{8, 7}},
			pointerNums:         IntSlice([]int{1, 2}),
			numImplicationSteps: 6,
		},

		/* Steps that got dropped out when we switched to DFS
		//This next one's particularly long implication chain
		{
			targetCells:  []cellRef{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{4, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
		},
		{
			targetCells:  []cellRef{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{5, 7}},
			pointerNums:  IntSlice([]int{1, 3}),
		},

		*/

		/* Steps that dropped out when we switched to backwards intersect
		{
			targetCells:  []cellRef{{0, 6}},
			targetNums:   IntSlice([]int{3}),
			pointerCells: []cellRef{{5, 1}},
			pointerNums:  IntSlice([]int{1, 2}),
		},
		{
			targetCells:  []cellRef{{5, 1}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []cellRef{{0, 6}},
			pointerNums:  IntSlice([]int{3, 7}),
		},

		*/

		/* Steps that are too long now
		{
			targetCells:  []cellRef{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{5, 4}},
			pointerNums:  IntSlice([]int{2, 3}),
		},
				{
			targetCells:  []cellRef{{8, 3}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{5, 4}},
			pointerNums:  IntSlice([]int{2, 3}),
		},
		{
			targetCells:  []cellRef{{8, 3}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{5, 7}},
			pointerNums:  IntSlice([]int{1, 3}),
		},

		{
			targetCells:  []cellRef{{0, 6}},
			targetNums:   IntSlice([]int{3}),
			pointerCells: []cellRef{{1, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
		},
		{
			targetCells:  []cellRef{{8, 3}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []cellRef{{7, 8}},
			pointerNums:  IntSlice([]int{2, 7}),
		},
		*/

		/*
			Steps that are valid, but that we don't expect the technique to find
			right now.
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

			{
				targetCells:  []cellRef{{8, 3}},
				targetNums:   IntSlice([]int{7}),
				pointerCells: []cellRef{{5, 7}},
				pointerNums:  IntSlice([]int{1, 3}),
			},
		*/
	}

	/*
		//Temp code that helps debug why the test isn't passing

		var stepDescriptions []string

		for _, step := range steps {
			stepDescriptions = append(stepDescriptions, fmt.Sprint(step.TargetCells.Description(), step.TargetNums.Description(), step.PointerCells.Description(), step.PointerNums.Description()))
		}

		sort.Strings(stepDescriptions)

		for _, str := range stepDescriptions {
			fmt.Println(str)
		}
	*/

	for _, test := range tests {

		options.targetCells = test.targetCells
		options.targetNums = test.targetNums
		options.pointerCells = test.pointerCells
		options.pointerNums = test.pointerNums
		options.description = test.description
		options.extra = test.numImplicationSteps

		humanSolveTechniqueTestHelper(t, "forcingchain_test1.sdk", "Forcing Chain", options)
	}

	if len(tests) != len(steps) {
		t.Error("We didn't have enough tests for all of the steps that forcing chains returned. Got", len(tests), "expected", len(steps))
	}

	//TODO: test all other valid steps that could be found at this grid state for this technique.

}
