package sudoku

import (
	"github.com/davecgh/go-spew/spew"
	"sort"
	"testing"
)

func TestBasicCellList(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	grid.Load(SOLVED_TEST_GRID)
	row := CellList(grid.Row(2))
	if !row.SameRow() {
		t.Log("The items of a row were not all of the same row.")
		t.Fail()
	}

	if row.SameCol() {
		t.Log("For some reason we thought all the cells in a row were in the same col")
		t.Fail()
	}

	var refs []cellRef

	for i := 0; i < DIM; i++ {
		refs = append(refs, cellRef{2, i})
	}

	if !row.sameAsRefs(refs) {
		t.Error("sameAsRefs didn't match values for row 2")
	}

	col := CellList(grid.Col(2))
	if !col.SameCol() {
		t.Log("The items in the col were not int he same col.")
		t.Fail()
	}

	if col.SameRow() {
		t.Log("For some reason we thought all the cells in a col were in the same row")
		t.Fail()
	}

	block := CellList(grid.Block(2))
	if !block.SameBlock() {
		t.Log("The items in the block were not int he same block.")
		t.Fail()
	}

	if block.SameRow() {
		t.Log("For some reason we thought all the cells in a col were in the same row")
		t.Fail()
	}

	if block.SameCol() {
		t.Log("For some reason we thought all the cells in a block were in the same col")
		t.Fail()
	}

	nums := row.CollectNums(func(cell *Cell) int {
		return cell.Row
	})

	if !nums.Same() {
		t.Log("Collecting rows gave us different numbers/.")
		t.Fail()
	}

	isZeroRow := func(cell *Cell) bool {
		return cell.Row == 0
	}

	cells := grid.Block(0).Filter(isZeroRow)

	if len(cells) != BLOCK_DIM {
		t.Log("We got back the wrong number of cells when filtering")
		t.Fail()
	}

	if !cells.SameRow() {
		t.Log("We got back cells not inthe same row.")
		t.Fail()
	}

	description := cells.Description()

	if description != "(0,0), (0,1), and (0,2)" {
		t.Log("Got wrong description of cellList: ", description)
		t.Fail()
	}

	unsortedList := CellList{grid.Cell(0, 1), grid.Cell(1, 2), grid.Cell(0, 0)}
	unsortedList.Sort()
	if unsortedList[0].Row != 0 || unsortedList[0].Col != 0 ||
		unsortedList[1].Row != 0 || unsortedList[1].Col != 1 ||
		unsortedList[2].Row != 1 || unsortedList[2].Col != 2 {
		t.Error("Cell List didn't get sorted: ", unsortedList)
	}

}

type chainTestConfiguration struct {
	name string
	one  []cellRef
	two  []cellRef
}

type chainTestResult struct {
	name          string
	originalIndex int
	value         float64
}

type chainTestResults []chainTestResult

func (self chainTestResults) Len() int {
	return len(self)
}

func (self chainTestResults) Less(i, j int) bool {
	return self[i].value < self[j].value
}

func (self chainTestResults) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func TestChainDissimilarity(t *testing.T) {

	//The first bit is where we configure the tests.
	//We should add cases here in the order of similar to dissimilar. The test will then verify
	//they come out in that order.

	tests := []chainTestConfiguration{
		{
			"same row same block",
			[]cellRef{{0, 0}},
			[]cellRef{{0, 1}},
		},
		//this next one verifies that it doesn't matter which of self or other you do first.
		{
			"same row same block, just flipped self and other",
			[]cellRef{{0, 1}},
			[]cellRef{{0, 0}},
		},
		//These next two should be the same difficulty.
		//TODO: might need to generalize the test to allow me to say
		//that two can be equivalent.
		{
			"same block 2 in same row 2 in same col 2 total",
			[]cellRef{{0, 0}},
			[]cellRef{{0, 1}, {1, 0}},
		},
		{
			"two full rows at opposite ends",
			[]cellRef{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8}},
			[]cellRef{{7, 0}, {7, 1}, {7, 2}, {7, 3}, {7, 4}, {7, 5}, {7, 6}, {7, 7}, {7, 8}},
		},
		{
			"same row different blocks",
			[]cellRef{{0, 0}, {0, 1}},
			[]cellRef{{0, 3}, {0, 4}},
		},
		{
			"same col different blocks",
			[]cellRef{{0, 0}, {1, 0}},
			[]cellRef{{3, 0}, {4, 0}},
		},
		{
			"same row different blocks, 2 vs 3",
			[]cellRef{{0, 0}, {0, 1}},
			[]cellRef{{0, 3}, {0, 4}, {0, 5}},
		},
		{
			"single cell opposite corners",
			[]cellRef{{0, 0}},
			[]cellRef{{8, 8}},
		},
	}

	//Now run the tests

	grid := NewGrid()

	var results chainTestResults

	for i, test := range tests {
		var listOne CellList
		var listTwo CellList
		for _, ref := range test.one {
			listOne = append(listOne, ref.Cell(grid))
		}
		for _, ref := range test.two {
			listTwo = append(listTwo, ref.Cell(grid))
		}
		dissimilarity := listOne.ChainDissimilarity(listTwo)
		if dissimilarity < 0.0 {
			t.Fatal(test.name, "failed with a dissimilarity less than 0.0: ", dissimilarity)
		}
		if dissimilarity > 1.0 {
			t.Fatal(test.name, "failed with a dissimilarity great than 1.0:", dissimilarity)
		}
		result := chainTestResult{test.name, i, dissimilarity}
		results = append(results, result)
	}

	//sort them and see if their originalIndexes are now now in order.
	sort.Sort(results)

	spew.Dump(results)

	for i, result := range results {
		if result.originalIndex != i {
			t.Error(result.name, "was in position", i, " but it was supposed to be in position", result.originalIndex, ". Value:", result.value)
		}
	}

}

