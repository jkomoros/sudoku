package sudoku

import (
	"testing"
)

const POINTING_PAIR_COL_GRID = `3|.|6|.|.|.|.|.|.
.|.|.|.|7|.|.|.|.
4|.|5|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.`

func TestSolveOnlyLegalNumber(t *testing.T) {
	grid := NewGrid()
	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)
	cell := grid.Cell(3, 3)
	num := cell.Number()
	cell.SetNumber(0)

	//Now that cell should be filled by this technique.

	solver := &nakedSingleTechnique{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The only legal number technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The only legal number technique identified the wrong cell.")
		t.Fail()
	}
	numFromStep := step.Nums[0]

	if numFromStep != num {
		t.Log("The only legal number technique identified the wrong number.")
		t.Fail()
	}
	if grid.Solved() {
		t.Log("The only legal number technique did actually mutate the grid.")
		t.Fail()
	}
}

func TestNecessaryInRow(t *testing.T) {
	grid := NewGrid()
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

	solver := &hiddenSingleInRow{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The necessary in row technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The necessary in row technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.Nums[0]

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

	solver := &hiddenSingleInCol{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The necessary in col technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The necessary in col technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.Nums[0]

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

	solver := &hiddenSingleInBlock{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The necessary in block technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The necessary in block technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.Nums[0]

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

func TestPointingPairCol(t *testing.T) {
	grid := NewGrid()
	grid.Load(POINTING_PAIR_COL_GRID)
	solver := &pointingPairCol{}
	step := solver.Find(grid)
	if step == nil {
		t.Log("The pointing pair col didn't find a cell it should have")
		t.Fail()
	}
	if len(step.TargetCells) != BLOCK_DIM*2 {
		t.Log("The pointing pair col gave back the wrong number of target cells")
		t.Fail()
	}
	if len(step.PointerCells) != BLOCK_DIM-1 {
		t.Log("The pointing pair col gave back the wrong number of pointer cells")
		t.Fail()
	}
	if !step.TargetCells.SameCol() || step.TargetCells.Col() != 1 {
		t.Log("The target cells in the pointing pair col technique were wrong col")
		t.Fail()
	}
	if len(step.Nums) != 1 || step.Nums[0] != 7 {
		t.Log("Pointing pair col technique gave the wrong number")
		t.Fail()
	}
}

func TestHumanSolve(t *testing.T) {
	grid := NewGrid()
	grid.Load(TEST_GRID)
	steps := grid.HumanSolve()
	//TODO: test to make sure that we use a wealth of different techniques. This will require a cooked random for testing.
	if steps == nil {
		t.Log("Human solve returned 0 techniques")
		t.Fail()
	}
	if !grid.Solved() {
		t.Log("Human solve failed to solve the simple grid.")
		t.Fail()
	}
}
