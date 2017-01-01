package sudoku

import (
	"testing"
)

func TestFoundStepCacheAddStep(t *testing.T) {
	cache := &foundStepCache{}

	stepOne := &SolveStep{
		TargetCells: []CellRef{
			{0, 1},
		},
	}

	stepTwo := &SolveStep{
		TargetCells: []CellRef{
			{0, 2},
		},
	}

	stepThree := &SolveStep{
		TargetCells: []CellRef{
			{0, 3},
		},
	}

	cache.AddStep(stepOne)

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

	cache.AddStep(stepTwo)

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

	cache.AddStep(stepThree)

	if cache.Len() != 3 {
		t.Error("Adding a third item didn't set length to three")
	}

	firstItem := cache.firstItem
	secondItem := firstItem.next
	thirdItem := secondItem.next

	cache.remove(secondItem)

	if cache.Len() != 2 {
		t.Error("removing middle item didn't reduce length by 1")
	}

	if firstItem.next != thirdItem {
		t.Error("Removing middle item, the first item didn't point to third")
	}

	if thirdItem.prev != firstItem {
		t.Error("Remvoing middle item, third item didn't point back to third")
	}

	//Test removing last step

	cache.AddStep(stepTwo)

	firstItem = cache.firstItem
	secondItem = firstItem.next
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

	cache.remove(firstItem)

	if cache.firstItem != secondItem {
		t.Error("removing the first item again didn't work right")
	}

	if cache.Len() != 1 {
		t.Error("Removing an item twice left wrong length")
	}

}

func TestFoundStepCacheDuplicates(t *testing.T) {
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

	//Make sure adding the same step multipel times only adds it once
	cache.AddStep(stepOne)
	cache.AddStep(stepOne)
	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
	}, "After adding stepone twice")

	//Make sure you can add other steps
	cache.AddStep(stepTwo)
	cache.AddStep(stepThree)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepThree,
		stepTwo,
		stepOne,
	}, "After adding htree two onex2")

	//make sure we can re-add a step after it's no longer included
	cache.RemoveStepsWithCells([]CellRef{{1, 0}})

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepThree,
		stepTwo,
	}, "After removing step one")

	cache.AddStep(stepOne)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
		stepThree,
		stepTwo,
	}, "Adding one after deleting")

	//Make sure I can add to queue... but it's never actually added.
	cache.AddStepToQueue(stepOne)

	if cache.queue == nil {
		t.Error("We couldn't add a dupe to the queue")
	}

	cache.AddQueue()

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
		stepThree,
		stepTwo,
	}, "After adding queue")
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
		t.Fatal("GetSteps on empty cache gave non-nil result")
	}

	cache.AddStepToQueue(stepOne)
	cache.AddStepToQueue(stepTwo)
	cache.AddStepToQueue(stepThree)

	if cache.GetSteps() != nil {
		t.Error("GetSteps with only items in queue gave non-nil result")
	}

	if cache.Len() != 0 {
		t.Error("GetSteps with only items in queue gave non-zero len")
	}

	cache.AddQueue()

	if cache.firstItem.prev != nil {
		t.Fatal("First item's prev was not nil when added from queue")
	}

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
		stepTwo,
		stepThree,
	}, "After queue added")

	cache.remove(cache.firstItem.next)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
		stepThree,
	}, "After queue added second item removed")

	cache.remove(cache.firstItem)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepThree,
	}, "after queue added first and second items removed")

	cache.AddStep(stepOne)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
		stepThree,
	}, "after queue added first and second removed, first added again")

}

func getStepsHelper(t *testing.T, result []*SolveStep, golden []*SolveStep, description string) {
	if len(result) != len(golden) {
		t.Fatal("Length mismatch for", description, "Got", len(result), "wanted", len(golden))
	}
	for i, item := range result {
		other := golden[i]

		if item != other {
			t.Error(description, "At item", i, "got wrong item. Got", item, "wanted", other)
		}
	}

}

func TestFoundStepCacheRemoveStepsWithCells(t *testing.T) {
	cache := &foundStepCache{}

	stepOne := &SolveStep{
		TargetCells: []CellRef{
			{1, 0},
		},
	}
	stepTwo := &SolveStep{
		TargetCells: []CellRef{
			{1, 0}, {2, 0},
		},
	}
	stepThree := &SolveStep{
		PointerCells: []CellRef{
			{3, 0},
		},
	}

	cache.AddStep(stepThree)
	cache.AddStep(stepTwo)
	cache.AddStep(stepOne)

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
		stepTwo,
		stepThree,
	}, "After three steps added")

	cache.RemoveStepsWithCells([]CellRef{
		{1, 0},
	})

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepThree,
	}, "after 1,0 removed")

	cache.AddStep(stepTwo)
	cache.AddStep(stepOne)

	cache.RemoveStepsWithCells([]CellRef{
		{3, 0},
	})

	getStepsHelper(t, cache.GetSteps(), []*SolveStep{
		stepOne,
		stepTwo,
	}, "after one two added again and 3,0 removed")

}
