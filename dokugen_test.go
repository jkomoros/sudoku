package dokugen

import (
	"testing"
)

func TestCellCreation(t *testing.T) {
	cell := NewCell(nil, 0, 0, "1")
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
	//TODO: test failing for values that are too high.
}

func TestBasic(t *testing.T) {
	_ = Grid{}
}
