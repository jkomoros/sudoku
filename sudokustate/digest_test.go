package sudokustate

import (
	"github.com/jkomoros/sudoku"
	"testing"
)

func TestDigest(t *testing.T) {
	model := &Model{}
	model.SetGrid(sudoku.NewGrid())

	model.SetMarks(sudoku.CellRef{3, 4}, map[int]bool{
		3: true,
		4: true,
	})

	model.SetNumber(sudoku.CellRef{0, 0}, 3)

	model.StartGroup("test group")

	model.SetNumber(sudoku.CellRef{0, 1}, 4)
	model.SetMarks(sudoku.CellRef{0, 2}, map[int]bool{
		3: true,
		4: true,
	})
	model.FinishGroupAndExecute()

	digest := model.Digest()

	if digest == nil {
		t.Error("Got nil digest for legitimate digest")
	}

	//TODO: actually test the result against a golden.
}
