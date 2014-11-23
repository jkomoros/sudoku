package sudoku

import (
	"strconv"
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
	cell := NewCell(nil, 0, 0)

	if cell.Rank() != DIM {
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

		if cell.Rank() != DIM-1 {
			t.Log("Cell reported an incorrect rank: ", cell.Rank())
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
		cell.setExcluded(i, true)
		if cell.Possible(i) {
			t.Log("A cell reported it was possible even though its number had been manually excluded")
			t.Fail()
		}
		cell.resetExcludes()
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

	cell.Load(data)
	if cell.Number() != number {
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

	if cell.Rank() != 0 {
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

	if cell.Diagram() != TOP_LEFT_CELL {
		t.Log("Diagram for the top left cell printed out incorrectly: \n", cell.Diagram())
		t.Fail()
	}

	cell.setImpossible(BLOCK_DIM)

	if cell.Diagram() != TOP_LEFT_CELL_NO_BLOCK_DIM {
		t.Log("Diagram for the top left cell printed out incorrectly: \n", cell.Diagram())
		t.Fail()
	}

	cell.SetNumber(1)

	if cell.Diagram() != TOP_LEFT_CELL_FILLED {
		t.Log("Diagram for the top left cell filled printed out incorrectly: \n", cell.Diagram())
		t.Fail()
	}
}

func TestSymmetry(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()

	cell := grid.Cell(3, 3)

	partner := cell.SymmetricalPartner(SYMMETRY_BOTH)

	if partner.Row != 5 || partner.Col != 5 {
		t.Error("Got wrong symmetrical partner (both) for 3,3: ", partner)
	}

	partner = cell.SymmetricalPartner(SYMMETRY_HORIZONTAL)

	if partner.Row != 5 || partner.Col != 3 {
		t.Error("Got wrong symmetrical partner (horizontal) for 3,3: ", partner)
	}

	partner = cell.SymmetricalPartner(SYMMETRY_VERTICAL)

	if partner.Row != 3 || partner.Col != 5 {
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
