package dokugen

import (
	"strings"
)

const DIM = 9
const ROW_SEP = "||"
const COL_SEP = "|"

type Grid struct {
	cells [DIM * DIM]Cell
	rows  [DIM]*simpleCellList
	cols  [DIM]*simpleCellList
}

func NewGrid(data string) *Grid {
	result := &Grid{}
	i := 0
	for r, row := range strings.Split(data, ROW_SEP) {
		for c, cell := range strings.Split(row, COL_SEP) {
			result.cells[i] = NewCell(result, r, c, cell)
			i++
		}
	}
	return result
}

func (self *Grid) Row(index int) CellList {
	if self.rows[index] == nil {
		self.rows[index] = &simpleCellList{self, index * DIM, index*DIM + DIM, 0, nil}
	}
	return self.rows[index]
}

func (self *Grid) Col(index int) CellList {
	if self.cols[index] == nil {
		self.cols[index] = &simpleCellList{self, index, DIM*(DIM-1) + index + 1, DIM, nil}
	}
	return self.cols[index]
}

func (self *Grid) DataString() string {
	var rows []string
	for r := 0; r < DIM; r++ {
		var row []string
		for c := 0; c < DIM; c++ {
			row = append(row, self.cells[r*DIM+c].DataString())
		}
		rows = append(rows, strings.Join(row, COL_SEP))
	}
	return strings.Join(rows, ROW_SEP)
}
