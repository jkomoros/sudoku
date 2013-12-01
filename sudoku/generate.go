package sudoku

import (
	"math/rand"
)

//Fill will find a random filling of the puzzle that is valid. If it cannot find one,
// it will return False and leave the grid as it found it. Generally you would only want to call this on
//grids that have more than one solution (e.g. a fully blank grid)
func (self *Grid) Fill() bool {

	solutions := self.nOrFewerSolutions(1)

	if len(solutions) != 0 {
		self.Load(solutions[0].DataString())
		return true
	}

	return false
}

func GenerateGrid() *Grid {
	grid := NewGrid()
	//Do a random fill of the grid
	grid.Fill()

	keepGoing := true

	for keepGoing {
		//Unless we make a successful change this loop, don't bother continuing.
		keepGoing = false

		cells := make([]*Cell, len(grid.cells[:]))

		//TODO: remove cells to make the grid well-balanced.

		for i, j := range rand.Perm(len(grid.cells[:])) {
			cells[i] = &grid.cells[j]
		}

		for _, cell := range cells {
			num := cell.Number()
			if num == 0 {
				continue
			}
			//Unfill it.
			cell.SetNumber(0)
			if grid.HasMultipleSolutions() {
				//Put it back in.
				cell.SetNumber(num)
			} else {
				//we had a success! keep going around again.
				keepGoing = true
			}
		}
	}

	return grid
}
