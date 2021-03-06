package sudoku

import (
	"container/list"
	"io/ioutil"
	"math"
	"runtime"
	"strings"
	"testing"
)

//This grid is #27 from Totally Pocket Sudoku. It's included for testing purposes only.

const PUZZLES_DIRECTORY = "puzzles"

const TEST_GRID = `6|1|2|.|.|.|4|.|3
.|3|.|4|9|.|.|7|2
.|.|7|.|.|.|.|6|5
.|.|.|.|6|1|.|8|.
1|.|3|.|4|.|2|.|6
.|6|.|5|2|.|.|.|.
.|9|.|.|.|.|5|.|.
7|2|.|.|8|5|.|3|.
5|.|1|.|.|.|9|4|7`

const TRANSPOSED_TEST_GRID = `6|.|.|.|1|.|.|7|5
1|3|.|.|.|6|9|2|.
2|.|7|.|3|.|.|.|1
.|4|.|.|.|5|.|.|.
.|9|.|6|4|2|.|8|.
.|.|.|1|.|.|.|5|.
4|.|.|.|2|.|5|.|9
.|7|6|8|.|.|.|3|4
3|2|5|.|6|.|.|.|7`

const SOLVED_TEST_GRID = `6|1|2|7|5|8|4|9|3
8|3|5|4|9|6|1|7|2
9|4|7|2|1|3|8|6|5
2|5|9|3|6|1|7|8|4
1|7|3|8|4|9|2|5|6
4|6|8|5|2|7|3|1|9
3|9|6|1|7|4|5|2|8
7|2|4|9|8|5|6|3|1
5|8|1|6|3|2|9|4|7`

const TEST_GRID_DIAGRAM = `XXX|XXX|XXX||   |   |   ||XXX|   |XXX
X6X|X1X|X2X||   | 5 |   ||X4X|   |X3X
XXX|XXX|XXX||78 |7  |78 ||XXX|  9|XXX
---+---+---||---+---+---||---+---+---
   |XXX|   ||XXX|XXX|   ||1  |XXX|XXX
   |X3X| 5 ||X4X|X9X|  6||   |X7X|X2X
 8 |XXX| 8 ||XXX|XXX| 8 || 8 |XXX|XXX
---+---+---||---+---+---||---+---+---
   |   |XXX||123|1 3| 23||1  |XXX|XXX
4  |4  |X7X||   |   |   ||   |X6X|X5X
 89| 8 |XXX|| 8 |   | 8 || 8 |XXX|XXX
-----------++-----------++-----------
-----------++-----------++-----------
 2 |   |   ||  3|XXX|XXX||  3|XXX|   
4  |45 |45 ||   |X6X|X1X||   |X8X|4  
  9|7  |  9||7 9|XXX|XXX||7  |XXX|  9
---+---+---||---+---+---||---+---+---
XXX|   |XXX||   |XXX|   ||XXX|   |XXX
X1X| 5 |X3X||   |X4X|   ||X2X| 5 |X6X
XXX|78 |XXX||789|XXX|789||XXX|  9|XXX
---+---+---||---+---+---||---+---+---
   |XXX|   ||XXX|XXX|  3||1 3|1  |1  
4  |X6X|4  ||X5X|X2X|   ||   |   |4  
 89|XXX| 89||XXX|XXX|789||7  |  9|  9
-----------++-----------++-----------
-----------++-----------++-----------
  3|XXX|   ||123|1 3| 23||XXX|12 |1  
4  |X9X|4 6||  6|   |4 6||X5X|   |   
 8 |XXX| 8 ||7  |7  |7  ||XXX|   | 8 
---+---+---||---+---+---||---+---+---
XXX|XXX|   ||1  |XXX|XXX||1  |XXX|1  
X7X|X2X|4 6||  6|X8X|X5X||  6|X3X|   
XXX|XXX|   ||  9|XXX|XXX||   |XXX|   
---+---+---||---+---+---||---+---+---
XXX|   |XXX|| 23|  3| 23||XXX|XXX|XXX
X5X|   |X1X||  6|   |  6||X9X|X4X|X7X
XXX| 8 |XXX||   |   |   ||XXX|XXX|XXX`

