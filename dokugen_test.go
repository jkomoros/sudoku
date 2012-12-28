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

	for count := 0; count < DIM; count++ {
		col := grid.Col(count)
		if num := len(col); num != DIM {
			t.Log("We got back a column but it had the wrong amount of items: ", num, "\n")
			t.Fail()
		}
		for i, cell := range col {
			if cell.Col != count {
				t.Log("One of the cells we got back when asking for column ", count, " was not in the right column.")
				t.Fail()
			}
			if cell.Row != i {
				t.Log("One of the cells we got back when asking for column ", count, " was not in the right row.")
				t.Fail()
			}
		}

		row := grid.Row(count)
		if len(row) != DIM {
			t.Log("We got back a row but it had the wrong number of items.")
			t.Fail()
		}
		for i, cell := range row {
			if cell.Row != count {
				t.Log("One of the cells we got back when asking for row ", count, " was not in the right rows.")
				t.Fail()
			}
			if cell.Col != i {
				t.Log("One of the cells we got back from row ", count, " was not in the right column.")
				t.Fail()
			}
		}

	}

	count := 2

	//TODO: move this inside of the loop.
	block := grid.Block(count)
	if len(block) != DIM {
		t.Log("We got back a block but it had the wrong number of items.")
		t.Fail()
	}

	if block[0].Row != 0 || block[0].Col != 6 {
		t.Log("We got back the wrong first cell from block ", count, ": ", block[0])
		t.Fail()
	}

	if block[DIM-1].Row != 2 || block[DIM-1].Col != 8 {
		t.Log("We got back the wrong last cell from block ", count, ": ", block[0])
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
