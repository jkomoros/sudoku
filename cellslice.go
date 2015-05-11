package sudoku

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

//CellSlice is a list of cells with many convenience methods for doing common operations on them.
type CellSlice []*Cell

//IntSlice is a list of ints, with many convenience methods specific to sudoku.
type IntSlice []int

type stringSlice []string

type intSet map[int]bool

type cellSet map[*Cell]bool

type cellRef struct {
	row int
	col int
}

type CellSliceSorter struct {
	CellSlice
}

func getRow(cell *Cell) int {
	return cell.Row()
}

func getCol(cell *Cell) int {
	return cell.Col()
}

func getBlock(cell *Cell) int {
	return cell.Block()
}

//SameRow returns true if all cells are in the same row.
func (self CellSlice) SameRow() bool {
	return self.CollectNums(getRow).Same()
}

//SameCol returns true if all cells are in the same column.
func (self CellSlice) SameCol() bool {
	return self.CollectNums(getCol).Same()
}

//SameBlock returns true if all cells are in the same block.
func (self CellSlice) SameBlock() bool {
	return self.CollectNums(getBlock).Same()
}

//Row returns the row that at least one of the cells is in. If SameRow() is false, the Row
//may be any of the rows in the set.
func (self CellSlice) Row() int {
	//Will return the row of a random item.
	if len(self) == 0 {
		return 0
	}
	return self[0].Row()
}

//Col returns the column that at least one of the cells is in. If SameCol() is false, the column
//may be any of the columns in the set.
func (self CellSlice) Col() int {
	if len(self) == 0 {
		return 0
	}
	return self[0].Col()
}

//Block returns the row that at least one of the cells is in. If SameBlock() is false, the Block
//may be any of the blocks in the set.
func (self CellSlice) Block() int {
	if len(self) == 0 {
		return 0
	}
	return self[0].Block()
}

//AddExclude sets the given number to excluded on all cells in the set.
func (self CellSlice) AddExclude(exclude int) {
	mapper := func(cell *Cell) {
		cell.SetExcluded(exclude, true)
	}
	self.Map(mapper)
}

//FilterByPossible returns a new CellSlice with only the cells in the list that have the given number
//as an active possibility.
func (self CellSlice) FilterByPossible(possible int) CellSlice {
	//TODO: test this
	filter := func(cell *Cell) bool {
		return cell.Possible(possible)
	}
	return self.Filter(filter)
}

//FilterByNumPossibles returns a new CellSlice with only cells that have precisely the provided
//number of possible numbers.
func (self CellSlice) FilterByNumPossibilities(target int) CellSlice {
	//TODO: test this
	filter := func(cell *Cell) bool {
		return len(cell.Possibilities()) == target
	}
	return self.Filter(filter)
}

//FilterByHasPossibilities returns a new CellSlice with only cells that have 0 or more open possibilities.
func (self CellSlice) FilterByHasPossibilities() CellSlice {
	//Returns a list of cells that have possibilities.
	//TODO: test this.
	filter := func(cell *Cell) bool {
		return len(cell.Possibilities()) > 0
	}
	return self.Filter(filter)
}

