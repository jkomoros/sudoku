package sudokustate

import (
	"github.com/jkomoros/sudoku"
	"io/ioutil"
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

	model.StartGroup("second group")

	model.SetNumber(sudoku.CellRef{0, 3}, 4)
	model.FinishGroupAndExecute()

	digestJson := model.Digest()

	//Uncomment to resave a new golden.
	//ioutil.WriteFile("test/golden.json", digestJson, 0644)

	if digestJson == nil {
		t.Error("Got nil digest for legitimate digest")
	}

	golden, err := ioutil.ReadFile("test/golden.json")

	if err != nil {
		t.Fatal("Couldn't load golden file at golden.json", err)
	}

	if string(digestJson) != string(golden) {
		t.Error("Got incorrect golden json. Got", digestJson, "wanted", golden)
	}

}
