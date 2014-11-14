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

func init() {
	runtime.GOMAXPROCS(NUM_SOLVER_THREADS)
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

	grid.Done()
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

	//Twiddle an exclude to make sure it copies over correctly.
	grid.Cell(2, 0).setExcluded(4, true)

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

	for c, cell := range grid.cells {
		copyCell := cell.InGrid(copy)
		if !IntSlice(cell.impossibles[:]).SameAs(IntSlice(copyCell.impossibles[:])) {
			t.Error("Cells at position", c, "had different impossibles:", cell.impossibles, copyCell.impossibles)
		}
		for i := 1; i <= DIM; i++ {
			if cell.Possible(i) != copyCell.Possible(i) {
				t.Error("The copy of the grid did not have the same possible at cell ", c, " i ", i)
			}
		}
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

	grid.Done()
	copy.Done()

}

func TestAdvancedSolve(t *testing.T) {
	grid := NewGrid()
	grid.Load(ADVANCED_TEST_GRID)

	if grid.DataString() != ADVANCED_TEST_GRID {
		t.Log("Advanced grid didn't survive a roundtrip to DataString")
		t.Fail()
	}

	if grid.numFilledCells != 27 {
		t.Log("The advanced grid's rank was wrong at load: ", grid.Rank())
		t.Fail()
	}

	grid.HasMultipleSolutions()

	if grid.DataString() != ADVANCED_TEST_GRID {
		t.Log("HasMultipleSolutions mutated the underlying grid.")
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

	//TODO: test that nOrFewerSolutions does stop at max (unless cached)
	//TODO: test HasMultipleSolutions

	grid.Done()
	copy.Done()

}

func TestTranspose(t *testing.T) {
	grid := NewGrid()
	grid.Load(TEST_GRID)
	transposedGrid := grid.transpose()
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
	grid.Done()
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
		grid.Load(ADVANCED_TEST_GRID)
		grid.Solve()
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
	grid := GenerateGrid(SYMMETRY_NONE, 0.0)

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

	grid.Done()
}

func TestSymmetricalGenerate(t *testing.T) {
	grid := GenerateGrid(SYMMETRY_VERTICAL, 1.0)

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

	//Now test a non 1.0 symmetry
	percentage := 0.5
	grid = GenerateGrid(SYMMETRY_VERTICAL, percentage)

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
	if !grid.LoadFromFile(puzzlePath("hiddenpair1.sdk")) {
		t.Log("Grid loading failed.")
		t.Fail()
	}
	if grid.Cell(0, 1).Number() != 6 {
		t.Log("We didn't get back a grid looking like what we expected.")
		t.Fail()
	}
	grid.Done()
}
