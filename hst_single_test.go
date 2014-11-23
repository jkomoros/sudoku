package sudoku

import (
	"testing"
)

func TestObviousInCollectionRow(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells: []cellRef{{2, 3}},
		targetSame:  GROUP_ROW,
		targetGroup: 2,
		targetNums:  IntSlice([]int{7}),
		description: "(2,3) is the only cell in row 2 that is unfilled, and it must be 7",
	}
	humanSolveTechniqueTestHelper(t, "obviousrow.sdk", "Obvious In Row", options)

}

func TestObviousInCollectionCol(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		transpose:   true,
		targetCells: []cellRef{{3, 2}},
		targetSame:  GROUP_COL,
		targetGroup: 2,
		targetNums:  IntSlice([]int{7}),
		description: "(3,2) is the only cell in column 2 that is unfilled, and it must be 7",
	}
	humanSolveTechniqueTestHelper(t, "obviousrow.sdk", "Obvious In Col", options)

}

func TestObviousInCollectionBlock(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells: []cellRef{{4, 1}},
		targetSame:  GROUP_BLOCK,
		targetGroup: 3,
		targetNums:  IntSlice([]int{9}),
		description: "(4,1) is the only cell in block 3 that is unfilled, and it must be 9",
	}
	humanSolveTechniqueTestHelper(t, "obviousblock.sdk", "Obvious In Block", options)

}

func TestSolveOnlyLegalNumber(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)
	cell := grid.Cell(3, 3)
	num := cell.Number()
	cell.SetNumber(0)

	//Now that cell should be filled by this technique.

	techniqueName := "Only Legal Number"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	steps := solver.Find(grid)

	if len(steps) == 0 {
		t.Log("The only legal number technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	step := steps[0]

	description := solver.Description(step)
	if description != "3 is the only remaining valid number for that cell" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The only legal number technique identified the wrong cell.")
		t.Fail()
	}
	numFromStep := step.TargetNums[0]

	if numFromStep != num {
		t.Log("The only legal number technique identified the wrong number.")
		t.Fail()
	}
	if grid.Solved() {
		t.Log("The only legal number technique did actually mutate the grid.")
		t.Fail()
	}
}

//TODO: use the test solve helper func for these three tests.
func TestNecessaryInRow(t *testing.T) {
	grid := NewGrid()

	//We DON'T call grid.done because we will have poked some unrealistic values into the cells.

	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)

	//We're going to cheat an set up an unrealistic grid.

	impossibles := make([]int, DIM)

	for i := 0; i < DIM-1; i++ {
		impossibles[i] = 0
	}
	impossibles[DIM-1] = 1

	//SetNumber will affect the other cells in row, so do it first.
	for _, cell := range grid.Row(3) {
		cell.number = 0
		copy(cell.impossibles[:], impossibles)
	}

	cell := grid.Cell(3, 3)
	//This is the only cell where DIM will be allowed.
	cell.impossibles[DIM-1] = 0

	//Now that cell should be filled by this technique.

	techniqueName := "Necessary In Row"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	steps := solver.Find(grid)

	if len(steps) == 0 {
		t.Log("The necessary in row technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	step := steps[0]

	description := solver.Description(step)
	if description != "9 is required in the 3 row, and 3 is the only column it fits" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The necessary in row technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.TargetNums[0]

	if numFromStep != DIM {
		t.Log("The necessary in row technique identified the wrong number.")
		t.Fail()
	}
	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in row technique did actually mutate the grid.")
		t.Fail()
	}
}

func TestNecessaryInCol(t *testing.T) {
	grid := NewGrid()

	//We DON'T call grid.done because we will have poked some unrealistic values into the cells.

	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)

	//We're going to cheat an set up an unrealistic grid.

	impossibles := make([]int, DIM)

	for i := 0; i < DIM-1; i++ {
		impossibles[i] = 0
	}
	impossibles[DIM-1] = 1

	//SetNumber will affect the other cells in row, so do it first.
	for _, cell := range grid.Col(3) {
		cell.number = 0
		copy(cell.impossibles[:], impossibles)
	}

	cell := grid.Cell(3, 3)
	//This is the only cell where DIM will be allowed.
	cell.impossibles[DIM-1] = 0

	//Now that cell should be filled by this technique.

	techniqueName := "Necessary In Col"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	steps := solver.Find(grid)

	if len(steps) == 0 {
		t.Log("The necessary in col technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	step := steps[0]

	description := solver.Description(step)
	if description != "9 is required in the 3 column, and 3 is the only row it fits" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The necessary in col technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.TargetNums[0]

	if numFromStep != DIM {
		t.Log("The necessary in col technique identified the wrong number.")
		t.Fail()
	}
	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in col technique did actually mutate the grid.")
		t.Fail()
	}
}

func TestNecessaryInBlock(t *testing.T) {
	grid := NewGrid()

	//We DON'T call grid.done because we will have poked some unrealistic values into the cells.

	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)

	//We're going to cheat an set up an unrealistic grid.

	impossibles := make([]int, DIM)

	for i := 0; i < DIM-1; i++ {
		impossibles[i] = 0
	}
	impossibles[DIM-1] = 1

	//SetNumber will affect the other cells in row, so do it first.
	for _, cell := range grid.Block(4) {
		cell.number = 0
		copy(cell.impossibles[:], impossibles)
	}

	cell := grid.Cell(3, 3)
	//This is the only cell where DIM will be allowed.
	cell.impossibles[DIM-1] = 0

	//Now that cell should be filled by this technique.

	techniqueName := "Necessary In Block"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	steps := solver.Find(grid)

	if len(steps) == 0 {
		t.Log("The necessary in block technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	step := steps[0]

	description := solver.Description(step)
	if description != "9 is required in the 4 block, and (3,3) is the only cell it fits" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The necessary in block technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.TargetNums[0]

	if numFromStep != DIM {
		t.Log("The necessary in block technique identified the wrong number.")
		t.Fail()
	}
	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in block technique did actually mutate the grid.")
		t.Fail()
	}

}
