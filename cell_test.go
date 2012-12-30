package dokugen

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

func TestCellCreation(t *testing.T) {

	number := 1
	data := strconv.Itoa(number)
	cell := NewCell(nil, 0, 0)

	if cell.Rank() != DIM {
		t.Log("Cell's rank was not DIM when initalized")
		t.Fail()
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
		t.Log("Diagram for the top left cell printed out incorrectly")
		t.Fail()
	}

	cell.setImpossible(BLOCK_DIM)

	if cell.Diagram() != TOP_LEFT_CELL_NO_BLOCK_DIM {
		t.Log("Diagram for the top left cell printed out incorrectly")
		t.Fail()
	}
}
