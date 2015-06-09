package sudoku

import (
	"testing"
)

//TODO: test a few more puzzles to make sure I'm exercising it correctly.

func TestSwordfishCol(t *testing.T) {

	techniqueVariantsTestHelper(t, "Swordfish")

	grid := NewGrid()

	puzzleName := "swordfish_example.sdk"

	if !grid.LoadFromFile(puzzlePath(puzzleName)) {
		t.Fatal("Couldn't load puzzle ", puzzleName)
	}

	//Set up the grid correctly for the Swordfish technique to work. The
	//example we use is a grid that has other work done to exclude
	//possibilities from certain cells.

	//TODO: it's a smell that there's no way to serialize and load up a grid
	//with extra excludes set.
	excludedConfig := map[cellRef]IntSlice{
		cellRef{0, 0}: IntSlice{1, 8},
		cellRef{1, 3}: IntSlice{1},
		cellRef{1, 4}: IntSlice{1, 8},
		cellRef{2, 3}: IntSlice{1},
		cellRef{2, 5}: IntSlice{1, 8},
		cellRef{3, 0}: IntSlice{2, 8},
		cellRef{4, 0}: IntSlice{7},
		cellRef{4, 1}: IntSlice{7},
		cellRef{7, 3}: IntSlice{1, 6},
		cellRef{7, 5}: IntSlice{1},
	}

	for ref, ints := range excludedConfig {
		cell := ref.Cell(grid)
		for _, exclude := range ints {
			cell.SetExcluded(exclude, true)
		}
	}

	options := solveTechniqueTestHelperOptions{}
	options.stepsToCheck.grid = grid

	_, _, _ = humanSolveTechniqueTestHelperStepGenerator(t, "NOOP", "Swordfish", options)
}

//TODO: TestSwordfishRow (and implement Row!)
