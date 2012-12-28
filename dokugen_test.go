package dokugen

import (
	"strings"
	"testing"
)

func TestCellCreation(t *testing.T) {
	data := "1"
	cell := NewCell(nil, 0, 0, data)
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
	if cell.DataString() != data {
		t.Log("Cell round-tripped out with different string than data in")
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
	if grid.DataString() != data {
		t.Log("The grid round-tripped with different result than data in")
		t.Fail()
	}
	if grid.cells[10].Number != 1 {
		t.Log("A random spot check of a cell had the wrong number: %s", grid.cells[10])
		t.Fail()
	}
	//TODO: test that these are actually getting back the right cells...
	col := grid.Col(2)
	if num := len(col); num != DIM {
		t.Log("We got back a column but it had the wrong amount of items: ", num, "\n")
		t.Fail()
	}
	row := grid.Row(2)
	if len(row) != DIM {
		t.Log("We got back a row but it had the wrong number of items.")
		t.Fail()
	}

	block := grid.Block(2)
	if len(block) != DIM {
		t.Log("We got back a block but it had the wrong number of items.")
		t.Fail()
	}

	cell := grid.Cell(2, 2)

	if cell.Row != 2 || cell.Col != 2 {
		t.Log("We grabbed a cell but what we got back was the wrong row and col.")
		t.Fail()
	}

}

func nCopies(in string, copies int) (result []string) {
	for i := 0; i < copies; i++ {
		result = append(result, in)
	}
	return
}
