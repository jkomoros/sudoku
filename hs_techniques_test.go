package sudoku

import (
	"log"
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

type solveTechniqueMatchMode int

const (
	solveTechniqueMatchModeAll solveTechniqueMatchMode = iota
	solveTechniqueMatchModeAny
)

type solveTechniqueTestHelperOptions struct {
	transpose bool
	//Whether the descriptions of cells are a list of legal possible individual values, or must all match.
	matchMode    solveTechniqueMatchMode
	targetCells  []cellRef
	pointerCells []cellRef
	targetNums   IntSlice
	pointerNums  IntSlice
	targetSame   cellGroupType
	targetGroup  int
	//If description provided, the description MUST match.
	description string
	//If descriptions provided, ONE of the descriptions must match.
	//generally used in conjunction with solveTechniqueMatchModeAny.
	descriptions []string
	debugPrint   bool
}

func humanSolveTechniqueTestHelper(t *testing.T, puzzleName string, techniqueName string, options solveTechniqueTestHelperOptions) {
	//TODO: test for col and block as well
	grid := NewGrid()
	if !grid.LoadFromFile(puzzlePath(puzzleName)) {
		t.Fatal("Couldn't load puzzle ", puzzleName)
	}

	if options.transpose {
		grid = grid.transpose()
	}

	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	steps := solver.Find(grid)

	if len(steps) == 0 {
		t.Fatal(techniqueName, " didn't find a cell it should have.")
	}

	step := steps[0]

	if options.debugPrint {
		log.Println(step)
	}

	if options.matchMode == solveTechniqueMatchModeAll {

		//All must match

		if options.targetCells != nil {
			if !step.TargetCells.sameAsRefs(options.targetCells) {
				t.Error(techniqueName, " had the wrong target cells: ", step.TargetCells)
			}
		}
		if options.pointerCells != nil {
			if !step.PointerCells.sameAsRefs(options.pointerCells) {
				t.Error(techniqueName, " had the wrong pointer cells: ", step.PointerCells)
			}
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
		case GROUP_COL:
			if !step.TargetCells.SameCol() || step.TargetCells.Col() != options.targetGroup {
				t.Error("The target cells in the ", techniqueName, " were wrong col :", step.TargetCells.Col())
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

		if options.pointerNums != nil {
			if !step.PointerNums.SameContentAs(options.pointerNums) {
				t.Error(techniqueName, "found the wrong numbers:", step.PointerNums)
			}
		}
	} else if options.matchMode == solveTechniqueMatchModeAny {

		foundMatch := false

		if options.targetCells != nil {
			foundMatch = false
			for _, ref := range options.targetCells {
				for _, cell := range step.TargetCells {
					if ref.Cell(grid) == cell {
						//TODO: break out early
						foundMatch = true
					}
				}
			}
			if !foundMatch {
				t.Error(techniqueName, " had the wrong target cells: ", step.TargetCells)
			}
		}
		if options.pointerCells != nil {
			t.Error("Pointer cells in match mode any not yet supported.")
		}

		if options.targetSame != GROUP_NONE {
			t.Error("Target Same in match mode any not yet supported.")
		}

		if options.targetNums != nil {
			foundMatch = false
			for _, targetNum := range options.targetNums {
				for _, num := range step.TargetNums {
					if targetNum == num {
						foundMatch = true
						//TODO: break early here.
					}
				}
			}
			if !foundMatch {
				t.Error(techniqueName, " had the wrong target nums: ", step.TargetNums)
			}
		}

		if options.pointerNums != nil {
			foundMatch = false
			for _, pointerNum := range options.pointerNums {
				for _, num := range step.PointerNums {
					if pointerNum == num {
						foundMatch = true
						//TODO: break early here
					}
				}
			}
			if !foundMatch {
				t.Error(techniqueName, " had the wrong pointer nums: ", step.PointerNums)
			}
		}
	}

	if options.description != "" {
		//Normalize the step so that the description will be stable for the test.
		step.normalize()
		description := solver.Description(step)
		if description != options.description {
			t.Error("Wrong description for ", techniqueName, ". Got:*", description, "* expected: *", options.description, "*")
		}
	} else if options.descriptions != nil {
		foundMatch := false
		step.normalize()
		description := solver.Description(step)
		for _, targetDescription := range options.descriptions {
			if description == targetDescription {
				foundMatch = true
			}
		}
		if !foundMatch {
			t.Error("No descriptions matched for ", techniqueName, ". Got:*", description)
		}
	}

	//TODO: we should do exhaustive testing of SolveStep application. We used to test it here, but as long as targetCells and targetNums are correct it should be fine.

	grid.Done()
}