func TestFilledNums(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	if !grid.LoadFromFile(puzzlePath("nakedpairblock1.sdk")) {
		t.Fatal("Couldn't load file")
	}

	filledNums := grid.Row(0).FilledNums()

	if !filledNums.SameContentAs(IntSlice{3, 7, 8, 9}) {
		t.Error("Filled nums had wrong nums", filledNums)
	}

}

func TestIntList(t *testing.T) {
	numArr := [...]int{1, 1, 1}
	if !IntSlice(numArr[:]).Same() {
		t.Log("We didn't think that a num list with all of the same ints was the same.")
		t.Fail()
	}
	differentNumArr := [...]int{1, 2, 1}
	if IntSlice(differentNumArr[:]).Same() {
		t.Log("We thought a list of different ints were the same")
		t.Fail()
	}
	description := IntSlice(numArr[:]).Description()
	if description != "1, 1, and 1" {
		t.Log("Didn't get right description: ", description)
		t.Fail()
	}

	unsortedList := IntSlice{3, 2, 1}
	unsortedList.Sort()
	if !unsortedList.SameAs(IntSlice{1, 2, 3}) {
		t.Error("IntSlice.Sort did not sort the list.")
	}

	oneList := IntSlice{1}

	description = oneList.Description()

	if description != "1" {
		t.Log("Didn't get the right description for a short intlist: ", description)
		t.Fail()
	}

	twoList := IntSlice{1, 1}

	description = twoList.Description()

	if description != "1 and 1" {
		t.Log("Did'get the the right description for a two-item intList: ", description)
		t.Fail()
	}
}

func TestInverseSubset(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	cells := grid.Row(0)

	indexes := IntSlice([]int{4, 6, 2})

	subset := cells.InverseSubset(indexes)

	if len(subset) != DIM-3 {
		t.Error("Inverse subset gave wrong number of results")
	}

	for _, cell := range subset {
		if cell.Col == 2 || cell.Col == 4 || cell.Col == 6 {
			t.Error("Inverse subset included cells it shouldn't have.")
		}
	}

}

func TestIntSliceIntersection(t *testing.T) {
	one := IntSlice([]int{1, 3, 2, 5})
	two := IntSlice([]int{2, 7, 6, 5})

	result := one.Intersection(two)

	if len(result) != 2 {
		t.Error("Intersection had wrong number of items")
	}

	sort.Ints(result)

	if result[0] != 2 || result[1] != 5 {
		t.Error("Intersection result was wrong.")
	}
}

func TestIntSliceDifference(t *testing.T) {
	one := IntSlice([]int{1, 2, 3, 4, 5, 6})
	two := IntSlice([]int{3, 4, 7})

	result := one.Difference(two)

	if !result.SameContentAs(IntSlice([]int{1, 2, 5, 6})) {
		t.Error("Int slice difference gave wrong result: ", result)
	}
}

func TestSameContentAs(t *testing.T) {
	one := IntSlice([]int{2, 3, 1})
	two := IntSlice([]int{2, 1, 3})

	if !one.SameContentAs(two) {
		t.Log("Didn't think two equivalent slices were the same.")
		t.Fail()
	}

	if !one.SameAs([]int{2, 3, 1}) {
		t.Log("We mutated one")
		t.Fail()
	}

	if !two.SameAs([]int{2, 1, 3}) {
		t.Log("We mutated two")
		t.Fail()
	}

	onePair := IntSlice([]int{3, 2})
	twoPair := IntSlice([]int{2, 3})

	if !onePair.SameContentAs(twoPair) {
		t.Log("Didn't think two equivalent pairs were the same.")
		t.Fail()
	}

}
