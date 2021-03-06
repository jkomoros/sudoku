package sudoku

import (
	"reflect"
	"sort"
	"testing"
)

func TestBasicCellSlice(t *testing.T) {
	grid := LoadSDK(SOLVED_TEST_GRID)
	row := CellSlice(grid.Row(2))
	if !row.SameRow() {
		t.Log("The items of a row were not all of the same row.")
		t.Fail()
	}

	if row.SameCol() {
		t.Log("For some reason we thought all the cells in a row were in the same col")
		t.Fail()
	}

	var refs CellRefSlice

	for i := 0; i < DIM; i++ {
		refs = append(refs, CellRef{2, i})
	}

	if !row.sameAsRefs(refs) {
		t.Error("sameAsRefs didn't match values for row 2")
	}

	col := CellSlice(grid.Col(2))
	if !col.SameCol() {
		t.Log("The items in the col were not int he same col.")
		t.Fail()
	}

	if col.SameRow() {
		t.Log("For some reason we thought all the cells in a col were in the same row")
		t.Fail()
	}

	block := CellSlice(grid.Block(2))
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

	nums := row.CollectNums(func(cell Cell) int {
		return cell.Row()
	})

	if !nums.Same() {
		t.Log("Collecting rows gave us different numbers/.")
		t.Fail()
	}

	isZeroRow := func(cell Cell) bool {
		return cell.Row() == 0
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

	description := cells.CellReferenceSlice().Description()

	if description != "(0,0), (0,1), and (0,2)" {
		t.Log("Got wrong description of cellList: ", description)
		t.Fail()
	}

	unsortedList := CellSlice{grid.Cell(0, 1), grid.Cell(1, 2), grid.Cell(0, 0)}
	unsortedList.Sort()
	if unsortedList[0].Row() != 0 || unsortedList[0].Col() != 0 ||
		unsortedList[1].Row() != 0 || unsortedList[1].Col() != 1 ||
		unsortedList[2].Row() != 1 || unsortedList[2].Col() != 2 {
		t.Error("Cell List didn't get sorted: ", unsortedList)
	}

}

type chainTestConfiguration struct {
	name                 string
	one                  CellRefSlice
	two                  CellRefSlice
	equivalentToPrevious bool
}

type chainTestResult struct {
	name             string
	value            float64
	equivalenceGroup int
}

type chainTestResults []chainTestResult

func (self chainTestResults) Len() int {
	return len(self)
}

func (self chainTestResults) Less(i, j int) bool {
	//Want it high to low
	return self[i].value > self[j].value
}

func (self chainTestResults) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func TestChainDissimilarity(t *testing.T) {

	//The first bit is where we configure the tests.
	//We should add cases here in the order of similar to dissimilar. The test will then verify
	//they come out in that order.

	//TODO: test more cases
	// For example, ones with cells that overlap.

	tests := []chainTestConfiguration{
		{
			"same row same block",
			CellRefSlice{{0, 0}},
			CellRefSlice{{0, 1}},
			false,
		},
		//this next one verifies that it doesn't matter which of self or other you do first.
		{
			"same row same block, just flipped self and other",
			CellRefSlice{{0, 1}},
			CellRefSlice{{0, 0}},
			true,
		},
		//These next two should be the same difficulty.
		{
			"same block 2 in same row 2 in same col 2 total",
			CellRefSlice{{0, 0}},
			CellRefSlice{{0, 1}, {1, 0}},
			false,
		},
		{
			"two full rows at opposite ends",
			CellRefSlice{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8}},
			CellRefSlice{{7, 0}, {7, 1}, {7, 2}, {7, 3}, {7, 4}, {7, 5}, {7, 6}, {7, 7}, {7, 8}},
			true,
		},
		{
			"same row different blocks",
			CellRefSlice{{0, 0}, {0, 1}},
			CellRefSlice{{0, 3}, {0, 4}},
			true,
		},
		{
			"same col different blocks",
			CellRefSlice{{0, 0}, {1, 0}},
			CellRefSlice{{3, 0}, {4, 0}},
			true,
		},
		{
			"same row different blocks, 2 vs 3",
			CellRefSlice{{0, 0}, {0, 1}},
			CellRefSlice{{0, 3}, {0, 4}, {0, 5}},
			true,
		},
		{
			"same block opposite corners 1 x 1",
			CellRefSlice{{0, 0}},
			CellRefSlice{{2, 2}},
			true,
		},
		{
			"adjacent rows two different blocks",
			CellRefSlice{{0, 0}, {0, 1}},
			CellRefSlice{{1, 3}, {1, 4}},
			false,
		},
		{
			"single cell opposite corners",
			CellRefSlice{{0, 0}},
			CellRefSlice{{8, 8}},
			false,
		},
	}

	//Now run the tests

	var results chainTestResults

	equivalenceGroup := -1

	for _, test := range tests {
		if !test.equivalentToPrevious {
			equivalenceGroup++
		}
		similarity := test.one.chainSimilarity(test.two)
		if similarity < 0.0 {
			t.Fatal(test.name, "failed with a dissimilarity less than 0.0: ", similarity)
		}
		if similarity > 1.0 {
			t.Fatal(test.name, "failed with a dissimilarity great than 1.0:", similarity)
		}
		result := chainTestResult{test.name, similarity, equivalenceGroup}
		results = append(results, result)
	}

	//sort them and see if their originalIndexes are now now in order.
	sort.Sort(results)

	lastEquivalenceGroup := 0
	for _, result := range results {
		if result.equivalenceGroup < lastEquivalenceGroup {
			t.Error(result.name, "was in equivalence group", result.equivalenceGroup, " but it was smaller than last group seen:", equivalenceGroup, ". Value:", result.value)
		}
		lastEquivalenceGroup = result.equivalenceGroup
	}

}

