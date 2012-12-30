package dokugen

import (
	"strings"
	"testing"
)

//This grid is #27 from Totally Pocket Sudoku. It's included for testing purposes only.

const TEST_GRID = `6|1|2|.|.|.|4|.|3
.|3|.|4|9|.|.|7|2
.|.|7|.|.|.|.|6|5
.|.|.|.|6|1|.|8|.
1|.|3|.|4|.|2|.|6
.|6|.|5|2|.|.|.|.
.|9|.|.|.|.|5|.|.
7|2|.|.|8|5|.|3|.
5|.|1|.|.|.|9|4|7`

const SOLVED_TEST_GRID = `6|1|2|7|5|8|4|9|3
8|3|5|4|9|6|1|7|2
9|4|7|2|1|3|8|6|5
2|5|9|3|6|1|7|8|4
1|7|3|8|4|9|2|5|6
4|6|8|5|2|7|3|1|9
3|9|6|1|7|4|5|2|8
7|2|4|9|8|5|6|3|1
5|8|1|6|3|2|9|4|7`

const TEST_GRID_DIAGRAM = `•••|•••|•••||   |   |   ||•••|   |•••
•6•|•1•|•2•||   | 5 |   ||•4•|   |•3•
•••|•••|•••||78 |7  |78 ||•••|  9|•••
---+---+---||---+---+---||---+---+---
   |•••|   ||•••|•••|   ||1  |•••|•••
   |•3•| 5 ||•4•|•9•|  6||   |•7•|•2•
 8 |•••| 8 ||•••|•••| 8 || 8 |•••|•••
---+---+---||---+---+---||---+---+---
   |   |•••||123|1 3| 23||1  |•••|•••
4  |4  |•7•||   |   |   ||   |•6•|•5•
 89| 8 |•••|| 8 |   | 8 || 8 |•••|•••
-----------++-----------++-----------
-----------++-----------++-----------
 2 |   |   ||  3|•••|•••||  3|•••|   
4  |45 |45 ||   |•6•|•1•||   |•8•|4  
  9|7  |  9||7 9|•••|•••||7  |•••|  9
---+---+---||---+---+---||---+---+---
•••|   |•••||   |•••|   ||•••|   |•••
•1•| 5 |•3•||   |•4•|   ||•2•| 5 |•6•
•••|78 |•••||789|•••|789||•••|  9|•••
---+---+---||---+---+---||---+---+---
   |•••|   ||•••|•••|  3||1 3|1  |1  
4  |•6•|4  ||•5•|•2•|   ||   |   |4  
 89|•••| 89||•••|•••|789||7  |  9|  9
-----------++-----------++-----------
-----------++-----------++-----------
  3|•••|   ||123|1 3| 23||•••|12 |1  
4  |•9•|4 6||  6|   |4 6||•5•|   |   
 8 |•••| 8 ||7  |7  |7  ||•••|   | 8 
---+---+---||---+---+---||---+---+---
•••|•••|   ||1  |•••|•••||1  |•••|1  
•7•|•2•|4 6||  6|•8•|•5•||  6|•3•|   
•••|•••|   ||  9|•••|•••||   |•••|   
---+---+---||---+---+---||---+---+---
•••|   |•••|| 23|  3| 23||•••|•••|•••
•5•|   |•1•||  6|   |  6||•9•|•4•|•7•
•••| 8 |•••||   |   |   ||•••|•••|•••`

//This grid is #102 from Totally Pocket Sudoku. It's included for testing purposes only.

const ADVANCED_TEST_GRID = `.|5|.|.|.|.|7|4|.
7|2|.|.|3|.|.|.|.
.|.|1|5|.|.|.|.|8
.|8|3|.|.|2|.|.|.
1|.|.|.|5|.|.|.|2
.|.|.|9|.|.|1|7|.
9|.|.|.|.|1|4|.|.
.|.|.|.|8|.|.|9|1
.|1|6|.|.|.|.|3|.`

