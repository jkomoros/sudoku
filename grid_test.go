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

func BenchmarkGridCopy(b *testing.B) {
	grid := NewGrid()
	grid.LoadSDK(ADVANCED_TEST_GRID)
	for i := 0; i < b.N; i++ {
		gridCopy := grid.Copy()
		defer gridCopy.Done()
	}
	grid.Done()
}

func TestGridCopy(t *testing.T) {
	grid := NewGrid()
	grid.LoadSDK(ADVANCED_TEST_GRID)

	cell := grid.MutableCell(0, 0)
	cell.SetMark(3, true)
	cell.SetMark(4, true)

	cell = grid.MutableCell(0, 2)
	cell.SetExcluded(3, true)
	cell.SetExcluded(4, true)

	gridCopy := grid.Copy()

	if grid.Diagram(true) != gridCopy.Diagram(true) {
		t.Error("Grid and copy don't match in marks")
	}

	if grid.Diagram(false) != gridCopy.Diagram(false) {
		t.Error("Grid and copy don't match in terms of excludes")
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
	grid := NewGrid()
	grid.LoadSDK(data)
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

	if grid.numFilledCells != DIM*DIM {
		t.Log("We didn't think all cells were filled, but they were!")
		t.Fail()
	}

	for count := 0; count < DIM; count++ {
		col := grid.Col(count)
		if num := len(col); num != DIM {
			t.Log("We got back a column but it had the wrong amount of items: ", num, "\n")
			t.Fail()
		}
		for i, cell := range col {
			if cell.Col() != count {
				t.Log("One of the cells we got back when asking for column ", count, " was not in the right column.")
				t.Fail()
			}
			if cell.Row() != i {
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
			if cell.Row() != count {
				t.Log("One of the cells we got back when asking for row ", count, " was not in the right rows.")
				t.Fail()
			}
			if cell.Col() != i {
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
			if cell.Block() != count {
				t.Log("We got a cell back in a block with the wrong block number")
				t.Fail()
			}
		}

		if block[0].Row() != blockUpperLeftRow[count] || block[0].Col() != blockUpperLeftCol[count] {
			t.Log("We got back the wrong first cell from block ", count, ": ", block[0])
			t.Fail()
		}

		if block[DIM-1].Row() != blockUpperLeftRow[count]+BLOCK_DIM-1 || block[DIM-1].Col() != blockUpperLeftCol[count]+BLOCK_DIM-1 {
			t.Log("We got back the wrong last cell from block ", count, ": ", block[0])
			t.Fail()
		}

	}

	cell := grid.Cell(2, 2)

	if cell.Row() != 2 || cell.Col() != 2 {
		t.Log("We grabbed a cell but what we got back was the wrong row and col.")
		t.Fail()
	}

	neighbors := cell.Neighbors()

	if len(neighbors) != _NUM_NEIGHBORS {
		t.Log("We got a different number of neighbors than what we were expecting: ", len(neighbors))
		t.Fail()
	}
	neighborsMap := make(map[Cell]bool)
	for _, neighbor := range neighbors {
		if neighbor == nil {
			t.Log("We found a nil neighbor")
			t.Fail()
		}
		if neighbor.Row() != cell.Row() && neighbor.Col() != cell.Col() && neighbor.Block() != cell.Block() {
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

	grid.Done()
}

func TestGridCells(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()

	grid.LoadSDK(TEST_GRID)

	cells := grid.Cells()

	if len(cells) != DIM*DIM {
		t.Fatal("Grid.cells gave back a cellslice with wrong number of cells", len(cells))
	}

	//Make sure it's the same cells
	for i, cell := range cells {
		if cell.grid() != grid {
			t.Error("cell #", i, "had wrong grid")
		}
		if cell != grid.Cell(cell.Row(), cell.Col()) {
			t.Error("cell #", i, "was not the same as the right one in the grid")
		}
	}

	//make sure mutating a cell from the celllist mutates the grid.
	cell := cells[3].Mutable()

	if cell.Number() != 0 {
		t.Fatal("We expected cell #3 to be empty, but had", cell.Number())
	}
	cell.SetNumber(3)

	if grid.Cell(0, 3).Number() != 3 {
		t.Error("Mutating a cellin cells did not mutate the grid.")
	}
}

func TestGridLoad(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	grid.LoadSDK(TEST_GRID)

	cell := grid.MutableCell(0, 0)

	if cell.Number() != 6 {
		t.Log("The loaded grid did not have a 6 in the upper left corner")
		t.Fail()
	}

	cell = grid.MutableCell(DIM-1, DIM-1)

	if cell.Number() != 7 {
		t.Log("The loaded grid did not have a 7 in the bottom right corner")
		t.Fail()
	}

	if grid.DataString() != TEST_GRID {
		t.Log("The real test grid did not survive a round trip via DataString: \n", grid.DataString(), "\n\n", TEST_GRID)
		t.Fail()
	}

	if grid.Diagram(false) != TEST_GRID_DIAGRAM {
		t.Log("The grid did not match the expected diagram: \n", grid.Diagram(false))
		t.Fail()
	}

	//Twiddle an exclude to make sure it copies over correctly.
	grid.MutableCell(2, 0).SetExcluded(4, true)

	if grid.Diagram(false) != TEST_GRID_EXCLUDED_DIAGRAM {
		t.Error("Diagram did not reflect the manually excluded item: \n", grid.Diagram(false))
	}

	//Test copying.

	copy := grid.Copy()
	defer copy.Done()

	if grid.DataString() != copy.DataString() {
		t.Log("Copied grid does not have the same datastring!")
		t.Fail()
	}

	for c, cell := range grid.Cells() {
		copyCell := cell.InGrid(copy)
		cellI := cell.Mutable().impl()
		copyCellI := copyCell.Mutable().impl()
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
		t.Log("When we modified the copy's cell, it also affected the original.")
		t.Fail()
	}

	if grid.Solved() {
		t.Log("Grid reported it was solved when it was not.")
		t.Fail()
	}

	if grid.Invalid() {
		t.Log("Grid thought it was invalid when it wasn't: \n", grid.Diagram(false))
		t.Fail()
	}

	if num := grid.fillSimpleCells(); num != 45 {
		t.Log("We filled simple cells on the test grid but didn't get as many as we were expecting: ", num, "/", 45)
		t.Fail()
	}

	if grid.Invalid() {
		t.Log("fillSimpleCells filled in something that made the grid invalid: \n", grid.Diagram(false))
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
	defer grid.Done()
	grid.LoadSDKFromFile(puzzlePath("advancedtestgrid.sdk"))

	if grid.DataString() != ADVANCED_TEST_GRID {
		t.Log("Advanced grid didn't survive a roundtrip to DataString")
		t.Fail()
	}

	if grid.numFilledCells != 27 {
		t.Log("The advanced grid's rank was wrong at load: ", grid.rank())
		t.Fail()
	}

	grid.HasMultipleSolutions()

	if grid.DataString() != ADVANCED_TEST_GRID {
		t.Log("HasMultipleSolutions mutated the underlying grid.")
		t.Fail()
	}

	copy := grid.Copy()
	defer copy.Done()

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

	if grid.NumSolutions() != 1 {
		t.Log("Grid didn't find any solutions but there is one.")
		t.Fail()
	}

	if !grid.HasSolution() {
		t.Log("Grid didn't find any solutions but there is one.")
		t.Fail()
	}

	grid.Solve()

	if !grid.Solved() {
		t.Log("The grid itself didn't get mutated to a solved state.")
		t.Fail()
	}

	if grid.numFilledCells != DIM*DIM {
		t.Log("After solving, we didn't think all cells were filled.")
		t.Fail()
	}

	if grid.cachedSolutions != nil {
		t.Log("The cache of solutions was supposed to be expired when we copied in the solution, but it wasn't")
		t.Fail()
	}

	if !grid.Solve() {
		t.Error("Solve called on already solved grid did not return true")
	}

	//TODO: test that nOrFewerSolutions does stop at max (unless cached)
	//TODO: test HasMultipleSolutions

}

func TestMultiSolutions(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping TestMultiSolutions in short test mode,")
	}

	var grid *Grid

	files := map[string]int{
		"multiple-solutions.sdk":  4,
		"multiple-solutions2.sdk": 2,
	}

	for i := 0; i < 1000; i++ {

		for file, numSolutions := range files {

			grid = NewGrid()
			grid.LoadSDKFromFile(puzzlePath(file))

			//Test num solutions from the beginning.
			if num := grid.NumSolutions(); num != numSolutions {
				t.Fatal("On run", i, "Grid", file, " with", numSolutions, "solutions was found to only have", num)
			}

			grid.Done()

			//Get a new version of grid to reset all caches
			grid = NewGrid()
			grid.LoadSDKFromFile(puzzlePath(file))

			if !grid.HasMultipleSolutions() {
				t.Fatal("On run", i, "Grid", file, "with multiple solutions was reported as only having one.")
			}

			//Test num solutions after already having done other solution gathering.
			//this is in here because at one point calling this after HasMultipleSolutions
			//would find 2 vs 1 calling it fresh.
			if num := grid.NumSolutions(); num != numSolutions {
				t.Fatal("On run", i, "Grid", file, "with", numSolutions, "solutions was found to only have", num, "after calling HasMultipleSolutions first.")
			}

			grid.Done()
		}
	}

}

func TestTranspose(t *testing.T) {

	//TODO: this test doesn't verify that a grid with specific excludes set is transposed as well
	//(although that does work)

	grid := NewGrid()
	defer grid.Done()
	grid.LoadSDK(TEST_GRID)
	transposedGrid := grid.transpose()
	defer transposedGrid.Done()
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

	grid.Done()

}

func BenchmarkFill(b *testing.B) {
	for i := 0; i < b.N; i++ {
		grid := NewGrid()
		grid.Fill()
		grid.Done()
	}
}

func BenchmarkAdvancedSolve(b *testing.B) {
	for i := 0; i < b.N; i++ {
		grid := NewGrid()
		grid.LoadSDK(ADVANCED_TEST_GRID)
		grid.Solve()
		grid.Done()
	}
}

func BenchmarkDifficulty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		grid := NewGrid()
		grid.LoadSDK(ADVANCED_TEST_GRID)
		grid.Difficulty()
		grid.Done()
	}
}

func TestGridCache(t *testing.T) {
	//TODO: these tests aren't that great.

	//Make sure we're in a known state.
	dropGrids()

	grid := getGrid()
	if grid == nil {
		t.Log("We got back an empty grid from GetGrid")
		t.Fail()
	}
	other := getGrid()
	if grid == other {
		t.Log("We got back the same grid without returning it first.")
		t.Fail()
	}
	returnGrid(grid)
	third := getGrid()
	if third != grid {
		t.Log("We aren't reusing grids as often as we should be.")
		t.Fail()
	}
}

func TestGenerate(t *testing.T) {
	grid := GenerateGrid(nil)

	defer grid.Done()

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

	if !grid.Empty() {
		t.Error("Fresh grid wasn't empty")
	}

	grid.LoadSDK(TEST_GRID)

	if grid.Empty() {
		t.Error("A filled grid was reported as empty")
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

	var grid *Grid

	for i := 0; i < 10; i++ {
		grid = GenerateGrid(nil)

		defer grid.Done()

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

	if grid.numFilledCells < options.MinFilledCells {
		t.Error("Grid came back with too few cells filled: ", grid.numFilledCells, "expected", options.MinFilledCells)
	}

	options.Symmetry = SYMMETRY_VERTICAL

	grid = GenerateGrid(&options)

	if grid.numFilledCells < options.MinFilledCells {
		t.Error("Grid came back with too few cells filled: ", grid.numFilledCells, "expected", options.MinFilledCells)
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

	defer grid.Done()

	if grid == nil {
		t.Fatal("Did not get a generated grid back")
	}

	for r := 0; r < DIM; r++ {
		//Go through all left side columns, skipping middle column.
		for c := 0; c < (DIM / 2); c++ {
			cell := grid.Cell(r, c)
			otherCell := cell.SymmetricalPartner(SYMMETRY_VERTICAL)
			if cell.Number() != 0 && otherCell.Number() == 0 {
				t.Error("Cell ", cell.ref().String(), "'s partner not filled but should be")
			}
			if cell.Number() == 0 && otherCell.Number() != 0 {
				t.Error("Cell ", cell.ref().String(), "'s partner IS filled and should be empty")
			}
		}
	}

	//We're going to clobber this variable with another grid.
	grid.Done()

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
	grid := NewGrid()
	if !grid.LoadSDKFromFile(puzzlePath("hiddenpair1.sdk")) {
		t.Log("Grid loading failed.")
		t.Fail()
	}
	if grid.Cell(0, 1).Number() != 6 {
		t.Log("We didn't get back a grid looking like what we expected.")
		t.Fail()
	}
	grid.Done()
}

func TestUnlockCells(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()

	for i := 0; i < DIM; i++ {
		grid.MutableCell(i, i).Lock()
	}

	someCellsLocked := false
	for i := range grid.cells {
		cell := grid.cells[i]
		if cell.Locked() {
			someCellsLocked = true
		}
	}

	if !someCellsLocked {
		t.Error("After locking some cells, no cells were locked")
	}

	grid.UnlockCells()

	for i := range grid.cells {
		cell := grid.cells[i]
		if cell.Locked() {
			t.Fatal("Found a locked cell after calling grid.UnlockCells", cell)
		}
	}
}

func TestResetUnlockedCells(t *testing.T) {
	grid := NewGrid()
	grid.LoadSDK(TEST_GRID)
	defer grid.Done()

	grid.LockFilledCells()

	beforeExcludesDiagram := grid.Diagram(false)
	beforeMarksDiagram := grid.Diagram(true)

	grid.MutableCell(0, 4).SetNumber(3)
	grid.MutableCell(0, 5).SetMark(1, true)
	grid.MutableCell(0, 6).SetExcluded(1, true)

	grid.ResetUnlockedCells()

	if grid.Diagram(false) != beforeExcludesDiagram {
		t.Error("Reseting unlocked cells didn't get right outcome. Got\n\n", grid.Diagram(false), "\n\nwanted\n\n", beforeExcludesDiagram)
	}

	if grid.Diagram(true) != beforeMarksDiagram {
		t.Error("Reseting unlocked cells didn't get right outcome. Got\n\n", grid.Diagram(true), "\n\nwanted\n\n", beforeMarksDiagram)
	}

}

func TestNumFilledCells(t *testing.T) {
	grid := NewGrid()

	if grid.numFilledCells != 0 {
		t.Error("New grid thought it already had filled cells")
	}

	grid.MutableCell(0, 0).SetNumber(1)

	if grid.numFilledCells != 1 {
		t.Error("Grid with one cell set didn't think it had any filled cells.")
	}

	grid.MutableCell(0, 0).SetNumber(2)

	if grid.numFilledCells != 1 {
		t.Error("Grid with a number set on a cell after another cell didn't notice that it was still just one cell.")
	}

	grid.MutableCell(0, 0).SetNumber(0)

	if grid.numFilledCells != 0 {
		t.Error("Grid with cell unset didn't notice that it was now zero again")
	}

	grid.MutableCell(0, 0).SetNumber(0)

	if grid.numFilledCells != 0 {
		t.Error("Setting a cell to 0 that was already zero got wrong num filled cells.")
	}
}

func TestLockFilledCells(t *testing.T) {
	grid := NewGrid()
	grid.LoadSDK(TEST_GRID)
	defer grid.Done()

	var lockedCell Cell

	for i := range grid.cells {
		cell := &grid.cells[i]

		if cell.Number() == 0 {
			cell.Lock()
			lockedCell = cell
			break
		}
	}

	grid.LockFilledCells()

	for i := range grid.cells {
		cell := &grid.cells[i]

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
