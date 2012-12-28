package dokugen

import (
	"log"
	"strconv"
)

type Cell struct {
	grid      *Grid
	Number    int
	Row       int
	Col       int
	neighbors []*Cell
}

func NewCell(grid *Grid, row int, col int, data string) Cell {
	//Format, for now, is just the number itself, or 0 if no number.
	num, _ := strconv.Atoi(data)
	return Cell{grid, num, row, col, nil}
}

func (self *Cell) Neighbors() []*Cell {
	if !self.grid.initalized {
		log.Println("Neighbors called before the grid had been finalized")
		return nil
	}
	if self.neighbors == nil {
		//We'll have DIM -1 neighbors for row, col, block
		self.neighbors = make([]*Cell, (DIM-1)*3)
		outputIndex := 0
		for _, cell := range self.grid.Row(self.Row) {
			self.neighbors[outputIndex] = cell
			outputIndex++
		}
		for _, cell := range self.grid.Col(self.Col) {
			self.neighbors[outputIndex] = cell
			outputIndex++
		}
		//TODO: also put in each block neighbor.
	}
	return self.neighbors

}

func (self *Cell) DataString() string {
	return strconv.Itoa(self.Number)
}

func (self *Cell) String() string {
	return "Cell[" + strconv.Itoa(self.Row) + "][" + strconv.Itoa(self.Col) + "]:" + strconv.Itoa(self.Number) + "\n"
}
