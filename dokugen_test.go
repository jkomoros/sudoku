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

	blockUpperLeftRow := make([]int, DIM)
	blockUpperLeftCol := make([]int, DIM)

	row := 0
	col := -1 * BLOCK_DIM

	for i := 0; i < DIM; i++ {
		col += BLOCK_DIM
		if col >= DIM {
			col = 0
			row += BLOCK_DIM
		}
		blockUpperLeftRow[i] = row
		blockUpperLeftCol[i] = col
	}

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

		block := grid.Block(count)
		if len(block) != DIM {
			t.Log("We got back a block but it had the wrong number of items.")
			t.Fail()
		}

		for _, cell := range block {
			if cell.Block != count {
				t.Log("We got a cell back in a block with the wrong block number")
				t.Fail()
			}
		}

		if block[0].Row != blockUpperLeftRow[count] || block[0].Col != blockUpperLeftCol[count] {
			t.Log("We got back the wrong first cell from block ", count, ": ", block[0])
			t.Fail()
		}

		if block[DIM-1].Row != blockUpperLeftRow[count]+BLOCK_DIM-1 || block[DIM-1].Col != blockUpperLeftCol[count]+BLOCK_DIM-1 {
			t.Log("We got back the wrong last cell from block ", count, ": ", block[0])
			t.Fail()
		}

	}

	//TODO: test neighbors list is set correctly.

	cell := grid.Cell(2, 2)

	if cell.Row != 2 || cell.Col != 2 {
		t.Log("We grabbed a cell but what we got back was the wrong row and col.")
		t.Fail()
	}

	neighbors := cell.Neighbors()

	if len(neighbors) != (DIM-1)*3-(BLOCK_DIM-1)*2 {
		t.Log("We got a different number of neighbors than what we were expecting: ", len(neighbors))
		t.Fail()
	}
	for _, neighbor := range neighbors {
		if neighbor == nil {
			t.Log("We found a nil neighbor")
			t.Fail()
		}
		if neighbor.Row != cell.Row && neighbor.Col != cell.Col && neighbor.Block != cell.Block {
			t.Log("We found a neighbor in ourselves that doesn't appear to be related: Neighbor: ", neighbor, " Cell: ", cell)
			t.Fail()
		}
		//TODO: Check to make sure we don't get duplicates in this list.
	}

}

func nCopies(in string, copies int) (result []string) {
	for i := 0; i < copies; i++ {
		result = append(result, in)
	}
	return
}