const TEST_GRID_EXCLUDED_DIAGRAM = `XXX|XXX|XXX||   |   |   ||XXX|   |XXX
X6X|X1X|X2X||   | 5 |   ||X4X|   |X3X
XXX|XXX|XXX||78 |7  |78 ||XXX|  9|XXX
---+---+---||---+---+---||---+---+---
   |XXX|   ||XXX|XXX|   ||1  |XXX|XXX
   |X3X| 5 ||X4X|X9X|  6||   |X7X|X2X
 8 |XXX| 8 ||XXX|XXX| 8 || 8 |XXX|XXX
---+---+---||---+---+---||---+---+---
   |   |XXX||123|1 3| 23||1  |XXX|XXX
   |4  |X7X||   |   |   ||   |X6X|X5X
 89| 8 |XXX|| 8 |   | 8 || 8 |XXX|XXX
-----------++-----------++-----------
-----------++-----------++-----------
 2 |   |   ||  3|XXX|XXX||  3|XXX|   
4  |45 |45 ||   |X6X|X1X||   |X8X|4  
  9|7  |  9||7 9|XXX|XXX||7  |XXX|  9
---+---+---||---+---+---||---+---+---
XXX|   |XXX||   |XXX|   ||XXX|   |XXX
X1X| 5 |X3X||   |X4X|   ||X2X| 5 |X6X
XXX|78 |XXX||789|XXX|789||XXX|  9|XXX
---+---+---||---+---+---||---+---+---
   |XXX|   ||XXX|XXX|  3||1 3|1  |1  
4  |X6X|4  ||X5X|X2X|   ||   |   |4  
 89|XXX| 89||XXX|XXX|789||7  |  9|  9
-----------++-----------++-----------
-----------++-----------++-----------
  3|XXX|   ||123|1 3| 23||XXX|12 |1  
4  |X9X|4 6||  6|   |4 6||X5X|   |   
 8 |XXX| 8 ||7  |7  |7  ||XXX|   | 8 
---+---+---||---+---+---||---+---+---
XXX|XXX|   ||1  |XXX|XXX||1  |XXX|1  
X7X|X2X|4 6||  6|X8X|X5X||  6|X3X|   
XXX|XXX|   ||  9|XXX|XXX||   |XXX|   
---+---+---||---+---+---||---+---+---
XXX|   |XXX|| 23|  3| 23||XXX|XXX|XXX
X5X|   |X1X||  6|   |  6||X9X|X4X|X7X
XXX| 8 |XXX||   |   |   ||XXX|XXX|XXX`

//This grid is #102 from Totally Pocket Sudoku. It's included for testing purposes only.

//This is also in a costant because we want to just string compare it to the loaded file.
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

func init() {
	runtime.GOMAXPROCS(_NUM_SOLVER_THREADS)
}

//returns true if the provided grid is a gridImpl
func isGridImpl(grid Grid) bool {
	switch grid.(type) {
	default:
		return false
	case *gridImpl:
		return true
	}
}

//returns true if the provided grid is a mutableGridImpl
func isMutableGridImpl(grid Grid) bool {
	switch grid.(type) {
	default:
		return false
	case *mutableGridImpl:
		return true
	}
}

func BenchmarkGridCopy(b *testing.B) {
	grid := LoadSDK(ADVANCED_TEST_GRID)
	for i := 0; i < b.N; i++ {
		_ = grid.Copy()
	}
}

func TestGridCopy(t *testing.T) {

	grid := MutableLoadSDK(ADVANCED_TEST_GRID)

	cell := grid.MutableCell(0, 0)
	cell.SetMark(3, true)
	cell.SetMark(4, true)

	cell = grid.MutableCell(0, 2)
	cell.SetExcluded(3, true)
	cell.SetExcluded(4, true)

	gridCopy := grid.Copy()

	if !isGridImpl(gridCopy) {
		t.Error("Expected grid.copy() to return a gridImpl but it didn't")
	}

	if grid.Diagram(true) != gridCopy.Diagram(true) {
		t.Error("Grid and copy don't match in marks. Got", gridCopy.Diagram(true), "wanted", grid.Diagram(true))
	}

	if grid.Diagram(false) != gridCopy.Diagram(false) {
		t.Error("Grid and copy don't match in terms of excludes. Got", gridCopy.Diagram(false), "wanted", grid.Diagram(false))
	}

	//Make sure that the grid that was returned from copy does not change when the grid it was derived from is modified.

	grid.MutableCell(1, 2).SetNumber(5)

	if grid.Diagram(true) == gridCopy.Diagram(true) {
		t.Error("A read-only grid copy changed when the grid it was created from was mutated. Got", gridCopy.Diagram(true), "from both")
	}
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
	grid := MutableLoadSDK(data)

	if !isMutableGridImpl(grid) {
		t.Error("Expected grid to be mutable under the covers but it wasn't")
	}

	//We want to run all of the same tests on a Grid and MutableGrid.
	roGrid := grid.Copy()

	if !isGridImpl(roGrid) {
		t.Error("Expected ROgrid to be immutable under the covers but it wasn't")
	}

	grids := []struct {
		grid        Grid
		description string
	}{
		{
			grid:        grid,
			description: "Mutable grid",
		},
		{
			grid:        roGrid,
			description: "Read Only Grid",
		},
	}
	for _, config := range grids {

		grid := config.grid
		description := config.description

		if len(grid.Cells()) != DIM*DIM {
			t.Error(description, "Didn't generate enough cells")
		}
		if grid.DataString() != data {
			t.Error(description, "The grid round-tripped with different result than data in")
		}
		if grid.Cells()[10].Number() != 1 {
			t.Error(description, "A random spot check of a cell had the wrong number: %s", grid.Cells()[10])
		}

		if grid.numFilledCells() != DIM*DIM {
			t.Error(description, "We didn't think all cells were filled, but they were!")
		}

		for count := 0; count < DIM; count++ {
			col := grid.Col(count)
			if num := len(col); num != DIM {
				t.Error(description, "We got back a column but it had the wrong amount of items: ", num, "\n")
			}
			for i, cell := range col {
				if cell.Col() != count {
					t.Error(description, "One of the cells we got back when asking for column ", count, " was not in the right column.")
				}
				if cell.Row() != i {
					t.Error(description, "One of the cells we got back when asking for column ", count, " was not in the right row.")
				}
			}

			row := grid.Row(count)
			if len(row) != DIM {
				t.Error(description, "We got back a row but it had the wrong number of items.")
			}
			for i, cell := range row {
				if cell.Row() != count {
					t.Error(description, "One of the cells we got back when asking for row ", count, " was not in the right rows.")
				}
				if cell.Col() != i {
					t.Error(description, "One of the cells we got back from row ", count, " was not in the right column.")
				}
			}

			block := grid.Block(count)
			if len(block) != DIM {
				t.Error(description, "We got back a block but it had the wrong number of items.")
			}

			for _, cell := range block {
				if cell.Block() != count {
					t.Error(description, "We got a cell back in a block with the wrong block number")
				}
			}

			if block[0].Row() != blockUpperLeftRow[count] || block[0].Col() != blockUpperLeftCol[count] {
				t.Error(description, "We got back the wrong first cell from block ", count, ": ", block[0])
			}

			if block[DIM-1].Row() != blockUpperLeftRow[count]+BLOCK_DIM-1 || block[DIM-1].Col() != blockUpperLeftCol[count]+BLOCK_DIM-1 {
				t.Error(description, "We got back the wrong last cell from block ", count, ": ", block[0])
			}

		}

		cell := grid.Cell(2, 2)

		if cell.Row() != 2 || cell.Col() != 2 {
			t.Error(description, "We grabbed a cell but what we got back was the wrong row and col.")
		}

		neighbors := cell.Neighbors()

		if len(neighbors) != _NUM_NEIGHBORS {
			t.Error(description, "We got a different number of neighbors than what we were expecting: ", len(neighbors))
		}
		neighborsMap := make(map[Cell]bool)
		for _, neighbor := range neighbors {
			if neighbor == nil {
				t.Error(description, "We found a nil neighbor")
			}
			if neighbor.Row() != cell.Row() && neighbor.Col() != cell.Col() && neighbor.Block() != cell.Block() {
				t.Error(description, "We found a neighbor in ourselves that doesn't appear to be related: Neighbor: ", neighbor, " Cell: ", cell)
			}
			if _, ok := neighborsMap[neighbor]; ok {
				t.Error(description, "We found a duplicate in the neighbors list")
			} else {
				neighborsMap[cell] = true
			}
		}
	}

}

