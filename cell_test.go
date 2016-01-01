package sudoku

import (
	"log"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const TOP_LEFT_CELL = `123|
456|
789|
---+`

const TOP_LEFT_CELL_NO_BLOCK_DIM = `12 |
456|
789|
---+`

const TOP_LEFT_CELL_FILLED = `•••|
•1•|
•••|
---+`

func TestCellCreation(t *testing.T) {

	number := 1
	data := strconv.Itoa(number)
	cell := newCell(nil, 0, 0)

	if cell.rank() != DIM {
		t.Log("Cell's rank was not DIM when initalized")
		t.Fail()
	}

	ref := cell.ref()

	if ref.col != 0 || ref.row != 0 {
		t.Error("Cell ref came back wrong")
	}

	possibilities := cell.Possibilities()

	if len(possibilities) != DIM {
		t.Log("We got back the wrong number of possibilities")
		t.Fail()
	}

	for i, possibility := range possibilities {
		if possibility != i+1 {
			t.Log("The possibilities list was not a monotonically increasing list: ", possibility, "/", i)
			t.Fail()
		}
	}

	for i := 1; i <= DIM; i++ {
		if !cell.Possible(i) {
			t.Log("Cell reported ", i, " was impossible even though it hadn't been touched")
			t.Fail()
		}
		cell.setImpossible(i)
		if cell.Possible(i) {
			t.Log("Cell reported ", i, " was possible even though we explicitly set it as impossible")
			t.Fail()
		}

		if cell.Invalid() {
			t.Log("Cell reported it was invalid even though only one number is impossible.")
			t.Fail()
		}

		if cell.rank() != DIM-1 {
			t.Log("Cell reported an incorrect rank: ", cell.rank())
			t.Fail()
		}

		cell.setImpossible(i)
		cell.setPossible(i)
		if cell.Possible(i) {
			t.Log("Cell reported ", i, " was possible even though we'd only called Possible 1x and Impossible 2x")
			t.Fail()
		}
		cell.setPossible(i)
		if !cell.Possible(i) {
			t.Log("Cell reported ", i, " was impossible even after matched calls to setPossible/setImpossible")
			t.Fail()
		}
		cell.SetExcluded(i, true)
		if cell.Possible(i) {
			t.Log("A cell reported it was possible even though its number had been manually excluded")
			t.Fail()
		}
		cell.ResetExcludes()
		if !cell.Possible(i) {
			t.Log("A cell thought it was not possible even after excludes were cleared")
			t.Fail()
		}
	}

	for i := 1; i <= DIM; i++ {
		cell.setImpossible(i)
	}
	if !cell.Invalid() {
		t.Log("Cell didn't realize it was invalid even though every number is impossible.")
		t.Fail()
	}

	for i := 1; i <= DIM; i++ {
		cell.setPossible(i)
		if cell.implicitNumber() != i {
			t.Log("Implicit number failed to notice that ", i, " should be implict number.")
			t.Fail()
		}
		possibilities := cell.Possibilities()
		if len(possibilities) != 1 {
			t.Log("We got the wrong number of possibilities back")
			t.Fail()
		} else if possibilities[0] != i {
			t.Log("We got the wrong possibility back")
			t.Fail()
		}
		cell.setImpossible(i)
	}

	for i := 1; i <= DIM; i++ {
		cell.setPossible(i)
	}

	if cell.Invalid() {
		t.Log("Cell still thinks it's invalid even though we reset all possible counters.")
		t.Fail()
	}

	cell.load(data)
	if cell.Number() != number {
		t.Log("Number came back wrong")
		t.Fail()
	}
	if cell.Row() != 0 {
		t.Log("Row came back wrong")
		t.Fail()
	}
	if cell.Col() != 0 {
		t.Log("Cell came back wrong")
		t.Fail()
	}
	if cell.dataString() != data {
		t.Log("Cell round-tripped out with different string than data in")
		t.Fail()
	}
	//TODO: test failing for values that are too high.
	for i := 1; i <= DIM; i++ {
		if i == number {
			if !cell.Possible(i) {
				t.Log("We reported back a number we were explicitly set to was impossible")
				t.Fail()
			}
		} else if cell.Possible(i) {
			t.Log("We reported back that a number was possible when another number had been explicitly set")
			t.Fail()
		}
	}

	number = 2

	cell.SetNumber(number)

	if cell.Number() != number {
		t.Log("Number came back wrong after being set with SetNumber")
		t.Fail()
	}

	if cell.rank() != 0 {
		t.Log("Cell with an explicit number came back with a non-0 Rank")
		t.Fail()
	}

	for i := 1; i <= DIM; i++ {
		if i == number {
			if !cell.Possible(i) {
				t.Log("We reported back a number we were explicitly set to was impossible (2nd set)")
				t.Fail()
			}
		} else if cell.Possible(i) {
			t.Log("We reported back that a number was possible when another number had been explicitly set (2nd set)")
			t.Fail()
		}
	}

	number = 0

	cell.SetNumber(number)

	if cell.Number() != number {
		t.Log("Number came back wrong after being set with SetNumber to 0")
		t.Fail()
	}

	for i := 1; i <= DIM; i++ {
		if !cell.Possible(i) {
			t.Log("We reported back a number was not possible when 0 had been explicitly set.")
			t.Fail()
		}
	}

	if cell.diagram() != TOP_LEFT_CELL {
		t.Log("Diagram for the top left cell printed out incorrectly: \n", cell.diagram())
		t.Fail()
	}

	cell.setImpossible(BLOCK_DIM)

	if cell.diagram() != TOP_LEFT_CELL_NO_BLOCK_DIM {
		t.Log("Diagram for the top left cell printed out incorrectly: \n", cell.diagram())
		t.Fail()
	}

	cell.SetNumber(1)

	if cell.diagram() != TOP_LEFT_CELL_FILLED {
		t.Log("Diagram for the top left cell filled printed out incorrectly: \n", cell.diagram())
		t.Fail()
	}
}

func TestDiagramExtents(t *testing.T) {
	//Testing this for real is hard, but what we can do is as at least make
	//sure that the given indexes aren't running into any of the edges of the
	//diagram.

	//TODO: also test that the right cells are returned.
	grid := NewGrid()
	diagram := grid.Diagram(false)
	diagramRows := strings.Split(diagram, "\n")
	for i, cell := range grid.Cells() {
		top, left, height, width := cell.DiagramExtents()
		for r := top; r < top+height; r++ {
			row := diagramRows[r]
			for c := left; c < left+width; c++ {
				char := strings.Split(row, "")[c]
				if char == DIAGRAM_RIGHT || char == DIAGRAM_BOTTOM || char == DIAGRAM_CORNER {
					t.Error("In cell", i, "invalid char at", r, c)
				}
			}
		}
	}
}

func TestMarks(t *testing.T) {
	grid := NewGrid()
	cell := grid.Cell(0, 0)
	for i := 1; i < DIM+1; i++ {
		if cell.Mark(i) {
			t.Error("Zero cell had a mark:", i)
		}
	}
	if cell.Mark(0) {
		t.Error("Invalid index had true mark: 0")
	}
	if cell.Mark(DIM + 2) {
		t.Error("Invalid index had true mark: ", DIM+2)
	}

	if len(cell.Marks()) != 0 {
		t.Error("An empty cell already had marks")
	}

	cell.SetMark(1, true)
	if !cell.Mark(1) {
		t.Error("Cell with a mark on 1 did not read back")
	}

	cell.SetMark(2, true)

	if !reflect.DeepEqual(cell.Marks(), IntSlice{1, 2}) {
		t.Error("Cell with marks 1 and 2 set had wrong Marks List:", cell.Marks())
	}

	cell.ResetMarks()
	for i := 1; i < DIM; i++ {
		if cell.Mark(i) {
			t.Error("Cell that had called ResetMarks still had a mark set at", i)
		}
	}
}

func TestCellLock(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()

	cell := grid.Cell(3, 3)

	if cell.Locked() {
		t.Error("New cell was already locked")
	}

	cell.Lock()

	if !cell.Locked() {
		t.Error("Locked cell was not actually locked")
	}

	cell.Unlock()

	if cell.Locked() {
		t.Error("Unlocked cell was still locked")
	}

	cell.SetNumber(2)

	if cell.Number() != 2 {
		t.Error("Locking made it so SetNumber failed")
	}
}

func TestSymmetry(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()

	cell := grid.Cell(3, 3)

	partner := cell.SymmetricalPartner(SYMMETRY_BOTH)

	if partner.Row() != 5 || partner.Col() != 5 {
		t.Error("Got wrong symmetrical partner (both) for 3,3: ", partner)
	}

	partner = cell.SymmetricalPartner(SYMMETRY_HORIZONTAL)

	if partner.Row() != 5 || partner.Col() != 3 {
		t.Error("Got wrong symmetrical partner (horizontal) for 3,3: ", partner)
	}

	partner = cell.SymmetricalPartner(SYMMETRY_VERTICAL)

	if partner.Row() != 3 || partner.Col() != 5 {
		t.Error("Got wrong symmetrical partner (vertical) for 3,3: ", partner)
	}

	partner = cell.SymmetricalPartner(SYMMETRY_ANY)

	if partner == nil {
		t.Error("Didn't get back a symmerical partner for (any) for 3,3")
	}

	partner = cell.SymmetricalPartner(SYMMETRY_NONE)

	if partner != nil {
		t.Error("Should have gotten back nil for SYMMETRY_NONE for 3,3, got: ", partner)
	}

	cell = grid.Cell(4, 4)

	if cell.SymmetricalPartner(SYMMETRY_BOTH) != nil || cell.SymmetricalPartner(SYMMETRY_HORIZONTAL) != nil || cell.SymmetricalPartner(SYMMETRY_VERTICAL) != nil {
		t.Error("Middle cell got a symmetrical partner for some kind of symmetry.")
	}

}
