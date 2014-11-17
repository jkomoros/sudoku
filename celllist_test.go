package sudoku

import (
	"sort"
	"testing"
)

func TestBasicCellList(t *testing.T) {
	grid := NewGrid()
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

func TestFilledNums(t *testing.T) {
	grid := NewGrid()
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
