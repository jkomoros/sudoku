package sudoku

import (
	"testing"
)

func TestFoundStepCacheAddStep(t *testing.T) {
	cache := &foundStepCache{}

	step := &SolveStep{}

	cache.AddStep(step)

	if cache.Len() != 1 {
		t.Error("Added an item to cache but length wasn't 1")
	}

	if cache.firstItem == nil {
		t.Fatal("Added an item to cache but firstItem wasn't set")
	}

	if cache.firstItem.next != nil {
		t.Error("First cache item's next was not nil")
	}

	if cache.firstItem.prev != nil {
		t.Error("First cache item's prev was not nil")
	}

	cache.AddStep(step)

	if cache.Len() != 2 {
		t.Error("Added another item to the cache but length wasn't 2")
	}

	if cache.firstItem.next == nil {
		t.Fatal("After adding a second item the firstItem's next was not set")
	}

	if cache.firstItem.next.prev != cache.firstItem {
		t.Error("The second cache item doesn't point back to first cache item")
	}

	if cache.firstItem.next.next != nil {
		t.Error("The second cache item points on to another item but it shouldn't")
	}

}
