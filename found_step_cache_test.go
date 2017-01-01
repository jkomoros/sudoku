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

	//Test removing middle step

	cache.AddStep(step)

	if cache.Len() != 3 {
		t.Error("Adding a third item didn't set length to three")
	}

	firstItem := cache.firstItem
	secondItem := cache.firstItem.next
	thirdItem := secondItem.next

	cache.remove(cache.firstItem.next)

	if cache.Len() != 2 {
		t.Error("removing middle item didn't reduce length by 1")
	}

	if firstItem.next != thirdItem {
		t.Error("Removing middle item, the first item didn't point to third")
	}

	if thirdItem.prev != firstItem {
		t.Error("Remvoing middle item, third item didn't point back to third")
	}

	secondItem = thirdItem

	//Test removing last step

	cache.AddStep(step)

	thirdItem = secondItem.next

	cache.remove(thirdItem)

	if cache.Len() != 2 {
		t.Error("Removing last item in three-item list didn't leave length 2")
	}

	if firstItem.next != secondItem {
		t.Error("Removing last item messed up first and second")
	}

	if secondItem.next != nil {
		t.Error("Removing last item didn't set up second item to point to hil")
	}

	//Test removing first step

	cache.remove(firstItem)

	if cache.firstItem != secondItem {
		t.Error("removing first item didn't leave second item as first")
	}

	if cache.firstItem.prev != nil {
		t.Error("New first item wasn't pointing to nil")
	}

}

func TestFoundStepCacheGetSteps(t *testing.T) {
	cache := &foundStepCache{}

	stepOne := &SolveStep{
		TargetCells: []CellRef{
			{1, 0},
		},
	}
	stepTwo := &SolveStep{
		TargetCells: []CellRef{
			{2, 0},
		},
	}
	stepThree := &SolveStep{
		TargetCells: []CellRef{
			{3, 0},
		},
	}

	if cache.GetSteps() != nil {
		t.Error("GetSteps on empty cache gave non-nil result")
	}

	cache.AddStep(stepOne)
	cache.AddStep(stepTwo)
	cache.AddStep(stepThree)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
		stepTwo,
		stepThree,
	})

	cache.remove(cache.firstItem.next)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
		stepThree,
	})

	cache.remove(cache.firstItem)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepThree,
	})

	cache.AddStep(stepOne)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepThree,
		stepOne,
	})

}

func getStepsHelper(t *testing.T, result []*SolveStep, golden []*SolveStep) {
	if len(result) != len(golden) {
		t.Fatal("Length mismatch. Got", len(result), "wanted", len(golden))
	}
	for i, item := range result {
		other := golden[i]

		if item != other {
			t.Error("At item", i, "got wrong item. Got", item, "wanted", other)
		}
	}

}