const SOLVED_ADVANCED_TEST_GRID = `3|5|8|2|1|6|7|4|9
7|2|9|8|3|4|5|1|6
4|6|1|5|9|7|3|2|8
6|8|3|1|7|2|9|5|4
1|9|7|4|5|3|8|6|2
5|4|2|9|6|8|1|7|3
9|3|5|6|2|1|4|8|7
2|7|4|3|8|5|6|9|1
8|1|6|7|4|9|2|3|5`

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
	grid := NewGrid()
	grid.Load(data)
	if len(grid.cells) != DIM*DIM {
		t.Log("Didn't generate enough cells")
		t.Fail()
	}
	if grid.DataString() != data {
		t.Log("The grid round-tripped with different result than data in")
		t.Fail()
	}
	if grid.cells[10].Number() != 1 {
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

	cell := grid.Cell(2, 2)

	if cell.Row != 2 || cell.Col != 2 {
		t.Log("We grabbed a cell but what we got back was the wrong row and col.")
		t.Fail()
	}

	neighbors := cell.Neighbors()

	if len(neighbors) != NUM_NEIGHBORS {
		t.Log("We got a different number of neighbors than what we were expecting: ", len(neighbors))
		t.Fail()
	}
	neighborsMap := make(map[*Cell]bool)
	for _, neighbor := range neighbors {
		if neighbor == nil {
			t.Log("We found a nil neighbor")
			t.Fail()
		}
		if neighbor.Row != cell.Row && neighbor.Col != cell.Col && neighbor.Block != cell.Block {
			t.Log("We found a neighbor in ourselves that doesn't appear to be related: Neighbor: ", neighbor, " Cell: ", cell)
			t.Fail()
		}
		if _, ok := neighborsMap[neighbor]; ok {
			t.Log("We found a duplicate in the neighbors list")
			t.Fail()
		} else {
			neighborsMap[cell] = true
		}
	}
}

func TestGridLoad(t *testing.T) {
	grid := NewGrid()
	grid.Load(TEST_GRID)

	cell := grid.Cell(0, 0)

	if cell.Number() != 6 {
		t.Log("The loaded grid did not have a 6 in the upper left corner")
		t.Fail()
	}

	cell = grid.Cell(DIM-1, DIM-1)

	if cell.Number() != 7 {
		t.Log("The loaded grid did not have a 7 in the bottom right corner")
		t.Fail()
	}

	if grid.DataString() != TEST_GRID {
		t.Log("The real test grid did not survive a round trip via DataString: \n", grid.DataString(), "\n\n", TEST_GRID)
		t.Fail()
	}

	if grid.Diagram() != TEST_GRID_DIAGRAM {
		t.Log("The grid did not match the expected diagram: \n", grid.Diagram())
		t.Fail()
	}

	//Test copying.

	copy := grid.Copy()

	if grid.DataString() != copy.DataString() {
		t.Log("Copied grid does not have the same datastring!")
		t.Fail()
	}

	copy.Cell(0, 0).SetNumber(5)

	if copy.Cell(0, 0).Number() == grid.Cell(0, 0).Number() {
		t.Log("When we modified the copy's cell, it also affected the original.")
		t.Fail()
	}

	if grid.Solved() {
		t.Log("Grid reported it was solved when it was not.")
		t.Fail()
	}

	if grid.Invalid() {
		t.Log("Grid thought it was invalid when it wasn't: \n", grid.Diagram())
		t.Fail()
	}

	if num := grid.fillSimpleCells(); num != 45 {
		t.Log("We filled simple cells on the test grid but didn't get as many as we were expecting: ", num, "/", 45)
		t.Fail()
	}

	if grid.Invalid() {
		t.Log("fillSimpleCells filled in something that made the grid invalid: \n", grid.Diagram())
		t.Fail()
	}

	if !grid.Solved() {
		t.Log("Grid didn't think it was solved when it was.")
		t.Fail()
	}

	if grid.DataString() != SOLVED_TEST_GRID {
		t.Log("After filling simple cells, the grid was not actually solved correctly.")
		t.Fail()
	}

	cell.SetNumber(cell.Number() + 1)

	if !grid.Invalid() {
		t.Log("Grid didn't notice it was invalid when it actually was.")
		t.Fail()
	}

	cell.SetNumber(cell.Number() - 1)

	if grid.Invalid() {
		t.Log("Grid didn't noticed when it flipped from being invalid to being valid again.")
		t.Fail()
	}

	for i := 1; i <= DIM; i++ {
		cell.setImpossible(i)
	}

	if !grid.Invalid() {
		t.Log("Grid didn't notice when it became invalid because one of its cells has no more possibilities")
		t.Fail()
	}

}

func TestAdvancedSolve(t *testing.T) {
	grid := NewGrid()
	grid.Load(ADVANCED_TEST_GRID)

	if grid.DataString() != ADVANCED_TEST_GRID {
		t.Log("Advanced grid didn't survive a roundtrip to DataString")
		t.Fail()
	}

	copy := grid.Copy()

	copy.fillSimpleCells()

	if copy.Solved() {
		t.Log("Advanced grid was 'solved' with just fillSimpleCells")
		t.Fail()
	}

	solutions := grid.Solutions()

	if grid.DataString() != ADVANCED_TEST_GRID {
		t.Log("Calling Solutions() modified the original grid.")
		t.Fail()
	}

	if len(solutions) != 1 {
		t.Log("We found the wrong number of solutions in Advanced grid.")
		t.FailNow()
	}

	if solutions[0].DataString() != SOLVED_ADVANCED_TEST_GRID {
		t.Log("Solve found the wrong solution.")
		t.Fail()
	}

}

func nCopies(in string, copies int) (result []string) {
	for i := 0; i < copies; i++ {
		result = append(result, in)
	}
	return
}
