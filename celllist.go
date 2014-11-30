package sudoku

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

type CellList []*Cell

type IntSlice []int

type stringSlice []string

type intSet map[int]bool

type cellRef struct {
	row int
	col int
}

type cellListSorter struct {
	CellList
}

func getRow(cell *Cell) int {
	return cell.Row
}

func getCol(cell *Cell) int {
	return cell.Col
}

func getBlock(cell *Cell) int {
	return cell.Block
}

func (self CellList) SameRow() bool {
	return self.CollectNums(getRow).Same()
}

func (self CellList) SameCol() bool {
	return self.CollectNums(getCol).Same()
}

func (self CellList) SameBlock() bool {
	return self.CollectNums(getBlock).Same()
}

func (self CellList) Row() int {
	//Will return the row of a random item.
	if len(self) == 0 {
		return 0
	}
	return self[0].Row
}

func (self CellList) Col() int {
	if len(self) == 0 {
		return 0
	}
	return self[0].Col
}

func (self CellList) Block() int {
	if len(self) == 0 {
		return 0
	}
	return self[0].Block
}

func (self CellList) AddExclude(exclude int) {
	mapper := func(cell *Cell) {
		cell.setExcluded(exclude, true)
	}
	self.Map(mapper)
}

func (self CellList) FilterByPossible(possible int) CellList {
	//TODO: test this
	filter := func(cell *Cell) bool {
		return cell.Possible(possible)
	}
	return self.Filter(filter)
}

func (self CellList) FilterByNumPossibilities(target int) CellList {
	//TODO: test this
	filter := func(cell *Cell) bool {
		return len(cell.Possibilities()) == target
	}
	return self.Filter(filter)
}

func (self CellList) FilterByHasPossibilities() CellList {
	//Returns a list of cells that have possibilities.
	//TODO: test this.
	filter := func(cell *Cell) bool {
		return len(cell.Possibilities()) > 0
	}
	return self.Filter(filter)
}

func (self CellList) RemoveCells(targets CellList) CellList {
	//TODO: test this.
	targetCells := make(map[*Cell]bool)
	for _, cell := range targets {
		targetCells[cell] = true
	}
	filterFunc := func(cell *Cell) bool {
		return !targetCells[cell]
	}
	return self.Filter(filterFunc)
}

func (self CellList) PossibilitiesUnion() IntSlice {
	//Returns an IntSlice of the union of all possibilities.
	set := make(map[int]bool)

	for _, cell := range self {
		for _, possibility := range cell.Possibilities() {
			set[possibility] = true
		}
	}

	result := make(IntSlice, len(set))

	i := 0
	for possibility, _ := range set {
		result[i] = possibility
		i++
	}

	return result
}

func (self CellList) Subset(indexes IntSlice) CellList {
	//IntSlice.Subset is basically a carbon copy.
	//TODO: what's this behavior if indexes has dupes? What SHOULD it be?
	result := make(CellList, len(indexes))
	max := len(self)
	for i, index := range indexes {
		if index >= max {
			//This probably is indicative of a larger problem.
			continue
		}
		result[i] = self[index]
	}
	return result
}

func (self CellList) InverseSubset(indexes IntSlice) CellList {
	//TODO: figure out what this should do when presented with dupes.

	//LIke Subset, but returns all of the items NOT called out in indexes.
	var result CellList

	//Ensure indexes are in sorted order.
	sort.Ints(indexes)

	//Index into indexes we're considering
	currentIndex := 0

	for i := 0; i < len(self); i++ {
		if currentIndex < len(indexes) && i == indexes[currentIndex] {
			//Skip it!
			currentIndex++
		} else {
			//Output it!
			result = append(result, self[i])
		}
	}

	return result
}

func (self CellList) Sort() {
	sorter := cellListSorter{self}
	sort.Sort(sorter)
}

func (self CellList) FilledNums() IntSlice {
	set := make(intSet)
	for _, cell := range self {
		if cell.Number() == 0 {
			continue
		}
		set[cell.Number()] = true
	}
	return set.toSlice()
}

func (self CellList) CollectNums(fetcher func(*Cell) int) IntSlice {
	var result IntSlice
	for _, cell := range self {
		result = append(result, fetcher(cell))
	}
	return result
}

func (self cellListSorter) Len() int {
	return len(self.CellList)
}