func TestGridCells(t *testing.T) {

	grid := MutableLoadSDK(TEST_GRID)

	//Test the mutableGridImpl

	if !isMutableGridImpl(grid) {
		t.Error("Expected to get a mutableGridimpl from LoadSDK")
	}

	cells := grid.MutableCells()

	if len(cells) != DIM*DIM {
		t.Fatal("Grid.cells gave back a cellslice with wrong number of cells", len(cells))
	}

	//Make sure it's the same cells
	for i, cell := range cells {
		if cell.Grid() != grid {
			t.Error("cell #", i, "had wrong grid")
		}
		if cell != grid.MutableCell(cell.Row(), cell.Col()) {
			t.Error("cell #", i, "was not the same as the right one in the grid")
		}
	}

	//Now do the same test for a gridImpl

	roGrid := grid.Copy()

	if !isGridImpl(roGrid) {
		t.Error("Expected to get a gridImpl from grid.Copy")
	}

	roCells := roGrid.Cells()

	if len(roCells) != DIM*DIM {
		t.Fatal("Grid.cells gave back a cellslice with wrong number of cells", len(roCells))
	}

	//Make sure it's the same cells
	for i, cell := range roCells {
		if cell.Grid() != roGrid {
			t.Error("cell #", i, "had wrong grid")
		}
		if cell != roGrid.Cell(cell.Row(), cell.Col()) {
			t.Error("cell #", i, "was not the same as the right one in the grid")
		}
	}

	//make sure mutating a cell from the celllist mutates the grid.
	cell := cells[3]

	if cell.Number() != 0 {
		t.Fatal("We expected cell #3 to be empty, but had", cell.Number())
	}
	cell.SetNumber(3)

	if grid.Cell(0, 3).Number() != 3 {
		t.Error("Mutating a cellin cells did not mutate the grid.")
	}
}

