package dokugen

import (
	"strings"
)

const DIM = 9
const ROW_SEP = "||"
const COL_SEP = "|"

type Grid struct {
	cells [DIM * DIM]Cell
}

func NewGrid(data string) *Grid {
	result := &Grid{}
	i := 0
	for r, row := range strings.Split(data, ROW_SEP) {
		for c, cell := range strings.Split(row, COL_SEP) {
			result.cells[i] = NewCell(result, r, c, cell)
		}
	}
	return result
}
