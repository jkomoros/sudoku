package dokugen

import (
	"strconv"
)

type Cell struct {
	grid   *Grid
	Number int
	Row    int
	Col    int
	cells  []*Cell
}

func NewCell(grid *Grid, row int, col int, data string) Cell {
	//Format, for now, is just the number itself, or 0 if no number.
	num, _ := strconv.Atoi(data)
	return Cell{grid, num, row, col, nil}
}

func (self *Cell) DataString() string {
	return strconv.Itoa(self.Number)
}

func (self *Cell) String() string {
	return "Cell[" + strconv.Itoa(self.Row) + "][" + strconv.Itoa(self.Col) + "]:" + strconv.Itoa(self.Number) + "\n"
}
