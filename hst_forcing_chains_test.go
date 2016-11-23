package sudoku

import (
	"strconv"
	"testing"
)

func TestForcingChains(t *testing.T) {

	techniqueVariantsTestHelper(t, "Forcing Chain",
		"Forcing Chain (1 steps)",
		"Forcing Chain (2 steps)",
		"Forcing Chain (3 steps)",
		"Forcing Chain (4 steps)",
		"Forcing Chain (5 steps)",
		"Forcing Chain (6 steps)",
	)

	//TODO: the fact that every time we make a relatively small change to the forcing chain algo
	//we have to manually swizzle the test cases around reveals that the exact behavior of forcing
	//chains is fundamentally arbitrary. That makes me nervous. We should probably add a bunch more
	//tests.

	//Tester puzzle: http://www.komoroske.com/sudoku/index.php?puzzle=Q6Ur5iYGINSUFcyocqaY6G91DpttiqYzs

	tests := []multipleValidStepLoopOptions{
		{
			targetCells:  CellRefSlice{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: CellRefSlice{{1, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
			description:  "cell (1,0) only has two options, 1 and 2, and if you put either one in and see the chain of implications it leads to, both ones end up with 7 in cell (0,1), so we can just fill that number in",
			extra:        6,
		},
		{
			targetCells:  CellRefSlice{{0, 1}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: CellRefSlice{{5, 1}},
			pointerNums:  IntSlice([]int{1, 2}),
			//Explicitly don't test description after the first one.
			extra: 4,
		},
		//Another particularly long one
		{
			targetCells:  CellRefSlice{{1, 0}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: CellRefSlice{{0, 1}},
			pointerNums:  IntSlice([]int{2, 7}),
			extra:        5,
		},
		{
			targetCells:  []CellRef{{1, 0}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []CellRef{{0, 6}},
			pointerNums:  IntSlice([]int{3, 7}),
			extra:        5,
		},
		{
			targetCells:  []CellRef{{1, 8}},
			targetNums:   IntSlice([]int{4}),
			pointerCells: []CellRef{{1, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
			extra:        6,
		},
		{
			targetCells:  []CellRef{{1, 8}},
			targetNums:   IntSlice([]int{4}),
			pointerCells: []CellRef{{4, 0}},
			pointerNums:  IntSlice([]int{1, 2}),
			extra:        6,
		},
		{
			targetCells:  []CellRef{{1, 8}},
			targetNums:   IntSlice([]int{4}),
			pointerCells: []CellRef{{5, 1}},
			pointerNums:  IntSlice([]int{1, 2}),
			extra:        6,
		},
		{
			targetCells:  []CellRef{{1, 8}},
			targetNums:   IntSlice([]int{4}),
			pointerCells: []CellRef{{5, 7}},
			pointerNums:  IntSlice([]int{1, 3}),
			extra:        6,
		},
		{
			targetCells:  []CellRef{{4, 0}},
			targetNums:   IntSlice([]int{2}),
			pointerCells: []CellRef{{0, 1}},
			pointerNums:  IntSlice([]int{2, 7}),
			extra:        5,
		},
		{
			targetCells:  []CellRef{{4, 0}},
			targetNums:   IntSlice([]int{2}),
			pointerCells: []CellRef{{0, 6}},
			pointerNums:  IntSlice([]int{3, 7}),
			extra:        5,
		},
		{
			targetCells:  []CellRef{{4, 5}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []CellRef{{5, 4}},
			pointerNums:  IntSlice([]int{2, 3}),
			extra:        6,
		},
		{
			targetCells:  []CellRef{{5, 1}},
			targetNums:   IntSlice([]int{1}),
			pointerCells: []CellRef{{0, 1}},
			pointerNums:  IntSlice([]int{2, 7}),
			extra:        5,
		},
		{
			targetCells:  []CellRef{{8, 3}},
			targetNums:   IntSlice([]int{7}),
			pointerCells: []CellRef{{8, 7}},
			pointerNums:  IntSlice([]int{1, 2}),
			extra:        6,
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

	for i, _ := range tests {
		tests[i].variantName = "Forcing Chain (" + strconv.Itoa(tests[i].extra.(int)) + " steps)"
	}

	multipleValidStepsTestHelper(t, "forcingchain_test1.sdk", "Forcing Chain", tests)

	//TODO: test all other valid steps that could be found at this grid state for this technique.

}
