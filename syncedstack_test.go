package sudoku

import (
	"testing"
	"time"
)

func TestBasicSyncedStack(t *testing.T) {
	stack := newSyncedStack[map[string]int]()
	if stack == nil {
		t.Log("Didn't get a stack back.")
		t.Fail()
	}
	if stack.Length() != 0 {
		t.Log("A new stack did not have a length of one")
		t.Fail()
	}

	if !stack.IsDone() {
		t.Log("Stack doesn't think it's done when nothing has happened yet")
		t.Fail()
	}

	item := map[string]int{"a": 1}
	secondItem := map[string]int{"b": 2}
	stack.Insert(item)
	if stack.Length() != 1 {
		t.Log("We inserted an item but the length did not go up by one.")
		t.Fail()
	}

	if stack.IsDone() {
		t.Log("Stack thinks it's done but it has an item.")
		t.Fail()
	}

	result := stack.Pop()
	if result == nil {
		t.Log("We didn't get back an item from a queue with one item.")
		t.Fail()
	}
	if _, ok := result["a"]; !ok {
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

	if stack.IsDone() {
		t.Log("Stack thinks it's done but we didn't tell it we were done processing one item.")
		t.Fail()
	}

	stack.ItemDone()

	if !stack.IsDone() {
		t.Log("Stack doesn't think it's done even after we told it we processed an item.")
		t.Fail()
	}

	stack.Insert(item)
	stack.Insert(secondItem)
	result = stack.Pop()
	if result == nil {
		t.Log("We didn't get back an item from a queue with two items")
		t.Fail()
	}
	if _, ok := result["b"]; !ok {
		t.Log("We got the wrong item back from a two item queue.")
		t.Fail()
	}
	stack.Insert(secondItem)
	//This should always be the last item.
	result = stack.Get(0.0)
	if result == nil {
		t.Log("We didn't get back the first item with probability 0")
		t.Fail()
	}
	if _, ok := result["a"]; !ok {
		t.Log("We didn't get back the first item")
		t.Fail()
	}
	if stack.Length() != 1 {
		t.Log("We got back wrong length for a queue with one item")
		t.Fail()
	}
	result = stack.Get(0.0)
	if result == nil {
		t.Log("We didn't get back the only item with probability 0.0")
		t.Fail()
	}
	if _, ok := result["b"]; !ok {
		t.Log("We got the wrong item out of a queue with one item")
		t.Fail()
	}
}

func TestChanSyncedStack(t *testing.T) {
	doneChan := make(chan bool, 1)
	stack := newChanSyncedStack[int](doneChan)
	item := 1
	secondItem := 2
	var result int
	select {
	case <-stack.Output:
		t.Log("We got something on output before there was anything to get.")
		t.Fail()
	default:
		//Fine
	}

	stack.Insert(item)
	stack.Insert(secondItem)

	if stack.Pop() != 0 {
		t.Log("We were able to get something using Get")
		t.Fail()
	}

	select {
	case result = <-stack.Output:
		if result != 2 && result != 1 {
			t.Log("We got the wrong item out of the queue")
			t.Fail()
		}
	default:
		t.Log("We didn't get anything out of the queue but we should have.")
		t.Fail()
	}

	select {
	case result = <-stack.Output:
		if result != 2 && result != 1 {
			t.Log("We got the wrong item out of the queue the second time.")
			t.Fail()
		}
	}

	stack.ItemDone()
	stack.ItemDone()

	select {
	case _, ok := <-stack.Output:
		if ok {
			t.Log("We were still able to receive on what should have been a closed channel")
			t.Fail()
		}
	default:
		//Meh, that's fine.
	}

	select {
	case <-doneChan:
		//good
	case <-time.After(10000):
		t.Log("We didn't get anything on the done channel after awhile.")
		t.Fail()
	}
}