func TestMutableGridLoad(t *testing.T) {

	//Substantially recreated in TestGridLoad

	grid := MutableLoadSDK(TEST_GRID)

	if !isMutableGridImpl(grid) {
		t.Fatal("Expected grid from LoadSDK to be mutable")
	}

	cell := grid.MutableCell(0, 0)

	if cell.Number() != 6 {
		t.Error("The loaded grid did not have a 6 in the upper left corner")
	}

	cell = grid.MutableCell(DIM-1, DIM-1)

	if cell.Number() != 7 {
		t.Error("The loaded grid did not have a 7 in the bottom right corner")
	}

	if grid.DataString() != TEST_GRID {
		t.Error("The real test grid did not survive a round trip via DataString: \n", grid.DataString(), "\n\n", TEST_GRID)
	}

	if grid.Diagram(false) != TEST_GRID_DIAGRAM {
		t.Error("The grid did not match the expected diagram: \n", grid.Diagram(false))
	}

	//Twiddle an exclude to make sure it copies over correctly.
	grid.MutableCell(2, 0).SetExcluded(4, true)

	if grid.Diagram(false) != TEST_GRID_EXCLUDED_DIAGRAM {
		t.Error("Diagram did not reflect the manually excluded item: \n", grid.Diagram(false))
	}

	//Test copying.

	copy := grid.MutableCopy()

	if grid.DataString() != copy.DataString() {
		t.Error("Copied grid does not have the same datastring!")
	}

	for c, cell := range grid.MutableCells() {
		copyCell := cell.MutableInGrid(copy)
		cellI := cell.(*mutableCellImpl)
		copyCellI := copyCell.(*mutableCellImpl)
		if !IntSlice(cellI.impossibles[:]).SameAs(IntSlice(copyCellI.impossibles[:])) {
			t.Error("Cells at position", c, "had different impossibles:\n", cellI.impossibles, "\n", copyCellI.impossibles)
		}
		for i := 1; i <= DIM; i++ {
			if cell.Possible(i) != copyCell.Possible(i) {
				t.Error("The copy of the grid did not have the same possible at cell ", c, " i ", i)
			}
		}
	}

	copy.MutableCell(0, 0).SetNumber(5)

	if copy.Cell(0, 0).Number() == grid.Cell(0, 0).Number() {
		t.Error("When we modified the copy's cell, it also affected the original.")
	}

	if grid.Solved() {
		t.Error("Grid reported it was solved when it was not.")
	}

	if grid.Invalid() {
		t.Error("Grid thought it was invalid when it wasn't: \n", grid.Diagram(false))
	}

	previousRank := grid.rank()

	grid = withSimpleCellsFilled(grid).MutableCopy()
	cell = cell.MutableInGrid(grid)

	if num := previousRank - grid.rank(); num != 45 {
		t.Error("We filled simple cells on the test grid but didn't get as many as we were expecting: ", num, "/", 45)
	}

	if grid.Invalid() {
		t.Error("fillSimpleCells filled in something that made the grid invalid: \n", grid.Diagram(false))
	}

	if !grid.Solved() {
		t.Error("Grid didn't think it was solved when it was.")
	}

	if grid.DataString() != SOLVED_TEST_GRID {
		t.Error("After filling simple cells, the grid was not actually solved correctly.")
	}

	cell.SetNumber(cell.Number() + 1)

	if !grid.Invalid() {
		t.Error("Grid didn't notice it was invalid when it actually was.", grid)
	}

	cell.SetNumber(cell.Number() - 1)

	if grid.Invalid() {
		t.Error("Grid didn't noticed when it flipped from being invalid to being valid again.")
	}

	for i := 1; i <= DIM; i++ {
		cell.setImpossible(i)
	}

	if !grid.Invalid() {
		t.Error("Grid didn't notice when it became invalid because one of its cells has no more possibilities")
	}

}

func TestGridLoad(t *testing.T) {

	//Substantially recreated in TestMutableGridLoad

	mutableGrid := LoadSDK(TEST_GRID)

	grid := mutableGrid.Copy()

	if !isGridImpl(grid) {
		t.Fatal("Expected grid from mutable copy to be immutable")
	}

	cell := grid.Cell(0, 0)

	if cell.Number() != 6 {
		t.Error("The loaded grid did not have a 6 in the upper left corner")
	}

	cell = grid.Cell(DIM-1, DIM-1)

	if cell.Number() != 7 {
		t.Error("The loaded grid did not have a 7 in the bottom right corner")
	}

	if grid.DataString() != TEST_GRID {
		t.Error("The real test grid did not survive a round trip via DataString: \n", grid.DataString(), "\n\n", TEST_GRID)
	}

	if grid.Diagram(false) != TEST_GRID_DIAGRAM {
		t.Error("The grid did not match the expected diagram: \n", grid.Diagram(false))
	}

	//Test copying.

	copy := grid.Copy()

	if grid.DataString() != copy.DataString() {
		t.Error("Copied grid does not have the same datastring!")
	}

	for c, cell := range grid.Cells() {
		copyCell := cell.InGrid(copy)
		cellI := cell.(*cellImpl)
		copyCellI := copyCell.(*cellImpl)
		if !IntSlice(cellI.impossibles[:]).SameAs(IntSlice(copyCellI.impossibles[:])) {
			t.Error("Cells at position", c, "had different impossibles:\n", cellI.impossibles, "\n", copyCellI.impossibles)
		}
		for i := 1; i <= DIM; i++ {
			if cell.Possible(i) != copyCell.Possible(i) {
				t.Error("The copy of the grid did not have the same possible at cell ", c, " i ", i)
			}
		}
	}

	if grid.Solved() {
		t.Error("Grid reported it was solved when it was not.")
	}

	if grid.Invalid() {
		t.Error("Grid thought it was invalid when it wasn't: \n", grid.Diagram(false))
	}

	previousRank := grid.rank()

	grid = withSimpleCellsFilled(grid)
	cell = cell.InGrid(grid)

	if num := previousRank - grid.rank(); num != 45 {
		t.Error("We filled simple cells on the test grid but didn't get as many as we were expecting: ", num, "/", 45)
	}

	if grid.Invalid() {
		t.Error("fillSimpleCells filled in something that made the grid invalid: \n", grid.Diagram(false))
	}

	if !grid.Solved() {
		t.Error("Grid didn't think it was solved when it was.")
	}

	if grid.DataString() != SOLVED_TEST_GRID {
		t.Error("After filling simple cells, the grid was not actually solved correctly.")
	}

	grid = grid.CopyWithModifications(GridModification{
		&CellModification{
			Cell:   cell.Reference(),
			Number: cell.Number() + 1,
		},
	})

	if !grid.Invalid() {
		t.Error("Grid didn't notice it was invalid when it actually was.", grid)
	}

	grid = grid.CopyWithModifications(GridModification{
		&CellModification{
			Cell:   cell.Reference(),
			Number: cell.Number(),
		},
	})

	if grid.Invalid() {
		t.Error("Grid didn't noticed when it flipped from being invalid to being valid again.")
	}

	grid = grid.CopyWithModifications(GridModification{
		&CellModification{
			Cell:   cell.Reference(),
			Number: 0,
		},
	})

	excludes := map[int]bool{}

	for i := 1; i <= DIM; i++ {
		excludes[i] = true
	}

	grid = grid.CopyWithModifications(GridModification{
		&CellModification{
			Cell:            cell.Reference(),
			ExcludesChanges: excludes,
		},
	})

	if !grid.Invalid() {
		t.Error("Grid didn't notice when it became invalid because one of its cells has no more possibilities", grid.Diagram(false))
	}

}

