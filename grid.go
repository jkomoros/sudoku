package dokugen

import (
	"log"
	"strings"
)

//TODO: Support non-squared DIMS (logic in Block() would need updating)
const BLOCK_DIM = 3
const DIM = BLOCK_DIM * BLOCK_DIM
const ROW_SEP = "\n"
const COL_SEP = "|"
const ALT_COL_SEP = "||"

type Grid struct {
	initalized bool
	cells      [DIM * DIM]Cell
	rows       [DIM][]*Cell
	cols       [DIM][]*Cell
	blocks     [DIM][]*Cell
	queue      *FiniteQueue
}

func NewGrid() *Grid {
	result := &Grid{}
	result.queue = NewFiniteQueue(1, DIM)
	i := 0
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			result.cells[i] = NewCell(result, r, c)
			//The cell can't insert itself because it doesn't know where it will actually live in memory yet.
			result.queue.Insert(&result.cells[i])
			i++
		}
	}
	result.initalized = true
	return result
}

func LoadGrid(data string) *Grid {
	result := NewGrid()
	data = strings.Replace(data, ALT_COL_SEP, COL_SEP, -1)
	for r, row := range strings.Split(data, ROW_SEP) {
		for c, data := range strings.Split(row, COL_SEP) {
			cell := result.Cell(r, c)
			cell.Load(data)
		}
	}
	return result
}

func (self *Grid) Row(index int) []*Cell {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Row: ", index)
		return nil
	}
	if self.rows[index] == nil {
		self.rows[index] = self.cellList(index, 0, index, DIM-1)
	}
	return self.rows[index]
}

func (self *Grid) Col(index int) []*Cell {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Col: ", index)
		return nil
	}
	if self.cols[index] == nil {
		self.cols[index] = self.cellList(0, index, DIM-1, index)
	}
	return self.cols[index]
}

func (self *Grid) Block(index int) []*Cell {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Block: ", index)
		return nil
	}
	if self.blocks[index] == nil {
		//Conceptually, we'll pretend like the grid is made up of blocks that are arrayed with row/column
		//Once we find the block r/c, we'll multiply by the actual dim to get the upper left corner.

		blockCol := index % BLOCK_DIM
		blockRow := (index - blockCol) / BLOCK_DIM

		col := blockCol * BLOCK_DIM
		row := blockRow * BLOCK_DIM

		self.blocks[index] = self.cellList(row, col, row+BLOCK_DIM-1, col+BLOCK_DIM-1)
	}
	return self.blocks[index]
}

func (self *Grid) blockForCell(row int, col int) int {
	blockCol := col / BLOCK_DIM
	blockRow := row / BLOCK_DIM
	return blockRow*BLOCK_DIM + blockCol
}

func (self *Grid) Cell(row int, col int) *Cell {
	index := row*DIM + col
	if index >= DIM*DIM || index < 0 {
		log.Println("Invalid row/col index passed to Cell: ", row, ", ", col)
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
		result[i] = self.Cell(currentRow, currentCol)
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

//Fills in all of the cells it can without branching or doing any advanced
//techniques that require anything more than a single cell's possibles list.
func (self *Grid) fillSimpleCells() int {
	count := 0
	obj := self.queue.GetSmallerThan(2)
	for obj != nil {
		cell, ok := obj.(*Cell)
		if !ok {
			continue
		}
		cell.SetNumber(cell.implicitNumber())
		count++
		obj = self.queue.GetSmallerThan(2)
	}
	return count
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

func (self *Grid) String() string {
	var rows []string
	for r := 0; r < DIM; r++ {
		var row []string
		for c := 0; c < DIM; c++ {
			row = append(row, self.cells[r*DIM+c].String())
		}
		rows = append(rows, strings.Join(row, COL_SEP))
	}
	return strings.Join(rows, ROW_SEP)
}
