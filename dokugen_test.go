package dokugen

import (
	"strings"
	"testing"
)

func TestCellCreation(t *testing.T) {
	cell := NewCell(nil, 0, 0, "1")
	if cell.Number != 1 {
		t.Log("Number came back wrong")
		t.Fail()
	}
	if cell.Row != 0 {
		t.Log("Row came back wrong")
		t.Fail()
	}
	if cell.Col != 0 {
		t.Log("Cell came back wrong")
		t.Fail()
	}
	//TODO: test failing for values that are too high.
}

func TestGridCreation(t *testing.T) {
	cellData := "1"
	rowData := strings.Join(nCopies(cellData, DIM), COL_SEP)
	data := strings.Join(nCopies(rowData, DIM), ROW_SEP)
	grid := NewGrid(data)
	if len(grid.cells) != DIM*DIM {
		t.Log("Didn't generate enough cells")
		t.Fail()
	}
	if grid.cells[10].Number != 1 {
		t.Log("A random spot check of a cell had the wrong number: %s", grid.cells[10])
		t.Fail()
	}
}

func nCopies(in string, copies int) (result []string) {
	for i := 0; i < copies; i++ {
		result = append(result, in)
	}
	return
}
