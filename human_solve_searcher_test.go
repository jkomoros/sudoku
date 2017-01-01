package sudoku

import (
	"testing"
)

func TestHumanSolveSearcher(t *testing.T) {

	grid := MutableLoadSDK(TEST_GRID)

	//TODO: test that when we pass in stepsCache it works as intended.
	searcher := newHumanSolveSearcher(grid, nil, DefaultHumanSolveOptions(), nil)

	if searcher.itemsToExplore.Len() != 1 {
		t.Error("Expected new frontier to have exactly one item in it, but got", searcher.itemsToExplore.Len())
	}

	if searcher.grid == nil {
		t.Error("No grid in frontier")
	}

	basePotentialNextStep := searcher.NextPossibleStep()

	if basePotentialNextStep == nil {
		t.Error("Didn'get base potential next step")
	}

	baseGrid := basePotentialNextStep.Grid()

	if baseGrid.DataString() != grid.DataString() {
		t.Error("the grid in the base item in the frontier was not right. Got", baseGrid.DataString(), "wanted", grid.DataString())
	}

	if searcher.itemsToExplore.Len() != 0 {
		t.Error("Getting the base potential next step should have emptied it, but len is", searcher.itemsToExplore.Len())
	}

	nInRowTechnique := techniquesByName["Necessary In Row"]

	simpleFillStep := &SolveStep{
		Technique: nInRowTechnique,
		TargetCells: CellRefSlice{
			CellRef{0, 0},
		},
		TargetNums: IntSlice{1},
	}

	if simpleFillStep.Technique == nil {
		t.Fatal("couldn't find necessary in row technique")
	}

	simpleFillStepItem := basePotentialNextStep.AddStep(simpleFillStep)

	if simpleFillStepItem == nil {
		t.Fatal("Adding fill step didn't return anything")
	}

	if simpleFillStepItem.heapIndex != -1 {
		t.Fatal("Adding completed item to frontier didn't have -1 index")
	}

	if len(searcher.completedItems) != 1 {
		t.Error("Expected the completed item to go into CompletedItems,but it's empty")
	}

	if len(searcher.itemsToExplore) != 0 {
		t.Error("Expected the completed item to go into COmpletedItems, but it apparently went into items.")
	}

	//This is a fragile way to test this; it will need to be updated every
	//time we change the twiddlers. :-(
	expectedGoodness := 28.907004830917874

	if simpleFillStepItem.Goodness() != expectedGoodness {
		t.Error("Goodness of simple fill step was wrong. Execpted", expectedGoodness, "got", simpleFillStepItem.Goodness(), simpleFillStepItem.explainGoodness())
	}

	cell := simpleFillStepItem.Grid().Cell(0, 0)

	if cell.Number() != 1 {
		t.Error("Cell in grid was not set correctly. Got", cell.Number(), "wanted 1")
	}

	nonFillStep := &SolveStep{
		Technique: techniquesByName["Pointing Pair Row"],
		TargetCells: CellRefSlice{
			CellRef{0, 1},
		},
		TargetNums: IntSlice{2},
	}

	if nonFillStep.Technique == nil {
		t.Fatal("Couldn't find pointing pair row techhnique")
	}

	nonFillStepItem := basePotentialNextStep.AddStep(nonFillStep)

	if nonFillStepItem == nil {
		t.Fatal("Adding non fill step didn't return a frontier object")
	}

	if searcher.itemsToExplore.Len() != 1 {
		t.Error("Frontier had wrong length after adding one complete and one incomplete items. Got", searcher.itemsToExplore.Len(), "expected 1")
	}

	if searcher.itemsToExplore[0] != nonFillStepItem {
		t.Error("We though that simpleFillStep should be at the end of the queue but it wasn't.")
	}

	expensiveStep := &SolveStep{
		Technique: techniquesByName["Hidden Quad Block"],
		TargetCells: CellRefSlice{
			CellRef{0, 2},
		},
		TargetNums: IntSlice{3},
	}

	expensiveStepItem := basePotentialNextStep.AddStep(expensiveStep)

	if searcher.itemsToExplore.Len() != 2 {
		t.Error("Wrong length after adding two items to frontier. Got", searcher.itemsToExplore.Len(), "expected 2")
	}

	if searcher.itemsToExplore[0] != nonFillStepItem {
		t.Error("We expected the expensive step to be worse", searcher.String())
	}

	//We'll have to twiddle down by the goodness...

	goodness := expensiveStepItem.Goodness()

	//I apologize to the programming gods for the next 5 lines of unbelievable
	//hackiness...

	//Horrendous hack to allow us to twiddle again
	expensiveStepItem.doneTwiddling = false
	expensiveStepItem.cachedGoodness = 0.0
	//Twiddle will reject negative amounts, so we have to (ugh) do most of the item.Twiddle ourselves...
	expensiveStepItem.twiddles = append(expensiveStepItem.twiddles, twiddleRecord{"Very small amount to make this #1", (probabilityTweak(goodness) - 0.00001) * -1.0})
	expensiveStepItem.searcher.ItemValueChanged(expensiveStepItem)

	if searcher.itemsToExplore[0] != expensiveStepItem {
		t.Error("Even after twiddling up guess step by a lot it still wasn't in the top position in frontier", searcher.itemsToExplore[0], searcher.itemsToExplore[1])
	}

	poppedItem := searcher.NextPossibleStep()

	if poppedItem != expensiveStepItem {
		t.Error("Expected popped item to be the non-fill step now that its goodness is higher, but got", poppedItem)
	}

	if searcher.itemsToExplore.Len() != 1 {
		t.Error("Wrong frontier length after popping item. Got", searcher.itemsToExplore.Len(), "expected 1")
	}

	poppedItem = searcher.NextPossibleStep()
	//Should be nonFillStepItem

	currentGoodness := nonFillStepItem.Goodness()

	completedNonFillStemItem := nonFillStepItem.AddStep(simpleFillStep)

	if completedNonFillStemItem.Goodness() == currentGoodness {
		t.Error("Adding a step to end of nonfill step didn't change goodness.")
	}

	if searcher.itemsToExplore.Len() != 0 {
		t.Error("Adding an item gave wrong len. Got", searcher.itemsToExplore.Len(), "wanted 0")
	}

	if len(searcher.completedItems) != 2 {
		t.Error("Got wrong number of completed items. Got", len(searcher.completedItems), "expected 2")
	}

	steps := completedNonFillStemItem.Steps()

	if len(steps) != 2 {
		t.Error("Expected two steps back, got", len(steps))
	}

	if steps[0] != nonFillStepItem.step {
		t.Error("Expected first step to be the step of nonFillStepItem. Got", steps[0])
	}

}
