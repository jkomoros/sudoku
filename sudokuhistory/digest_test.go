package sudokuhistory

import (
	"encoding/json"
	"github.com/jkomoros/sudoku"
	"os"
	"reflect"
	"testing"
	"time"
)

func defaultDigest() *Digest {
	model := &Model{}
	model.SetGrid(sudoku.NewGrid())

	model.SetMarks(sudoku.CellRef{Row: 3, Col: 4}, map[int]bool{
		3: true,
		4: true,
	})

	model.SetNumber(sudoku.CellRef{Row: 0, Col: 0}, 3)

	model.StartGroup("test group")

	model.SetNumber(sudoku.CellRef{Row: 0, Col: 1}, 4)
	model.SetMarks(sudoku.CellRef{Row: 0, Col: 2}, map[int]bool{
		3: true,
		4: true,
	})
	model.FinishGroupAndExecute()

	model.StartGroup("second group")

	model.SetNumber(sudoku.CellRef{Row: 0, Col: 3}, 4)
	model.FinishGroupAndExecute()

	//This move undoes the earlier one and make sure that the digest handles
	//setting a cell to zero appropriately.
	model.SetNumber(sudoku.CellRef{Row: 0, Col: 0}, 0)

	result := model.Digest()
	return &result
}

func TestDigestValid(t *testing.T) {
	digest := defaultDigest()

	if err := digest.valid(); err != nil {
		t.Error("Got an error for a valid digest:", err)
	}

	oldPuzzle := digest.Puzzle

	digest.Puzzle = "INVAILD_PUZZLE"

	if err := digest.valid(); err == nil {
		t.Error("Didn't notice that a digest with an invalid puzzle was invalid")
	}

	digest.Puzzle = ""

	if err := digest.valid(); err == nil {
		t.Error("Didn't notice that a non existent puzzle was invalid")
	}

	digest.Puzzle = oldPuzzle

	moveGroup := &digest.MoveGroups[3]

	oldTime := moveGroup.TimeOffset

	moveGroup.TimeOffset = 10

	if err := digest.valid(); err == nil {
		t.Error("Didn't notice that a move with an invalid time")
	}

	moveGroup.TimeOffset = oldTime

	move := &moveGroup.Moves[0]

	oldNumber := move.Number

	move.Number = nil

	if err := digest.valid(); err == nil {
		t.Error("Didn't notice that a move with no marks or numbers was invalid")
	}

	move.Number = oldNumber

	move.Marks = map[int]bool{
		3: true,
	}

	if err := digest.valid(); err == nil {
		t.Error("Didn't notice that a move with both a marks and a move was invalid", move)
	}

	moveGroup.Moves = make([]MoveDigest, 0)

	if err := digest.valid(); err == nil {
		t.Error("Didn't notice that a move group had no moves")
	}

}

func TestDigest(t *testing.T) {

	digest := defaultDigest()

	//Time will be set to a time that isn't the same in golden. So check for
	//it to be reasonable now, then reset to a specific number so it compares
	//to Golden ok.

	var lastTime time.Duration

	for i, _ := range digest.MoveGroups {
		moveGroup := &digest.MoveGroups[i]
		if moveGroup.TimeOffset < lastTime {
			t.Error("The timestamp was smaller than a previous time stamp at movegroup", i)
		}
		lastTime = moveGroup.TimeOffset
		//Set the time to an arbitraty but consitent inceasing value for
		//golden comparison
		moveGroup.TimeOffset = time.Duration(100 + (i * 17))
	}

	//Uncomment to resave a new golden.
	outputGolden := false

	if outputGolden {
		jsonOutput, err := json.MarshalIndent(digest, "", "  ")
		if err != nil {
			t.Fatal("Golden JSON wasn't output", err)
		}
		os.WriteFile("test/golden.json", jsonOutput, 0644)
		t.Fatal("Wrote out a new golden")
	}

	golden, err := os.ReadFile("test/golden.json")

	if err != nil {
		t.Fatal("Couldn't load golden file at golden.json", err)
	}

	var goldenDigest Digest

	if err := json.Unmarshal(golden, &goldenDigest); err != nil {
		t.Fatal("Couldn't unmarshall golden file", err)
	}

	if !reflect.DeepEqual(goldenDigest, *digest) {
		t.Error("Got incorrect golden json. Got", digest, "wanted", goldenDigest)
	}

	newModel := &Model{}

	if err := newModel.LoadDigest(*digest); err != nil {
		t.Error("Load digest returned an error")
	}

	newDigest := newModel.Digest()

	if !reflect.DeepEqual(*digest, newDigest) {
		t.Error("Loading up a digest didn't set the state correctly")
	}

}