func TestAdvancedSolve(t *testing.T) {

	grid, _ := MutableLoadSDKFromFile(puzzlePath("advancedtestgrid.sdk"))

	if !isMutableGridImpl(grid) {
		t.Fatal("Expected load sdk from file to return mutable grid")
	}

	roGrid := grid.Copy()

	if !isGridImpl(roGrid) {
		t.Fatal("Expected copy to return roGrid")
	}

	//Run all the tests on mutable and non-mutable grids.

	data := []struct {
		grid        Grid
		mGridTest   bool
		description string
	}{
		{
			grid,
			true,
			"mutable",
		},
		{
			roGrid,
			false,
			"ro",
		},
	}

	for _, rec := range data {

		grid := rec.grid

		description := rec.description

		if grid.DataString() != ADVANCED_TEST_GRID {
			t.Error(description, "Advanced grid didn't survive a roundtrip to DataString")
		}

		if grid.numFilledCells() != 27 {
			t.Error(description, "The advanced grid's rank was wrong at load: ", grid.rank())
		}

		grid.HasMultipleSolutions()

		if grid.DataString() != ADVANCED_TEST_GRID {
			t.Error(description, "HasMultipleSolutions mutated the underlying grid.")
		}

		copy := withSimpleCellsFilled(grid)

		if copy.Solved() {
			t.Error(description, "Advanced grid was 'solved' with just fillSimpleCells")
		}

		solutions := grid.Solutions()

		if grid.DataString() != ADVANCED_TEST_GRID {
			t.Error(description, "Calling Solutions() modified the original grid.")
		}

		if len(solutions) != 1 {
			t.Error(description, "We found the wrong number of solutions in Advanced grid:", len(solutions))
		}

		if solutions[0].DataString() != SOLVED_ADVANCED_TEST_GRID {
			t.Error(description, "Solve found the wrong solution.")
		}

		if grid.NumSolutions() != 1 {
			t.Error(description, "Grid didn't find any solutions but there is one.")
		}

		if !grid.HasSolution() {
			t.Error(description, "Grid didn't find any solutions but there is one.")
		}

		if rec.mGridTest {

			mGrid := grid.MutableCopy()

			mGrid.Solve()

			if !mGrid.Solved() {
				t.Error("The grid itself didn't get mutated to a solved state.")
			}

			if mGrid.numFilledCells() != DIM*DIM {
				t.Error("After solving, we didn't think all cells were filled.")
			}

			if mGrid.(*mutableGridImpl).cachedSolutions != nil {
				t.Error("The cache of solutions was supposed to be expired when we copied in the solution, but it wasn't")
				t.Fail()
			}

			if !mGrid.Solve() {
				t.Error("Solve called on already solved grid did not return true")
			}
		}

		//TODO: test that nOrFewerSolutions does stop at max (unless cached)
		//TODO: test HasMultipleSolutions
	}

}

func TestMultiSolutions(t *testing.T) {

	//TODO: consider testing immutable grids here, too. Currently don't since
	//this is such an expensive test anyway and we should have good coverage
	//elsewhere.

	if testing.Short() {
		t.Skip("Skipping TestMultiSolutions in short test mode,")
	}

	var grid MutableGrid

	files := map[string]int{
		"multiple-solutions.sdk":  4,
		"multiple-solutions2.sdk": 2,
	}

	for i := 0; i < 1000; i++ {

		for file, numSolutions := range files {

			grid, _ = MutableLoadSDKFromFile(puzzlePath(file))

			//Test num solutions from the beginning.
			if num := grid.NumSolutions(); num != numSolutions {
				t.Fatal("On run", i, "Grid", file, " with", numSolutions, "solutions was found to only have", num)
			}

			//Get a new version of grid to reset all caches
			grid, _ = MutableLoadSDKFromFile(puzzlePath(file))

			if !grid.HasMultipleSolutions() {
				t.Fatal("On run", i, "Grid", file, "with multiple solutions was reported as only having one.")
			}

			//Test num solutions after already having done other solution gathering.
			//this is in here because at one point calling this after HasMultipleSolutions
			//would find 2 vs 1 calling it fresh.
			if num := grid.NumSolutions(); num != numSolutions {
				t.Fatal("On run", i, "Grid", file, "with", numSolutions, "solutions was found to only have", num, "after calling HasMultipleSolutions first.")
			}

		}
	}

}

