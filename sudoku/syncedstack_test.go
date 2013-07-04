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
	rawResult := stack.Pop()
	if rawResult == nil {
		t.Log("We didn't get back an item from a queue with one item.")
		t.Fail()
	}
	result := rawResult.(map[string]int)
	if result["a"] != item["a"] {
		t.Log("We didn't get back the item we put in.")
		t.Fail()
	}
	if stack.Length() != 0 {
		t.Log("We removed an item but the stack still has one.")
		t.Fail()
	}
	if stack.Pop() != nil {
		t.Log("We were able to get another item out of the stack even though there should have only been one.")
		t.Fail()
	}
}