func (self cellListSorter) Less(i, j int) bool {
	//Sort based on the index of the cell.
	one := self.CellList[i]
	two := self.CellList[j]

	return (one.Row*DIM + one.Col) < (two.Row*DIM + two.Col)
}

func (self cellListSorter) Swap(i, j int) {
	self.CellList[i], self.CellList[j] = self.CellList[j], self.CellList[i]
}

func (self CellList) Filter(filter func(*Cell) bool) CellList {
	var result CellList
	for _, cell := range self {
		if filter(cell) {
			result = append(result, cell)
		}
	}
	return result
}

func (self CellList) Map(mapper func(*Cell)) {
	for _, cell := range self {
		mapper(cell)
	}
}

//TODO: should this be in this file? It's awfully specific to HumanSolve needs, and extremely complex.
//TODO: is this how you spell this?
func (self CellList) ChainDissimilarity(other CellList) float64 {
	//Returns a value between 0.0 and 1.0 depending on how 'similar' the cellLists are.

	//Note: it doesn't ACTUALLY guarantee a value lower than 1.0 (it might be possible to hit those; reasoning about the maximum value is tricky).

	//Note: a 0.0 means extremely similar, and 1.0 means extremely dissimilar. (This is natural because HumanSolve wnats invertedWeights)

	//Similarity, here, does not mean the overlap of cells that are in both sets--it means how related the blocks/rows/groups are
	//to one another. This is used in HumanSolve to boost the likelihood of picking steps that are some how 'chained' to the step
	//before them.
	//An example with a very high similarity would be if two cells in a row in a block were in self, and other consisted of a DIFFERENT cell
	//in the same row in the same block.
	//An example with a very low similarity would be cells that don't share any of the same row/col/block.

	//The overall approach is for self to create three []float64 of DIM length, one for row,col, block id's. Then, go through
	//And record the proprotion of the targetCells that fell in that group.
	//Then, you do the same for other.
	//Then, you sum up the differences in all of the vectors and record a diff for row, block, and col.
	//Then, you sort the diffs so that the one with the lowest is weighted at 4, 2, 1. This last bit captures the fact that if they're
	//all in the same row (but different columns) that's still quite good.
	//Then, we normalize the result based on the highest and lowest possible scores.

	selfRow := make([]float64, DIM)
	selfCol := make([]float64, DIM)
	selfBlock := make([]float64, DIM)

	otherRow := make([]float64, DIM)
	otherCol := make([]float64, DIM)
	otherBlock := make([]float64, DIM)

	//How much to add each time we find a cell with that row/col/block.
	//This saves us from having to loop through again to compute the average
	selfProportion := float64(1) / float64(len(self))
	otherProportion := float64(1) / float64(len(other))

	for _, cell := range self {
		selfRow[cell.Row] += selfProportion
		selfCol[cell.Col] += selfProportion
		selfBlock[cell.Block] += selfProportion
	}

	for _, cell := range other {
		otherRow[cell.Row] += otherProportion
		otherCol[cell.Col] += otherProportion
		otherBlock[cell.Block] += otherProportion
	}

	rowDiff := 0.0
	colDiff := 0.0
	blockDiff := 0.0

	//Now, compute the diffs.
	for i := 0; i < DIM; i++ {
		rowDiff += math.Abs(selfRow[i] - otherRow[i])
		colDiff += math.Abs(selfCol[i] - otherCol[i])
		blockDiff += math.Abs(selfBlock[i] - otherBlock[i])
	}

	//Now sort the diffs; we care disproportionately about the one that matches best.
	diffs := []float64{rowDiff, colDiff, blockDiff}
	sort.Float64s(diffs)

	//We care about the lowest diff the most (capturing the notion that if they line up in row but nothing else, that's still quite good!)
	weights := []int{4, 2, 1}

	result := 0.0

	for i := 0; i < 3; i++ {
		for j := 0; j < weights[i]; j++ {
			result += diffs[i]
		}
	}

	//Divide by 4 + 2 + 1 = 7
	result /= 7.0

	return result

}

func (self CellList) Description() string {
	strings := make(stringSlice, len(self))

	for i, cell := range self {
		strings[i] = cell.ref().String()
	}

	return strings.description()
}

