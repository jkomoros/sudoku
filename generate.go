package dokugen

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

	keepGoing := 50

	//TODO: have a more robust exit criteria.
	for keepGoing > 0 {
		cell := grid.Cell(rand.Intn(DIM), rand.Intn(DIM))
		num := cell.Number()
		if num == 0 {
			continue
		}
		//Unfill it.
		cell.SetNumber(0)
		if grid.HasMultipleSolutions() {
			//Put it back in.
			cell.SetNumber(num)
			keepGoing--
		}
	}

	return grid
}
