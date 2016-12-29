package sudokustate

import (
	"encoding/json"
	"github.com/jkomoros/sudoku"
	"io/ioutil"
	"reflect"
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

	digest := model.Digest()

	//Uncomment to resave a new golden.
	outputGolden := false

	if outputGolden {
		jsonOutput, err := json.MarshalIndent(digest, "", "  ")
		if err != nil {
			t.Fatal("Golden JSON wasn't output", err)
		}
		ioutil.WriteFile("test/golden.json", jsonOutput, 0644)
		t.Fatal("Wrote out a new golden")
	}

	golden, err := ioutil.ReadFile("test/golden.json")

	if err != nil {
		t.Fatal("Couldn't load golden file at golden.json", err)
	}

	var goldenDigest Digest

	if err := json.Unmarshal(golden, &goldenDigest); err != nil {
		t.Fatal("Couldn't unmarshall golden file", err)
	}

	if !reflect.DeepEqual(goldenDigest, digest) {
		t.Error("Got incorrect golden json. Got", digest, "wanted", goldenDigest)
	}

}
