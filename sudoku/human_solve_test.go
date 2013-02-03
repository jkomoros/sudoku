package sudoku

import (
	"testing"
)

func TestSolveOnlyLegalNumber(t *testing.T) {
	grid := NewGrid()
	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)
	cell := grid.Cell(3, 3)
	num := cell.Number()
	cell.SetNumber(0)

	//Now that cell should be filled by this technique.

	solver := &onlyLegalNumberTechnique{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The only legal number technique did not solve a puzzle it should have.")
		t.FailNow()
	}
	if step.Col != 3 || step.Row != 3 {
		t.Log("The only legal number technique identified the wrong cell.")
		t.Fail()
	}
	if step.Num != num {
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

	solver := &necessaryInRowTechnique{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The necessary in row technique did not solve a puzzle it should have.")
		t.FailNow()
	}
	if step.Col != 3 || step.Row != 3 {
		t.Log("The necessary in row technique identified the wrong cell.")
		t.Fail()
	}
	if step.Num != DIM {
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

	solver := &necessaryInColTechnique{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The necessary in col technique did not solve a puzzle it should have.")
		t.FailNow()
	}
	if step.Col != 3 || step.Row != 3 {
		t.Log("The necessary in col technique identified the wrong cell.")
		t.Fail()
	}
	if step.Num != DIM {
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

	solver := &necessaryInBlockTechnique{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The necessary in block technique did not solve a puzzle it should have.")
		t.FailNow()
	}
	if step.Col != 3 || step.Row != 3 {
		t.Log("The necessary in block technique identified the wrong cell.")
		t.Fail()
	}
	if step.Num != DIM {
		t.Log("The necessary in block technique identified the wrong number.")
		t.Fail()
	}
	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in block technique did actually mutate the grid.")
		t.Fail()
	}
}
