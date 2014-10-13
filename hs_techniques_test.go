package sudoku

import (
	"testing"
)

func TestSubsetIndexes(t *testing.T) {
	result := subsetIndexes(3, 1)
	expectedResult := [][]int{[]int{0}, []int{1}, []int{2}}
	subsetIndexHelper(t, result, expectedResult)

	result = subsetIndexes(3, 2)
	expectedResult = [][]int{[]int{0, 1}, []int{0, 2}, []int{1, 2}}
	subsetIndexHelper(t, result, expectedResult)

	result = subsetIndexes(5, 3)
	expectedResult = [][]int{[]int{0, 1, 2}, []int{0, 1, 3}, []int{0, 1, 4}, []int{0, 2, 3}, []int{0, 2, 4}, []int{0, 3, 4}, []int{1, 2, 3}, []int{1, 2, 4}, []int{1, 3, 4}, []int{2, 3, 4}}
	subsetIndexHelper(t, result, expectedResult)

	if subsetIndexes(1, 2) != nil {
		t.Log("Subset indexes returned a subset where the length is greater than the len")
		t.Fail()
	}

}

func subsetIndexHelper(t *testing.T, result [][]int, expectedResult [][]int) {
	if len(result) != len(expectedResult) {
		t.Log("subset indexes returned wrong number of results for: ", result, " :", expectedResult)
		t.FailNow()
	}
	for i, item := range result {
		if len(item) != len(expectedResult[0]) {
			t.Log("subset indexes returned a result with wrong numbrer of items ", i, " : ", result, " : ", expectedResult)
			t.FailNow()
		}
		for j, value := range item {
			if value != expectedResult[i][j] {
				t.Log("Subset indexes had wrong number at ", i, ",", j, " : ", result, " : ", expectedResult)
				t.Fail()
			}
		}
	}
}

type solveTechniqueTestHelperOptions struct {
	transpose       bool
	targetCellsLen  int
	pointerCellsLen int
	targetNums      IntSlice
	targetSame      cellGroupType
	targetGroup     int
	description     string
}

func humanSolveTechniqueTestHelper(t *testing.T, puzzleName string, techniqueName string, options solveTechniqueTestHelperOptions) {
	//TODO: test for col and block as well
	grid := NewGrid()
	grid.LoadFromFile(puzzlePath(puzzleName))

	if options.transpose {
		grid = grid.transpose()
	}

	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	steps := solver.Find(grid)

	if len(steps) == 0 {
		t.Error(techniqueName, " didn't find a cell it should have.")
	}

	step := steps[0]

	//TODO: allow a way to pass in the exact cell addreses you expect to get.
	if len(step.TargetCells) != options.targetCellsLen {
		t.Error(techniqueName, " had the wrong number of target cells: ", len(step.TargetCells))
	}
	if len(step.PointerCells) != options.pointerCellsLen {
		t.Error(techniqueName, " had the wrong number of pointer cells: ", len(step.PointerCells))
		t.Fail()
	}

	switch options.targetSame {
	case GROUP_ROW:
		if !step.TargetCells.SameRow() || step.TargetCells.Row() != options.targetGroup {
			t.Error("The target cells in the ", techniqueName, " were wrong row :", step.TargetCells.Row())
		}
	case GROUP_BLOCK:
		if !step.TargetCells.SameBlock() || step.TargetCells.Block() != options.targetGroup {
			t.Error("The target cells in the ", techniqueName, " were wrong block :", step.TargetCells.Block())
		}
	case GROUP_NONE:
		//Do nothing
	default:
		t.Error("human solve technique helper error: unsupported group type: ", options.targetSame)
	}

	if options.targetNums != nil {
		if !step.TargetNums.SameContentAs(options.targetNums) {
			t.Error(techniqueName, " found the wrong numbers: ", step.TargetNums)
		}
	}

	if options.description != "" {
		description := solver.Description(step)
		if description != options.description {
			t.Error("Wrong description for ", techniqueName, ": ", description)
		}
	}

	//TODO: we should do exhaustive testing of SolveStep application. We used to test it here, but as long as targetCells and targetNums are correct it should be fine.

	grid.Done()
}
