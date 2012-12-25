package dokugen

import (
	"strconv"
)

type Cell struct {
	grid   *Grid
	Number int
	Row    int
	Col    int
	cells  *simpleCellList
}

func NewCell(grid *Grid, row int, col int, data string) Cell {
	//Format, for now, is just the number itself, or 0 if no number.
	num, _ := strconv.Atoi(data)
	return Cell{grid, num, row, col, nil}
}

func (self *Cell) String() string {
	return strconv.Itoa(self.Number)
}