func TestTranspose(t *testing.T) {

	//TODO: this test doesn't verify that a grid with specific excludes set is transposed as well
	//(although that does work)

	grid := MutableLoadSDK(TEST_GRID)
	transposedGrid := grid.(*mutableGridImpl).transpose()
	if transposedGrid == nil {
		t.Log("Transpose gave us back a nil grid")
		t.FailNow()
	}
	if transposedGrid == grid {
		t.Log("Transpose did not return a copy")
		t.Fail()
	}
	if transposedGrid.DataString() != TRANSPOSED_TEST_GRID {
		t.Log("Transpose did not operate correctly")
		t.Fail()
	}
}

func TestFill(t *testing.T) {

	grid := NewGrid()
	if !grid.Fill() {
		t.Log("We were unable to find a fill for an empty grid.")
		t.Fail()
	}

	if !grid.Solved() {
		t.Log("The grid that came back from fill was not actually fully solved.")
		t.Fail()
	}

}

func BenchmarkFill(b *testing.B) {
	for i := 0; i < b.N; i++ {
		grid := NewGrid()
		grid.Fill()
	}
}

func BenchmarkAdvancedSolve(b *testing.B) {
	for i := 0; i < b.N; i++ {
		grid := MutableLoadSDK(ADVANCED_TEST_GRID)
		grid.Solve()
	}
}

func BenchmarkDifficulty(b *testing.B) {
	//NOTE: this benchmark is exceptionally noisy--it's heavily dependent on
	//how quickly the difficulty converges given the specific HumanSolutions
	//generated.
	for i := 0; i < b.N; i++ {
		grid := LoadSDK(ADVANCED_TEST_GRID)
		grid.Difficulty()
	}
}

func TestGenerate(t *testing.T) {
	grid := GenerateGrid(nil)

	if grid == nil {
		t.Log("We didn't get back a generated grid")
		t.Fail()
	}

	if grid.Solved() {
		t.Log("We got back a solved generated grid.")
		t.Fail()
	}

	if grid.HasMultipleSolutions() {
		t.Log("We got back a generated grid that has more than one solution.")
		t.Fail()
	}

	lockedCellCount := 0

	for _, cell := range grid.Cells() {
		if cell.Locked() {
			lockedCellCount++
		}
	}

	if lockedCellCount == 0 || lockedCellCount == DIM*DIM {
		t.Error("Found an unreasonable number of locked cells from generate:", lockedCellCount)
	}
}

func TestGridEmpty(t *testing.T) {

	grid := NewGrid()

	if !isMutableGridImpl(grid) {
		t.Fatal("Expected mutable grid impl from NewGrid")
	}

	if !grid.Empty() {
		t.Error("Fresh grid wasn't empty")
	}

	roGrid := grid.Copy()

	if !isGridImpl(roGrid) {
		t.Error("Expected immutable grid impl from NewGrid.Copy()")
	}

	if !grid.Empty() {
		t.Error("RO fresh grid wasn't empty")
	}

	grid.Load(TEST_GRID)

	if grid.Empty() {
		t.Error("A filled grid was reported as empty")
	}

	roGrid = grid.Copy()

	if grid.Diagram(false) != roGrid.Diagram(false) {
		t.Error("Grid and roGrid had different data strings")
	}

	if roGrid.Empty() {
		t.Error("A filled rogrid was reported as empty")
	}

	//Reset the grid
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			grid.MutableCell(r, c).SetNumber(0)
		}
	}

	if !grid.Empty() {
		t.Error("A forcibly cleared grid did not report as empty.")
	}

	roGrid = grid.Copy()

	if !roGrid.Empty() {
		t.Error("A forcibly cleared ro grid did not report as empty")
	}
}

func TestGenerationOptions(t *testing.T) {

	options := DefaultGenerationOptions()

	if options.Symmetry != SYMMETRY_VERTICAL {
		t.Error("Wrong symmetry")
	}

	if options.SymmetryPercentage != 0.7 {
		t.Error("Wrong Symmetry percentage")
	}

	if options.MinFilledCells != 0 {
		t.Error("Wrong MinFilledCells")
	}
}

//This is an extremely expensive test desgined to help ferret out #134.
//TODO: remove this test!
func TestGenerateMultipleSolutions(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping TestGenerateDiabolical in short test mode,")
	}

	var grid MutableGrid

	for i := 0; i < 10; i++ {
		grid = GenerateGrid(nil)

		if grid.HasMultipleSolutions() {
			t.Fatal("On run", i, "we got back a generated grid that has more than one solution: ", grid)
		}
	}
}

