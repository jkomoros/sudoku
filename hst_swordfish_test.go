package sudoku

import (
	"testing"
)

//TODO: test a few more puzzles to make sure I'm exercising it correctly.

func swordfishExampleGrid(t *testing.T) *Grid {
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

	return grid
}

func TestSwordfishCol(t *testing.T) {

	techniqueVariantsTestHelper(t, "Swordfish Col")

	grid := swordfishExampleGrid(t)

	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{1, 1}, {5, 4}},
		pointerCells: []cellRef{{1, 0}, {1, 5}, {5, 3}, {5, 5}, {8, 0}, {8, 3}},
		targetNums:   IntSlice{1},
		//TODO: test description
	}
	options.stepsToCheck.grid = grid

	//TODO: it's not possible to just pass in an override grid to humanSolveTechniqueTestHelper as
	//is, because we're overloading passing it to stepsToCheck. That's a smell.
	grid, solver, steps := humanSolveTechniqueTestHelperStepGenerator(t, "NOOP", "Swordfish Col", options)

	options.stepsToCheck.grid = grid
	options.stepsToCheck.solver = solver
	options.stepsToCheck.steps = steps

	humanSolveTechniqueTestHelper(t, "NOOP", "Swordfish Col", options)

}

func TestSwordfishRow(t *testing.T) {

	techniqueVariantsTestHelper(t, "Swordfish Row")

	grid := swordfishExampleGrid(t)
	grid = grid.transpose()

	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{1, 1}, {4, 5}},
		pointerCells: []cellRef{{0, 1}, {5, 1}, {3, 5}, {5, 5}, {0, 8}, {3, 8}},
		targetNums:   IntSlice{1},
		//TODO: test description
	}
	options.stepsToCheck.grid = grid

	//TODO: it's not possible to just pass in an override grid to humanSolveTechniqueTestHelper as
	//is, because we're overloading passing it to stepsToCheck. That's a smell.
	grid, solver, steps := humanSolveTechniqueTestHelperStepGenerator(t, "NOOP", "Swordfish Row", options)

	options.stepsToCheck.grid = grid
	options.stepsToCheck.solver = solver
	options.stepsToCheck.steps = steps

	humanSolveTechniqueTestHelper(t, "NOOP", "Swordfish Row", options)

}