//RemoveCells returns a new CellSlice that does not contain any of the cells included in the provided CellSlice.
func (self CellSlice) RemoveCells(targets CellSlice) CellSlice {
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

//PossibilitiesUnion returns an IntSlice that is the union of all active possibilities in cells in the set.
func (self CellSlice) PossibilitiesUnion() IntSlice {
	//Returns an IntSlice of the union of all possibilities.
	set := make(map[int]bool)

	for _, cell := range self {
		for _, possibility := range cell.Possibilities() {
			set[possibility] = true
		}
	}

	result := make(IntSlice, len(set))

	i := 0
	for possibility := range set {
		result[i] = possibility
		i++
	}

	return result
}

//Subset returns a new CellSlice that is the subset of the list including the items at the indexes provided
//in the IntSlice. See also InverseSubset.
func (self CellSlice) Subset(indexes IntSlice) CellSlice {
	//IntSlice.Subset is basically a carbon copy.
	//TODO: what's this behavior if indexes has dupes? What SHOULD it be?
	result := make(CellSlice, len(indexes))
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

//InverseSubset returns a new CellSlice that contains all of the elements from the list that are *not*
//at the indexes provided in the IntSlice. See also Subset.
func (self CellSlice) InverseSubset(indexes IntSlice) CellSlice {
	//TODO: figure out what this should do when presented with dupes.

	//LIke Subset, but returns all of the items NOT called out in indexes.
	var result CellSlice

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

//Sort mutates the provided CellSlice so that the cells are in order from left to right, top to bottom
//based on their position in the grid.
func (self CellSlice) Sort() {
	sorter := CellSliceSorter{self}
	sort.Sort(sorter)
}

//FilledNums returns an IntSlice representing all of the numbers that have been actively set on cells
//in the list. Cells that are empty (are set to '0') are not included.
func (self CellSlice) FilledNums() IntSlice {
	set := make(intSet)
	for _, cell := range self {
		if cell.Number() == 0 {
			continue
		}
		set[cell.Number()] = true
	}
	return set.toSlice()
}

//CollectNums collects the result of running fetcher across all items in the list.
func (self CellSlice) CollectNums(fetcher func(*Cell) int) IntSlice {
	var result IntSlice
	for _, cell := range self {
		result = append(result, fetcher(cell))
	}
	return result
}

func (self CellSliceSorter) Len() int {
	return len(self.CellSlice)
}

func (self CellSliceSorter) Less(i, j int) bool {
	//Sort based on the index of the cell.
	one := self.CellSlice[i]
	two := self.CellSlice[j]

	return (one.Row()*DIM + one.Col()) < (two.Row()*DIM + two.Col())
}

func (self CellSliceSorter) Swap(i, j int) {
	self.CellSlice[i], self.CellSlice[j] = self.CellSlice[j], self.CellSlice[i]
}

//Filter returns a new CellSlice that includes all cells where filter returned true.
func (self CellSlice) Filter(filter func(*Cell) bool) CellSlice {
	var result CellSlice
	for _, cell := range self {
		if filter(cell) {
			result = append(result, cell)
		}
	}
	return result
}

//Map executes the mapper function on each cell in the list.
func (self CellSlice) Map(mapper func(*Cell)) {
	for _, cell := range self {
		mapper(cell)
	}
}

//TODO: should this be in this file? It's awfully specific to HumanSolve needs, and extremely complex.
//TODO: is this how you spell this?
func (self CellSlice) chainDissimilarity(other CellSlice) float64 {
	//Returns a value between 0.0 and 1.0 depending on how 'similar' the CellSlices are.

	if other == nil || len(self) == 0 || len(other) == 0 {
		return 1.0
	}

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
		selfRow[cell.Row()] += selfProportion
		selfCol[cell.Col()] += selfProportion
		selfBlock[cell.Block()] += selfProportion
	}

	for _, cell := range other {
		otherRow[cell.Row()] += otherProportion
		otherCol[cell.Col()] += otherProportion
		otherBlock[cell.Block()] += otherProportion
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
	//So we';; basically put the lowest one into the average 4 times, 2 times for next, and 1 time for last.
	weights := []int{4, 2, 1}

	result := 0.0

	for i := 0; i < 3; i++ {
		for j := 0; j < weights[i]; j++ {
			result += diffs[i]
		}
	}

	//Divide by 4 + 2 + 1 = 7 to make it a weighted average
	result /= 7.0

	//Calculating the real upper bound is tricky, so we'll just assume it's 2.0 for simplicity and normalize based on that.
	if result > 2.0 {
		result = 2.0
	}

	//We strengthen the effect quite a bit here, otherwise we don't see much of an impact in SolveDirections.
	//The lower the dissimilarity, the stronger the effect will be.
	//TODO: logically this should belong in tweakChainedSteps, but for some reason when we do the math.Pow there it doesn't
	//appear to actually make a difference, which is maddening.
	return math.Pow(result/2.0, 10)

}

//Description returns a human-readable description of the cells in the list, like "(0,1), (0,2), and (0,3)"
func (self CellSlice) Description() string {
	strings := make(stringSlice, len(self))

	for i, cell := range self {
		strings[i] = cell.ref().String()
	}

	return strings.description()
}

func (self CellSlice) sameAsRefs(refs []cellRef) bool {
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

	for item := range cellSet {
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

//Description returns a human readable description of the ints in the set, like "7, 4, and 3"
func (self IntSlice) Description() string {

	strings := make(stringSlice, len(self))

	for i, num := range self {
		strings[i] = strconv.Itoa(num)
	}

	return strings.description()

}

//Unique returns a new IntSlice like the receiver, but with any duplicates removed. Order is not preserved.
func (self IntSlice) Unique() IntSlice {
	//TODO: test this.
	return self.toIntSet().toSlice()
}

//Same returns true if all ints in the slice are the same.
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

//SameContetnAs returns true if the receiver and otherSlice have the same list of ints
//(although not necessarily the same ordering of them)
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

//SameAs returns true if the receiver and otherSlice have the same ints in the same order. See also
//SameContentAs.
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

//Subset returns a new IntSlice like the receiver, but with only the ints at the provided indexes kept.
func (self IntSlice) Subset(indexes IntSlice) IntSlice {
	//TODO: test this.
	//Basically a carbon copy of CellSlice.Subset
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

//Sort sorts the IntSlice in place from small to large.
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

func (self CellSlice) toCellSet() cellSet {
	result := make(cellSet)
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

func (self cellSet) toSlice() CellSlice {
	var result CellSlice
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

func (self cellSet) intersection(other cellSet) cellSet {
	result := make(cellSet)
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

func (self cellSet) difference(other cellSet) cellSet {
	result := make(cellSet)
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

func (self cellSet) union(other cellSet) cellSet {
	result := make(cellSet)
	for item, value := range self {
		result[item] = value
	}
	for item, value := range other {
		result[item] = value
	}
	return result
}

//Intersection returns a new IntSlice that represents the intersection of the two IntSlices,
//that is, the ints that appear in both slices.
func (self IntSlice) Intersection(other IntSlice) IntSlice {
	//Returns an IntSlice of the union of both intSlices

	return self.toIntSet().intersection(other.toIntSet()).toSlice()
}

//Difference returns a new IntSlice that contains all of the ints in the receiver that are not
//also in other.
func (self IntSlice) Difference(other IntSlice) IntSlice {
	return self.toIntSet().difference(other.toIntSet()).toSlice()
}
