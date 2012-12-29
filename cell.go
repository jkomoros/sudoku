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
	Block     int
	neighbors []*Cell
}

func NewCell(grid *Grid, row int, col int, data string) Cell {
	//Format, for now, is just the number itself, or 0 if no number.
	num, _ := strconv.Atoi(data)
	return Cell{grid, num, row, col, grid.blockForCell(row, col), nil}
}

func (self *Cell) Neighbors() []*Cell {
	if !self.grid.initalized {
		log.Println("Neighbors called before the grid had been finalized")
		return nil
	}
	if self.neighbors == nil {
		neighborsMap := make(map[*Cell]bool)
		for _, cell := range self.grid.Row(self.Row) {
			if cell == self {
				continue
			}
			neighborsMap[cell] = true
		}
		for _, cell := range self.grid.Col(self.Col) {
			if cell == self {
				continue
			}
			neighborsMap[cell] = true
		}
		for _, cell := range self.grid.Block(self.Block) {
			if cell == self {
				continue
			}
			neighborsMap[cell] = true
		}
		self.neighbors = make([]*Cell, len(neighborsMap))
		i := 0
		for cell, _ := range neighborsMap {
			self.neighbors[i] = cell
			i++
		}
	}
	return self.neighbors

}

func (self *Cell) DataString() string {
	return strconv.Itoa(self.Number)
}

func (self *Cell) String() string {
	return "Cell[" + strconv.Itoa(self.Row) + "][" + strconv.Itoa(self.Col) + "]:" + strconv.Itoa(self.Number) + "\n"
}