func TestFilledNums(t *testing.T) {
	grid, err := LoadSDKFromFile(puzzlePath("nakedpairblock1.sdk"))
	if err != nil {
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
	cells := grid.Row(0)

	indexes := IntSlice([]int{4, 6, 2})

	subset := cells.InverseSubset(indexes)

	if len(subset) != DIM-3 {
		t.Error("Inverse subset gave wrong number of results")
	}

	for _, cell := range subset {
		if cell.Col() == 2 || cell.Col() == 4 || cell.Col() == 6 {
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

func TestCellSetBasicOperations(t *testing.T) {
	grid := NewGrid()
	cellOne := grid.Cell(0, 1)
	cellTwo := grid.Cell(0, 2)
	cellThree := grid.Cell(0, 3)

	oneSlice := CellSlice{cellOne, cellTwo}
	twoSlice := CellSlice{cellThree}
	threeSlice := CellSlice{cellTwo, cellThree}

	one := oneSlice.toCellSet()
	two := twoSlice.toCellSet()
	three := threeSlice.toCellSet()

	oneGolden := cellSet{cellOne.Reference(): true, cellTwo.Reference(): true}
	twoGolden := cellSet{cellThree.Reference(): true}
	threeGolden := cellSet{cellTwo.Reference(): true, cellThree.Reference(): true}

	if !reflect.DeepEqual(one, oneGolden) {
		t.Fatal("Creating cellSet failed. Got", one, "wanted", oneGolden)
	}

	if !reflect.DeepEqual(two, twoGolden) {
		t.Fatal("Creating cellset two failed. Got: ", two, "wanted", twoGolden)
	}

	if !reflect.DeepEqual(three, threeGolden) {
		t.Fatal("Creating cellset three failed. Got: ", three, "wanted", threeGolden)
	}

	oneTwoIntersection := one.intersection(two)

	if !reflect.DeepEqual(oneTwoIntersection, cellSet{}) {
		t.Error("One two intersection failed. Got:", oneTwoIntersection)
	}

	oneThreeIntersection := one.intersection(three)

	if !reflect.DeepEqual(oneThreeIntersection, cellSet{cellTwo.Reference(): true}) {
		t.Error("One three intersection failed. Got: ", oneThreeIntersection)
	}

	oneTwoUnion := one.union(two)

	if !reflect.DeepEqual(oneTwoUnion, cellSet{cellOne.Reference(): true, cellTwo.Reference(): true, cellThree.Reference(): true}) {
		t.Error("One two union failed. Got: ", oneTwoUnion)
	}

	oneThreeUnion := one.union(three)

	if !reflect.DeepEqual(oneThreeUnion, cellSet{cellOne.Reference(): true, cellTwo.Reference(): true, cellThree.Reference(): true}) {
		t.Error("One three union failed. Got: ", oneThreeUnion)
	}

	oneTwoDifference := one.difference(two)

	if !reflect.DeepEqual(oneTwoDifference, one) {
		t.Error("One two difference failed. Got: ", oneTwoDifference)
	}

	oneThreeDifference := one.difference(three)

	if !reflect.DeepEqual(oneThreeDifference, cellSet{cellOne.Reference(): true}) {
		t.Error("One three difference failed. Got: ", oneThreeDifference)
	}

}
