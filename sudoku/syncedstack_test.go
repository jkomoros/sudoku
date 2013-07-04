package sudoku

import (
	"testing"
)

func TestBasic(t *testing.T) {
	stack := NewSyncedStack()
	if stack == nil {
		t.Log("Didn't get a stack back.")
		t.Fail()
	}
	if stack.Length() != 0 {
		t.Log("A new stack did not have a length of one")
		t.Fail()
	}
	item := map[string]int{"a": 1}
	stack.Insert(item)
	if stack.Length() != 1 {
		t.Log("We inserted an item but the length did not go up by one.")
		t.Fail()
	}
}
