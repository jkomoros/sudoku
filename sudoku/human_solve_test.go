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

	step := solver.Apply(grid)

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
	if !grid.Solved() {
		t.Log("The only legal number technique did not actually mutate the grid.")
		t.Fail()
	}
}