func TestGenerateMinFilledCells(t *testing.T) {
	options := GenerationOptions{
		Symmetry:           SYMMETRY_NONE,
		SymmetryPercentage: 1.0,
		MinFilledCells:     DIM*DIM - 1,
	}

	grid := GenerateGrid(&options)

	if grid.numFilledCells() < options.MinFilledCells {
		t.Error("Grid came back with too few cells filled: ", grid.numFilledCells(), "expected", options.MinFilledCells)
	}

	options.Symmetry = SYMMETRY_VERTICAL

	grid = GenerateGrid(&options)

	if grid.numFilledCells() < options.MinFilledCells {
		t.Error("Grid came back with too few cells filled: ", grid.numFilledCells(), "expected", options.MinFilledCells)
	}
}

func TestSymmetricalGenerate(t *testing.T) {
	options := GenerationOptions{
		Symmetry:           SYMMETRY_VERTICAL,
		SymmetryPercentage: 1.5,
	}

	if options.SymmetryPercentage != 1.5 {
		t.Error("GenerateGrid mutated the provided options.")
	}

	grid := GenerateGrid(&options)

	if grid == nil {
		t.Fatal("Did not get a generated grid back")
	}

	for r := 0; r < DIM; r++ {
		//Go through all left side columns, skipping middle column.
		for c := 0; c < (DIM / 2); c++ {
			cell := grid.Cell(r, c)
			otherCell := cell.SymmetricalPartner(SYMMETRY_VERTICAL)
			if cell.Number() != 0 && otherCell.Number() == 0 {
				t.Error("Cell ", cell.Reference().String(), "'s partner not filled but should be")
			}
			if cell.Number() == 0 && otherCell.Number() != 0 {
				t.Error("Cell ", cell.Reference().String(), "'s partner IS filled and should be empty")
			}
		}
	}

	//Now test a non 1.0 symmetry
	percentage := 0.5
	options.SymmetryPercentage = percentage
	grid = GenerateGrid(&options)

	if grid == nil {
		t.Fatal("Did not get a generated grid back with 0.5 symmetry")
	}

	//Counter of how many cells were the opposite of what they should be.
	numWrong := 0.0
	base := 0.0

	for r := 0; r < DIM; r++ {
		//Go through all left side columns, skipping middle column.
		for c := 0; c < (DIM / 2); c++ {
			cell := grid.Cell(r, c)
			otherCell := cell.SymmetricalPartner(SYMMETRY_VERTICAL)

			base++

			if cell.Number() != 0 && otherCell.Number() == 0 {
				numWrong++
			}
			if cell.Number() == 0 && otherCell.Number() != 0 {
				numWrong++
			}
		}
	}

	//We've only looked at some percentage of cells, so that's the target we should be close to.
	percentageMultiplier := base / (DIM * DIM)

	//Note: this test is inherently flaky because there's randomness.
	if math.Abs((numWrong/base)-(percentage*percentageMultiplier)) > 0.1 {
		t.Error("More than the allowable percentage were not symmetrical within the tolerance", numWrong/base)
	}
}

func nCopies(in string, copies int) (result []string) {
	for i := 0; i < copies; i++ {
		result = append(result, in)
	}
	return
}

func puzzlePath(name string) string {

	//NOTE: This is currently duplicated exactly in sdkconverter/converters_test.go

	//Will look for the puzzle in all of the default directories and return its location if it exists. If it doesn't find it, will return ""
	name = strings.ToLower(name)

	directories := list.New()

	directories.PushFront(PUZZLES_DIRECTORY)

	e := directories.Front()

	for e != nil {
		directories.Remove(e)
		directory := e.Value.(string)
		infos, _ := ioutil.ReadDir(directory)
		for _, info := range infos {
			if info.IsDir() {
				//We'll search this one later.
				directories.PushBack(directory + "/" + info.Name())
			} else {
				//See if it's the file.
				if strings.ToLower(info.Name()) == name {
					//Found it!
					return directory + "/" + info.Name()
				}
				//Nope, not it... keep on searching.
			}
		}
		e = directories.Front()
	}
	//Hmm, guess we won't find it.
	return ""
}

func TestLoadFromFile(t *testing.T) {
	grid, err := MutableLoadSDKFromFile(puzzlePath("hiddenpair1.sdk"))
	if err != nil {
		t.Log("Grid loading failed.")
		t.Fail()
	}
	if grid.Cell(0, 1).Number() != 6 {
		t.Log("We didn't get back a grid looking like what we expected.")
		t.Fail()
	}
}

func TestUnlockCells(t *testing.T) {
	grid := NewGrid()

	if !isMutableGridImpl(grid) {
		t.Fatal("Expected Newgrid to return mutable grid")
	}

	for i := 0; i < DIM; i++ {
		grid.MutableCell(i, i).Lock()
	}

	someCellsLocked := false
	for _, cell := range grid.Cells() {
		if cell.Locked() {
			someCellsLocked = true
		}
	}

	if !someCellsLocked {
		t.Error("After locking some cells, no cells were locked")
	}

	roGrid := grid.Copy()

	for _, cell := range roGrid.Cells() {
		if cell.Locked() {
			someCellsLocked = true
		}
	}

	if !someCellsLocked {
		t.Error("After locking some cells and copying, no cells were locked in ro grid.")
	}

	grid.UnlockCells()

	for _, cell := range grid.Cells() {
		if cell.Locked() {
			t.Fatal("Found a locked cell after calling grid.UnlockCells", cell)
		}
	}

	roGrid = grid.Copy()

	for _, cell := range roGrid.Cells() {
		if cell.Locked() {
			t.Error("Found a locked cell after calling grid.UnlockCells and copying", cell)
		}
	}
}

