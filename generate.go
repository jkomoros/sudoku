package sudoku

import (
	"github.com/davecgh/go-spew/spew"
	"log"
	"math/rand"
)

//GenerationOptions provides configuration options for generating a sudoku puzzle.
type GenerationOptions struct {
	//symmetrty and symmetryType control the aesthetics of the generated grid. symmetryPercentage
	//controls roughly what percentage of cells with have a filled partner across the provided plane of
	//symmetry.
	Symmetry           SymmetryType
	SymmetryPercentage float64
	//The minimum number of cells to leave filled in the puzzle. The generated puzzle might have
	//more filled cells. A value of DIM * DIM - 1, for example, would return an extremely trivial
	//puzzle.
	MinFilledCells int
}

//TODO: consider providing a GenerationOptions.Default(), just like HumanSolveOptions does.
var defaultGenerationOptions GenerationOptions

func init() {
	defaultGenerationOptions = GenerationOptions{
		Symmetry:           SYMMETRY_VERTICAL,
		SymmetryPercentage: 0.7,
		MinFilledCells:     0,
	}
}

//Fill will find a random filling of the puzzle such that every cell is filled and no cells conflict with their neighbors. If it cannot find one,
// it will return false and leave the grid as it found it. Generally you would only want to call this on
//grids that have more than one solution (e.g. a fully blank grid). Fill provides a good starting point for generated puzzles.
func (self *Grid) Fill() bool {

	solutions := self.nOrFewerSolutions(1)

	if len(solutions) != 0 {
		self.Load(solutions[0].DataString())
		return true
	}

	return false
}

//GenerateGrid returns a new sudoku puzzle with a single unique solution and many of its cells unfilled--a
//puzzle that is appropriate (and hopefully fun) for humans to solve. GenerateGrid first finds a random
//full filling of the grid, then iteratively removes cells until just before the grid begins having
//multiple solutions. The result is a grid that has a single valid solution but many of its cells
//unfilled. Pass nil for options to use reasonable defaults.
//GenerateGrid doesn't currently give any way to define the desired difficulty; the best option is to
//repeatedly generate puzzles until you find one that matches your desired difficulty. cmd/dokugen
//applies this technique.
func GenerateGrid(options *GenerationOptions) *Grid {

	if options == nil {
		options = &defaultGenerationOptions
	}

	grid := NewGrid()
	//Do a random fill of the grid
	grid.Fill()

	log.Println("Original grid\n", grid)

	//Make a copy so we don't mutate the passed in dict
	symmetryPercentage := options.SymmetryPercentage

	//Make sure symmetry percentage is within the legal range.
	if symmetryPercentage < 0.0 {
		symmetryPercentage = 0.0
	}
	if symmetryPercentage > 1.0 {
		symmetryPercentage = 1.0
	}

	cells := make([]*Cell, len(grid.cells[:]))

	for i, j := range rand.Perm(len(grid.cells[:])) {
		cells[i] = &grid.cells[j]
	}

	type generateRecord struct {
		i                   int
		mainCell            *Cell
		otherCell           *Cell
		num                 int
		otherNum            int
		otherCellWasVacated bool
		didFill             bool
		mainCellWasEmpty    bool
		gridState           string
	}

	var records []generateRecord

	for i, cell := range cells {

		if grid.HasMultipleSolutions() {
			log.Println("On cell", i, "we had already gotten multiple solutions for grid", grid)
			spew.Dump(records)
			return grid
		}

		num := cell.Number()
		if num == 0 {
			records = append(records, generateRecord{i, cell, nil, 0, 0, false, false, true, grid.DataString()})
			continue
		}

		var otherNum int
		var otherCell *Cell
		var otherCellWasVacated bool
		didFill := true

		if rand.Float64() < symmetryPercentage {

			//Pick a symmetrical partner for symmetryPercentage number of cells.
			otherCell = cell.SymmetricalPartner(options.Symmetry)

			if otherCell != nil {
				if otherCell.Number() == 0 {
					//We must have already un-filled it as a primary cell.
					//If we were to unfill this, we could get in a weird state where
					//we get multiple solutions without noticing (which caused bug #134).
					//So pretend like we didn't draw one.
					otherCell = nil
					otherCellWasVacated = true
				} else {
					otherNum = otherCell.Number()
				}
			}
		}

		numCellsToFillThisStep := 1
		if otherCell != nil {
			numCellsToFillThisStep = 2
		}

		if grid.numFilledCells-numCellsToFillThisStep < options.MinFilledCells {
			//Doing this step would leave us with too few cells filled. Finish.
			continue
		}

		//Unfill it.
		cell.SetNumber(0)
		if otherCell != nil {
			otherCell.SetNumber(0)
		}

		if grid.HasMultipleSolutions() {
			//Put it back in.
			didFill = false
			cell.SetNumber(num)
			if otherCell != nil {
				otherCell.SetNumber(otherNum)
			}
		}

		records = append(records, generateRecord{i, cell, otherCell, num, otherNum, otherCellWasVacated, didFill, false, grid.DataString()})
	}

	return grid
}
