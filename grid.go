package dokugen

import (
	"strings"
)

const DIM = 9
const ROW_SEP = "||"
const COL_SEP = "|"

type Grid struct {
	cells [DIM * DIM]Cell
	rows  [DIM][]*Cell
	cols  [DIM][]*Cell
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

func (self *Grid) Row(index int) []*Cell {
	if self.rows[index] == nil {
		self.rows[index] = self.cellList(index*DIM, index*DIM+DIM, 1)
	}
	return self.rows[index]
}

func (self *Grid) Col(index int) []*Cell {
	if self.cols[index] == nil {
		self.cols[index] = self.cellList(index, DIM*(DIM-1)+index+1, DIM)
	}
	return self.cols[index]
}

func (self *Grid) Cell(row int, col int) *Cell {
	index := row*DIM + col
	if index >= DIM*DIM || index < 0 {
		return nil
	}
	return &self.cells[index]
}

func (self *Grid) cellList(start int, end int, stride int) []*Cell {
	var result [DIM]*Cell
	outputIndex := 0
	i := start
	if stride == 0 {
		stride = 1
	}
	for i < end {
		result[outputIndex] = &self.cells[i]
		i = i + stride
		outputIndex++
	}
	return result[:]
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
