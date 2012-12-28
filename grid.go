package dokugen

import (
	"strings"
)

//TODO: Support non-squared DIMS (logic in Block() would need updating)
const BLOCK_DIM = 3
const DIM = BLOCK_DIM * BLOCK_DIM
const ROW_SEP = "||"
const COL_SEP = "|"

type Grid struct {
	cells  [DIM * DIM]Cell
	rows   [DIM][]*Cell
	cols   [DIM][]*Cell
	blocks [DIM][]*Cell
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
		self.rows[index] = self.cellList(index, 0, index, DIM-1)
	}
	return self.rows[index]
}

func (self *Grid) Col(index int) []*Cell {
	if self.cols[index] == nil {
		self.cols[index] = self.cellList(0, index, DIM-1, index)
	}
	return self.cols[index]
}

func (self *Grid) Block(index int) []*Cell {
	if self.blocks[index] == nil {
		//Conceptually, we'll pretend like the grid is made up of blocks that are arrayed with row/column
		//Once we find the block r/c, we'll multiply by the actual dim to get the upper left corner.

		blockCol := index % BLOCK_DIM
		blockRow := index - blockCol

		col := blockCol * BLOCK_DIM
		row := blockRow * BLOCK_DIM

		self.blocks[index] = self.cellList(row, col, row+BLOCK_DIM-1, col+BLOCK_DIM-1)
	}
	return self.blocks[index]
}

func (self *Grid) Cell(row int, col int) *Cell {
	index := row*DIM + col
	if index >= DIM*DIM || index < 0 {
		return nil
	}
	return &self.cells[index]
}

func (self *Grid) cellList(rowOne int, colOne int, rowTwo int, colTwo int) []*Cell {
	length := (rowTwo - rowOne + 1) * (colTwo - colOne + 1)
	result := make([]*Cell, length)
	currentRow := rowOne
	currentCol := colOne
	for i := 0; i < length; i++ {
		result[i] = &self.cells[currentRow*DIM+currentCol]
		if colTwo > currentCol {
			currentCol++
		} else {
			if rowTwo > currentRow {
				currentRow++
				currentCol = colOne
			} else {
				//This should only happen the last time through the loop.
			}
		}
	}
	return result
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