func (self CellList) sameAsRefs(refs []cellRef) bool {
	cellSet := make(map[string]bool)
	for _, cell := range self {
		cellSet[cell.ref().String()] = true
	}

	refSet := make(map[string]bool)
	for _, ref := range refs {
		refSet[ref.String()] = true
	}

	if len(cellSet) != len(refSet) {
		return false
	}

	for item, _ := range cellSet {
		if _, ok := refSet[item]; !ok {
			return false
		}
	}

	return true
}

func (self cellRef) Cell(grid *Grid) *Cell {
	if grid == nil {
		return nil
	}
	return grid.Cell(self.row, self.col)
}

func (self cellRef) String() string {
	return "(" + strconv.Itoa(self.row) + "," + strconv.Itoa(self.col) + ")"
}

func (self stringSlice) description() string {
	if len(self) == 0 {
		return ""
	}

	if len(self) == 1 {
		return self[0]
	}

	if len(self) == 2 {
		return self[0] + " and " + self[1]
	}

	result := strings.Join(self[:len(self)-1], ", ")

	return result + ", and " + self[len(self)-1]
}

func (self IntSlice) Description() string {

	strings := make(stringSlice, len(self))

	for i, num := range self {
		strings[i] = strconv.Itoa(num)
	}

	return strings.description()

}

//returns an IntSlice like self, but with any dupes removed.
func (self IntSlice) Unique() IntSlice {
	//TODO: test this.
	return self.toIntSet().toSlice()
}

func (self IntSlice) Same() bool {
	if len(self) == 0 {
		return true
	}
	target := self[0]
	for _, num := range self {
		if target != num {
			return false
		}
	}
	return true
}

func (self IntSlice) SameContentAs(otherSlice IntSlice) bool {
	//Same as SameAs, but doesn't care about order.

	//TODO: impelement this using intSets. It's easier.

	selfToUse := make(IntSlice, len(self))
	copy(selfToUse, self)
	sort.IntSlice(selfToUse).Sort()

	otherToUse := make(IntSlice, len(otherSlice))
	copy(otherToUse, otherSlice)
	sort.IntSlice(otherToUse).Sort()

	return selfToUse.SameAs(otherToUse)
}

func (self IntSlice) SameAs(other IntSlice) bool {
	//TODO: test this.
	if len(self) != len(other) {
		return false
	}
	for i, num := range self {
		if other[i] != num {
			return false
		}
	}
	return true
}

func (self IntSlice) Subset(indexes IntSlice) IntSlice {
	//TODO: test this.
	//Basically a carbon copy of CellList.Subset
	//TODO: what's this behavior if indexes has dupes? What SHOULD it be?
	result := make(IntSlice, len(indexes))
	max := len(self)
	for i, index := range indexes {
		if index >= max {
			//This probably is indicative of a larger problem.
			continue
		}
		result[i] = self[index]
	}
	return result
}

func (self IntSlice) Sort() {
	sort.Ints(self)
}

func (self IntSlice) toIntSet() intSet {
	result := make(intSet)
	for _, item := range self {
		result[item] = true
	}
	return result
}

func (self intSet) toSlice() IntSlice {
	var result IntSlice
	for item, val := range self {
		if val {
			result = append(result, item)
		}
	}
	return result
}

//TODO: test this directly (tested implicitly via intSlice.Intersection)
func (self intSet) intersection(other intSet) intSet {
	result := make(intSet)
	for item, value := range self {
		if value {
			if val, ok := other[item]; ok && val {
				result[item] = true
			}
		}
	}
	return result
}

func (self intSet) difference(other intSet) intSet {
	result := make(intSet)
	for item, value := range self {
		if value {
			if val, ok := other[item]; !ok && !val {
				result[item] = true
			}
		}
	}
	return result
}

//TODO: test this.
func (self intSet) union(other intSet) intSet {
	result := make(intSet)
	for item, value := range self {
		result[item] = value
	}
	for item, value := range other {
		result[item] = value
	}
	return result
}

func (self IntSlice) Intersection(other IntSlice) IntSlice {
	//Returns an IntSlice of the union of both intSlices

	return self.toIntSet().intersection(other.toIntSet()).toSlice()
}

func (self IntSlice) Difference(other IntSlice) IntSlice {
	return self.toIntSet().difference(other.toIntSet()).toSlice()
}