func TestResetUnlockedCells(t *testing.T) {
	grid := MutableLoadSDK(TEST_GRID)

	beforeExcludesDiagram := grid.Diagram(false)
	beforeMarksDiagram := grid.Diagram(true)

	roGrid := grid.Copy()

	if !isGridImpl(roGrid) {
		t.Fatal("Expected grid.Copy to give a rogrid")
	}

	if roGrid.Diagram(false) != beforeExcludesDiagram {
		t.Error("RO Grid didn't have right excludes")
	}

	if roGrid.Diagram(true) != beforeMarksDiagram {
		t.Error("RO Grid didn't have right marks")
	}

	grid.MutableCell(0, 4).SetNumber(3)
	grid.MutableCell(0, 5).SetMark(1, true)
	grid.MutableCell(0, 6).SetExcluded(1, true)

	grid.ResetUnlockedCells()

	roGrid = grid.Copy()

	if grid.Diagram(false) != beforeExcludesDiagram {
		t.Error("Reseting unlocked cells didn't get right outcome. Got\n\n", grid.Diagram(false), "\n\nwanted\n\n", beforeExcludesDiagram)
	}

	if grid.Diagram(true) != beforeMarksDiagram {
		t.Error("Reseting unlocked cells didn't get right outcome. Got\n\n", grid.Diagram(true), "\n\nwanted\n\n", beforeMarksDiagram)
	}

	if roGrid.Diagram(false) != beforeExcludesDiagram {
		t.Error("Reseting unlocked cells didn't get right outcome in rogrid. Got\n\n", roGrid.Diagram(false), "\n\nwanted\n\n", beforeExcludesDiagram)
	}

	if roGrid.Diagram(true) != beforeMarksDiagram {
		t.Error("Reseting unlocked cells didn't get right outcome in rogrid. Got\n\n", roGrid.Diagram(true), "\n\nwanted\n\n", beforeMarksDiagram)
	}

}

func TestNumFilledCells(t *testing.T) {
	grid := NewGrid()

	if !isMutableGridImpl(grid) {
		t.Fatal("Expected NewGrid to give mutable grid")
	}

	roGrid := grid.Copy()

	if !isGridImpl(roGrid) {
		t.Fatal("Expected grid.Copy to give ro grid")
	}

	if grid.numFilledCells() != 0 {
		t.Error("New grid thought it already had filled cells")
	}

	if roGrid.numFilledCells() != 0 {
		t.Error("New grid thought it already had filled cells in ro copy")
	}

	grid.MutableCell(0, 0).SetNumber(1)

	roGrid = grid.Copy()

	if grid.numFilledCells() != 1 {
		t.Error("Grid with one cell set didn't think it had any filled cells.")
	}

	if roGrid.numFilledCells() != 1 {
		t.Error("Grid with one cell set didn't think it had any filled cells in ro copy.")
	}

	grid.MutableCell(0, 0).SetNumber(2)

	roGrid = grid.Copy()

	if grid.numFilledCells() != 1 {
		t.Error("Grid with a number set on a cell after another cell didn't notice that it was still just one cell.")
	}

	if roGrid.numFilledCells() != 1 {
		t.Error("Grid with a number set on a cell after another cell didn't notice that it was still just one cell in rocopy.")
	}

	grid.MutableCell(0, 0).SetNumber(0)

	roGrid = grid.Copy()

	if grid.numFilledCells() != 0 {
		t.Error("Grid with cell unset didn't notice that it was now zero again")
	}

	if roGrid.numFilledCells() != 0 {
		t.Error("Grid with cell unset didn't notice that it was now zero again in ro copy")
	}

	grid.MutableCell(0, 0).SetNumber(0)

	roGrid = grid.Copy()

	if grid.numFilledCells() != 0 {
		t.Error("Setting a cell to 0 that was already zero got wrong num filled cells.")
	}

	if roGrid.numFilledCells() != 0 {
		t.Error("Setting a cell to 0 that was already zero got wrong num filled cells in rogrid.")
	}

}

func TestLockFilledCells(t *testing.T) {
	grid := MutableLoadSDK(TEST_GRID)

	var lockedCell Cell

	for _, cell := range grid.MutableCells() {

		if cell.Number() == 0 {
			cell.Lock()
			lockedCell = cell
			break
		}
	}

	grid.LockFilledCells()

	for _, cell := range grid.Cells() {

		if cell.Number() != 0 && !cell.Locked() {
			t.Error("Found a cell that was filled but not locked after LockFilledCells", cell)
		}

		if cell.Number() == 0 && cell.Locked() {

			if cell != lockedCell {
				t.Error("Found a cell that was unfilled but locked after LockFilledCells", cell)
			}
		}
	}

	if !lockedCell.Locked() {
		t.Error("The specially locked cell was unlocked after calling LockFilledCells", lockedCell)
	}
}
