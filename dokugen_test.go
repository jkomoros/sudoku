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
	col := grid.Col(2)
	items := col.All().Now()
	if num := len(items); num != DIM {
		t.Log("We got back a column but it had the wrong amount of items: ", num, "\n", items)
		t.Fail()
	}
	row := grid.Row(2)
	if len(row.All().Now()) != DIM {
		t.Log("We got back a row but it had the wrong number of items.")
		t.Fail()
	}

	cellStream := row.All()
	_ = cellStream.Now()
	if len(cellStream.Now()) != 0 {
		t.Log("We called Now again on a cellstream and got something back!")
		t.Fail()
	}

}

func nCopies(in string, copies int) (result []string) {
	for i := 0; i < copies; i++ {
		result = append(result, in)
	}
	return
}
