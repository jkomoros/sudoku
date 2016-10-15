package sudoku

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

//CellSlice is a list of cells with many convenience methods for doing common
//operations on them. MutableCellSlice is similar, but operates on a slice of
//MutableCells.
type CellSlice []Cell

//MutableCellSlice is a CellSlice that contains references to MutableCells. It
//doesn't have analogues for each CellSlice method; only ones that return a
//CellSlice. Note: MutableCellSlices are not mutable themselves; the "mutable"
//refers to the MutableCell.
type MutableCellSlice []MutableCell

//TODO: Audit all uses of {CellSlice, MutableCellSlice}.CellReferenceSlice to
//make sure we aren't doing unnecessary conversions once we've implemented
//more nums

//TODO: remove more CellSlice methods that aren't necessary anymore.

//TODO: crazy idea: CellReference should just be an interface with Row(),
//Col(), Block(). That way actual cells could be used. ... But then actual
//cellsw ould have to grow a Cell, MutableCell.

//TODO: rename CellReferenceSlice to CellRefSlice, CellRef?

//TODO: test whether the CellSlice caching on mutableGrid for row,block,col is
//actually useful anymore.

//TODO: make Neighbors, Row, Col, Block, etc global public functions that give
//a CellReferenceSlice. Then audit all places we use Row, Col, Block and see
//if we ACTUALLY need those, or can just use the global. We might never need
//to expose this externally, because they're kind of annoying, and the need
//for performance is only really important internally to the package.

//CellReferenceSlice is a slice of CellReferences with many convenience methods.
type CellReferenceSlice []CellReference

//IntSlice is a list of ints, with many convenience methods specific to sudoku.
type IntSlice []int

type stringSlice []string

type intSet map[int]bool

//TODO: consider removing cellSet, since it's not actually used anywhere
//(it was built for forcing_chains, but we ended up not using it there)
type cellSet map[CellReference]bool

//CellReference is a reference to a generic cell located at a specific row and
//column.
type CellReference struct {
	Row int
	Col int
}

type cellSliceSorter struct {
	CellSlice
}

type mutableCellSliceSorter struct {
	MutableCellSlice
}

type cellReferenceSliceSorter struct {
	CellReferenceSlice
}

func getRow(cell Cell) int {
	return cell.Row()
}

func getCol(cell Cell) int {
	return cell.Col()
}

func getBlock(cell Cell) int {
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

//SameRow returns true if all cells are in the same row.
func (self CellReferenceSlice) SameRow() bool {
	return self.CollectNums(func(cell CellReference) int {
		return cell.Row
	}).Same()
}

//SameCol returns true if all cells are in the same column.
func (self CellReferenceSlice) SameCol() bool {
	return self.CollectNums(func(cell CellReference) int {
		return cell.Col
	}).Same()
}

//SameBlock returns true if all cells are in the same block.
func (self CellReferenceSlice) SameBlock() bool {
	return self.CollectNums(func(cell CellReference) int {
		return cell.Block()
	}).Same()
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

//Row returns the row that at least one of the cells is in. If SameRow() is false, the Row
//may be any of the rows in the set.
func (self CellReferenceSlice) Row() int {
	//Will return the row of a random item.
	if len(self) == 0 {
		return 0
	}
	return self[0].Row
}

//Col returns the column that at least one of the cells is in. If SameCol() is false, the column
//may be any of the columns in the set.
func (self CellReferenceSlice) Col() int {
	if len(self) == 0 {
		return 0
	}
	return self[0].Col
}

//Block returns the row that at least one of the cells is in. If SameBlock() is false, the Block
//may be any of the blocks in the set.
func (self CellReferenceSlice) Block() int {
	if len(self) == 0 {
		return 0
	}
	return self[0].Block()
}

//AllRows returns all of the rows for cells in this slice.
func (self CellSlice) AllRows() IntSlice {
	//TODO: test this.
	return self.CollectNums(getRow).Unique()
}

//AllCols returns all of the columns for cells in this slice.
func (self CellSlice) AllCols() IntSlice {
	//TODO: test this.
	return self.CollectNums(getCol).Unique()
}

//AllBlocks returns all of the blocks for cells in this slice.
func (self CellSlice) AllBlocks() IntSlice {
	//TODO: test this.
	return self.CollectNums(getBlock).Unique()
}

//AllRows returns all of the rows for cells in this slice.
func (self CellReferenceSlice) AllRows() IntSlice {
	//TODO: test this.
	return self.CollectNums(func(cell CellReference) int {
		return cell.Row
	}).Unique()
}

//AllCols returns all of the columns for cells in this slice.
func (self CellReferenceSlice) AllCols() IntSlice {
	//TODO: test this.
	return self.CollectNums(func(cell CellReference) int {
		return cell.Col
	}).Unique()
}

//AllBlocks returns all of the blocks for cells in this slice.
func (self CellReferenceSlice) AllBlocks() IntSlice {
	//TODO: test this.
	return self.CollectNums(func(cell CellReference) int {
		return cell.Block()
	}).Unique()
}

//CellReferenceSlice returns a CellReferenceSlice that corresponds to the
//cells in this MutableCellSlice.
func (self MutableCellSlice) CellReferenceSlice() CellReferenceSlice {
	result := make(CellReferenceSlice, len(self))

	for i, cell := range self {
		result[i] = cell.Reference()
	}

	return result
}

//FilterByUnfilled returns a new CellSlice with only the cells in the list
//that are not filled with any number.
func (self CellSlice) FilterByUnfilled() CellSlice {
	//TODO: test this
	filter := func(cell Cell) bool {
		return cell.Number() == 0
	}
	return self.Filter(filter)
}

//FilterByUnfilled returns a new CellSlice with only the cells in the list
//that are not filled with any number.
func (self MutableCellSlice) FilterByUnfilled() MutableCellSlice {
	//TODO: test all of the mutableCellSlice methods
	//TODO: test this
	filter := func(cell Cell) bool {
		return cell.Number() == 0
	}
	return self.Filter(filter)
}

//FilterByFilled returns a new CellSlice with only the cells that have a
//number in them.
func (self CellSlice) FilterByFilled() CellSlice {
	//TODO: test this
	filter := func(cell Cell) bool {
		return cell.Number() != 0
	}
	return self.Filter(filter)
}

//FilterByFilled returns a new CellSlice with only the cells that have a
//number in them.
func (self MutableCellSlice) FilterByFilled() MutableCellSlice {
	//TODO: test this
	filter := func(cell Cell) bool {
		return cell.Number() != 0
	}
	return self.Filter(filter)
}

//FilterByPossible returns a new CellSlice with only the cells in the list that have the given number
//as an active possibility.
func (self CellSlice) FilterByPossible(possible int) CellSlice {
	//TODO: test this
	filter := func(cell Cell) bool {
		return cell.Possible(possible)
	}
	return self.Filter(filter)
}

//FilterByPossible returns a new CellSlice with only the cells in the list that have the given number
//as an active possibility.
func (self MutableCellSlice) FilterByPossible(possible int) MutableCellSlice {
	//TODO: test this
	filter := func(cell Cell) bool {
		return cell.Possible(possible)
	}
	return self.Filter(filter)
}

//FilterByNumPossibles returns a new CellSlice with only cells that have precisely the provided
//number of possible numbers.
func (self CellSlice) FilterByNumPossibilities(target int) CellSlice {
	//TODO: test this
	filter := func(cell Cell) bool {
		return len(cell.Possibilities()) == target
	}
	return self.Filter(filter)
}

//FilterByNumPossibles returns a new CellSlice with only cells that have precisely the provided
//number of possible numbers.
func (self MutableCellSlice) FilterByNumPossibilities(target int) MutableCellSlice {
	//TODO: test this
	filter := func(cell Cell) bool {
		return len(cell.Possibilities()) == target
	}
	return self.Filter(filter)
}

//FilterByHasPossibilities returns a new CellSlice with only cells that have 0 or more open possibilities.
func (self CellSlice) FilterByHasPossibilities() CellSlice {
	//Returns a list of cells that have possibilities.
	//TODO: test this.
	filter := func(cell Cell) bool {
		return len(cell.Possibilities()) > 0
	}
	return self.Filter(filter)
}

//FilterByHasPossibilities returns a new CellSlice with only cells that have 0 or more open possibilities.
func (self MutableCellSlice) FilterByHasPossibilities() MutableCellSlice {
	//Returns a list of cells that have possibilities.
	//TODO: test this.
	filter := func(cell Cell) bool {
		return len(cell.Possibilities()) > 0
	}
	return self.Filter(filter)
}

//RemoveCells returns a new CellSlice that does not contain any of the cells included in the provided CellSlice.
func (self CellSlice) RemoveCells(targets CellSlice) CellSlice {
	//TODO: test this.
	targetCells := make(map[Cell]bool)
	for _, cell := range targets {
		targetCells[cell] = true
	}
	filterFunc := func(cell Cell) bool {
		return !targetCells[cell]
	}
	return self.Filter(filterFunc)
}

//RemoveCells returns a new CellSlice that does not contain any of the cells included in the provided CellSlice.
func (self MutableCellSlice) RemoveCells(targets CellSlice) MutableCellSlice {
	//TODO: test this.
	targetCells := make(map[Cell]bool)
	for _, cell := range targets {
		targetCells[cell] = true
	}
	filterFunc := func(cell Cell) bool {
		return !targetCells[cell]
	}
	return self.Filter(filterFunc)
}

//RemoveCells returns a new CellReferenceSlice that does not contain any of
//the cells included in the provided CellReferenceSlice.
func (self CellReferenceSlice) RemoveCells(targets CellReferenceSlice) CellReferenceSlice {
	//TODO: test this.
	targetCells := make(map[CellReference]bool)
	for _, cell := range targets {
		targetCells[cell] = true
	}
	filterFunc := func(cell CellReference) bool {
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

//Subset returns a new CellSlice that is the subset of the list including the items at the indexes provided
//in the IntSlice. See also InverseSubset.
func (self MutableCellSlice) Subset(indexes IntSlice) MutableCellSlice {
	//IntSlice.Subset is basically a carbon copy.
	//TODO: what's this behavior if indexes has dupes? What SHOULD it be?
	result := make(MutableCellSlice, len(indexes))
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

//Subset returns a new CellReferenceSlice that is the subset of the list including the items at the indexes provided
//in the IntSlice. See also InverseSubset.
func (self CellReferenceSlice) Subset(indexes IntSlice) CellReferenceSlice {
	//IntSlice.Subset is basically a carbon copy.
	//TODO: what's this behavior if indexes has dupes? What SHOULD it be?
	result := make(CellReferenceSlice, len(indexes))
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

//InverseSubset returns a new CellSlice that contains all of the elements from the list that are *not*
//at the indexes provided in the IntSlice. See also Subset.
func (self MutableCellSlice) InverseSubset(indexes IntSlice) MutableCellSlice {
	//TODO: figure out what this should do when presented with dupes.

	//LIke Subset, but returns all of the items NOT called out in indexes.
	var result MutableCellSlice

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

//InverseSubset returns a new CellReferenceSlice that contains all of the elements from the list that are *not*
//at the indexes provided in the IntSlice. See also Subset.
func (self CellReferenceSlice) InverseSubset(indexes IntSlice) CellReferenceSlice {
	//TODO: figure out what this should do when presented with dupes.

	//LIke Subset, but returns all of the items NOT called out in indexes.
	var result CellReferenceSlice

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
	sorter := cellSliceSorter{self}
	sort.Sort(sorter)
}

//Sort mutates the provided CellSlice so that the cells are in order from left to right, top to bottom
//based on their position in the grid.
func (self MutableCellSlice) Sort() {
	sorter := mutableCellSliceSorter{self}
	sort.Sort(sorter)
}

//Sort mutates the provided CellReferenceSlice so that the cells are in order
//from left to right, top to bottom based on their position in the grid.
func (self CellReferenceSlice) Sort() {
	//TODO: note that this is dangerous to have because we cache the public
	//rows,cols,blocks, and don't ahve locks.
	sorter := cellReferenceSliceSorter{self}
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
func (self CellSlice) CollectNums(fetcher func(Cell) int) IntSlice {
	var result IntSlice
	for _, cell := range self {
		result = append(result, fetcher(cell))
	}
	return result
}

//CollectNums collects the result of running fetcher across all items in the list.
func (self CellReferenceSlice) CollectNums(fetcher func(CellReference) int) IntSlice {
	var result IntSlice
	for _, cell := range self {
		result = append(result, fetcher(cell))
	}
	return result
}

func (self cellSliceSorter) Len() int {
	return len(self.CellSlice)
}

func (self mutableCellSliceSorter) Len() int {
	return len(self.MutableCellSlice)
}

func (self cellReferenceSliceSorter) Len() int {
	return len(self.CellReferenceSlice)
}

func (self cellSliceSorter) Less(i, j int) bool {
	//Sort based on the index of the cell.
	one := self.CellSlice[i]
	two := self.CellSlice[j]

	return (one.Row()*DIM + one.Col()) < (two.Row()*DIM + two.Col())
}

func (self mutableCellSliceSorter) Less(i, j int) bool {
	//Sort based on the index of the cell.
	one := self.MutableCellSlice[i]
	two := self.MutableCellSlice[j]

	return (one.Row()*DIM + one.Col()) < (two.Row()*DIM + two.Col())
}

func (self cellReferenceSliceSorter) Less(i, j int) bool {
	//Sort based on the index of the cell.
	one := self.CellReferenceSlice[i]
	two := self.CellReferenceSlice[j]

	return (one.Row*DIM + one.Col) < (two.Row*DIM + two.Col)
}

func (self cellSliceSorter) Swap(i, j int) {
	self.CellSlice[i], self.CellSlice[j] = self.CellSlice[j], self.CellSlice[i]
}

func (self mutableCellSliceSorter) Swap(i, j int) {
	self.MutableCellSlice[i], self.MutableCellSlice[j] = self.MutableCellSlice[j], self.MutableCellSlice[i]
}

func (self cellReferenceSliceSorter) Swap(i, j int) {
	self.CellReferenceSlice[i], self.CellReferenceSlice[j] = self.CellReferenceSlice[j], self.CellReferenceSlice[i]
}

//Filter returns a new CellSlice that includes all cells where filter returned true.
func (self CellSlice) Filter(filter func(Cell) bool) CellSlice {
	var result CellSlice
	for _, cell := range self {
		if filter(cell) {
			result = append(result, cell)
		}
	}
	return result
}

//Filter returns a new CellSlice that includes all cells where filter returned true.
func (self MutableCellSlice) Filter(filter func(Cell) bool) MutableCellSlice {
	var result MutableCellSlice
	for _, cell := range self {
		if filter(cell) {
			result = append(result, cell)
		}
	}
	return result
}

//Filter returns a new CellReferenceSlice that includes all cells where filter returned true.
func (self CellReferenceSlice) Filter(filter func(CellReference) bool) CellReferenceSlice {
	var result CellReferenceSlice
	for _, cell := range self {
		if filter(cell) {
			result = append(result, cell)
		}
	}
	return result
}

//Map executes the mapper function on each cell in the list.
func (self CellSlice) Map(mapper func(Cell)) {
	for _, cell := range self {
		mapper(cell)
	}
}

//Map executes the mapper function on each cell in the list.
func (self MutableCellSlice) Map(mapper func(MutableCell)) {
	for _, cell := range self {
		mapper(cell)
	}
}

//cellSlice returns a CellSlice of the same cells
func (self MutableCellSlice) cellSlice() CellSlice {
	result := make(CellSlice, len(self))
	for i, item := range self {
		result[i] = item
	}
	return result
}

//chainSimilarity returns a value between 0.0 and 1.0 depending on how
//'similar' the CellSlices are. For example, two cells that are in the same
//row within the same block are very similar; cells that are in different
//rows, cols, and blocks are extremelye dissimilar.
func (self CellReferenceSlice) chainSimilarity(other CellReferenceSlice) float64 {

	//TODO: should this be in this file? It's awfully specific to HumanSolve needs, and extremely complex.
	if other == nil || len(self) == 0 || len(other) == 0 {
		return 1.0
	}

	//Note: it doesn't ACTUALLY guarantee a value lower than 1.0 (it might be possible to hit those; reasoning about the maximum value is tricky).

	//Note: a 1.0 means extremely similar, and 0.0 means extremely dissimilar.

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

	//Keep track of how many of each we added to each row, col, and block so
	//we can do one more pass to normalize to proportions.
	rowCounter := 0.0
	colCounter := 0.0
	blockCounter := 0.0

	offbyOneIncrement := 0.25

	for _, cell := range self {
		selfRow[cell.Row] += 1.0
		rowCounter++
		if cell.Row > 0 {
			selfRow[cell.Row-1] += offbyOneIncrement
			rowCounter += offbyOneIncrement
		}
		if cell.Row < DIM-1 {
			selfRow[cell.Row+1] += offbyOneIncrement
			rowCounter += offbyOneIncrement
		}

		selfCol[cell.Col] += 1.0
		colCounter++
		if cell.Col > 0 {
			selfCol[cell.Col-1] += offbyOneIncrement
			colCounter += offbyOneIncrement
		}
		if cell.Col < DIM-1 {
			selfCol[cell.Col+1] += offbyOneIncrement
			colCounter += offbyOneIncrement
		}

		selfBlock[cell.Block()] += 1.0
		blockCounter++

		//Nearby blocks don't get the partial points. If we were to, we'd give
		//points for blocks that are directly adjacent, or half as many for
		//diagonally adjacent. But calculating that neighbors is non-trivial
		//right now so just skip it.
	}

	//Normalize self slices by count for each.
	for i := 0; i < DIM; i++ {
		selfRow[i] /= rowCounter
		selfCol[i] /= colCounter
		selfBlock[i] /= blockCounter
	}

	rowCounter, colCounter, blockCounter = 0.0, 0.0, 0.0

	for _, cell := range other {
		otherRow[cell.Row] += 1.0
		rowCounter++
		if cell.Row > 0 {
			otherRow[cell.Row-1] += offbyOneIncrement
			rowCounter += offbyOneIncrement
		}
		if cell.Row < DIM-1 {
			otherRow[cell.Row+1] += offbyOneIncrement
			rowCounter += offbyOneIncrement
		}

		otherCol[cell.Col] += 1.0
		colCounter++
		if cell.Col > 0 {
			otherCol[cell.Col-1] += offbyOneIncrement
			colCounter += offbyOneIncrement
		}
		if cell.Col < DIM-1 {
			otherCol[cell.Col+1] += offbyOneIncrement
			colCounter += offbyOneIncrement
		}

		otherBlock[cell.Block()] += 1.0
		blockCounter++

		//Nearby blocks don't get the partial points. If we were to, we'd give
		//points for blocks that are directly adjacent, or half as many for
		//diagonally adjacent. But calculating that neighbors is non-trivial
		//right now so just skip it.
	}

	//Normalize other slices by count for each.
	for i := 0; i < DIM; i++ {
		otherRow[i] /= rowCounter
		otherCol[i] /= colCounter
		otherBlock[i] /= blockCounter
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

	//Normalize between 0.0 and 1.0
	result = result / 2.0

	//Currently similar things are 0 and dissimilar things are 1.0; flip it.
	return 1.0 - result

}

//Description returns a human-readable description of the cells in the list, like "(0,1), (0,2), and (0,3)"
func (self CellSlice) Description() string {
	//TODO: consider getting rid of this method.
	return self.CellReferenceSlice().Description()
}

//Description returns a human-readable description of the cells in the list, like "(0,1), (0,2), and (0,3)"
func (self CellReferenceSlice) Description() string {
	strings := make(stringSlice, len(self))

	for i, cell := range self {
		strings[i] = cell.String()
	}

	return strings.description()
}

//CellReferenceSlice returns a CellReferenceSlice that corresponds to the
//cells in this CellSlice.
func (self CellSlice) CellReferenceSlice() CellReferenceSlice {
	result := make(CellReferenceSlice, len(self))

	for i, cell := range self {
		result[i] = cell.Reference()
	}

	return result
}

func (self CellSlice) sameAsRefs(refs CellReferenceSlice) bool {

	//TODO: audit all of the private methods on CellSlice, MutableCellSlice
	//now that we might not use them since we use something on
	//CellReferenceSlice.
	cellSet := make(map[string]bool)
	for _, cell := range self {
		cellSet[cell.Reference().String()] = true
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

func (self CellReferenceSlice) sameAs(refs CellReferenceSlice) bool {
	cellSet := make(map[string]bool)
	for _, cell := range self {
		cellSet[cell.String()] = true
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

//MutableCell returns the MutableCell in the given grid that this
//CellReference refers to.
func (self CellReference) MutableCell(grid MutableGrid) MutableCell {
	if grid == nil {
		return nil
	}
	return grid.MutableCell(self.Row, self.Col)
}

//Cell returns the Cell in the given grid that this CellReference refers to.
func (self CellReference) Cell(grid Grid) Cell {
	if grid == nil {
		return nil
	}
	return grid.Cell(self.Row, self.Col)
}

//Block returns the block that this CellReference is in.
func (self CellReference) Block() int {
	return blockForCell(self.Row, self.Col)
}

func (self CellReference) String() string {
	return "(" + strconv.Itoa(self.Row) + "," + strconv.Itoa(self.Col) + ")"
}

//CellSlice returns a CellSlice with Cells corresponding to our references, in
//the given grid.
func (self CellReferenceSlice) CellSlice(grid Grid) CellSlice {

	result := make(CellSlice, len(self))

	for i, ref := range self {
		result[i] = ref.Cell(grid)
	}

	return result

}

//CellSlice returns a MutableCellSlice with MutableCells corresponding to our
//references, in the given grid.
func (self CellReferenceSlice) MutableCellSlice(grid MutableGrid) MutableCellSlice {

	result := make(MutableCellSlice, len(self))

	for i, ref := range self {
		result[i] = ref.MutableCell(grid)
	}

	return result

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
		result[item.Reference()] = true
	}
	return result
}

func (self MutableCellSlice) toCellSet() cellSet {
	result := make(cellSet)
	for _, item := range self {
		result[item.Reference()] = true
	}
	return result
}

func (self CellReferenceSlice) toCellSet() cellSet {
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

func (self cellSet) toSlice(grid Grid) CellSlice {
	var result CellSlice
	for item, val := range self {
		if val {
			result = append(result, item.Cell(grid))
		}
	}
	return result
}

func (self cellSet) toReferenceSlice() CellReferenceSlice {
	var result CellReferenceSlice
	for item, val := range self {
		if val {
			result = append(result, item)
		}
	}
	return result
}

func (self cellSet) toMutableSlice(grid MutableGrid) MutableCellSlice {
	var result MutableCellSlice
	for item, val := range self {
		if val {
			result = append(result, item.MutableCell(grid))
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

//Intersection returns a new CellSlice that represents the intersection of the
//two CellSlices; that is, the cells that appear in both slices.
func (self CellSlice) Intersection(other CellSlice) CellSlice {
	if len(self) == 0 {
		return nil
	}
	grid := self[0].Grid()
	return self.toCellSet().intersection(other.toCellSet()).toSlice(grid)
}

//Intersection returns a new CellSlice that represents the intersection of the
//two CellSlices; that is, the cells that appear in both slices.
func (self MutableCellSlice) Intersection(other CellSlice) MutableCellSlice {
	if len(self) == 0 {
		return nil
	}
	grid := self[0].MutableGrid()
	return self.toCellSet().intersection(other.toCellSet()).toMutableSlice(grid)
}

//Intersection returns a new CellSlice that represents the intersection of the
//two CellReferenceSlices; that is, the cells that appear in both slices.
func (self CellReferenceSlice) Intersection(other CellReferenceSlice) CellReferenceSlice {
	if len(self) == 0 {
		return nil
	}
	return self.toCellSet().intersection(other.toCellSet()).toReferenceSlice()
}

//Difference returns a new CellSlice that contains all of the cells in the
//receiver that are not also in the other.
func (self CellSlice) Difference(other CellSlice) CellSlice {
	if len(self) == 0 {
		return nil
	}
	grid := self[0].Grid()
	return self.toCellSet().difference(other.toCellSet()).toSlice(grid)
}

//Difference returns a new CellSlice that contains all of the cells in the
//receiver that are not also in the other.
func (self MutableCellSlice) Difference(other CellSlice) MutableCellSlice {
	if len(self) == 0 {
		return nil
	}
	grid := self[0].MutableGrid()
	return self.toCellSet().difference(other.toCellSet()).toMutableSlice(grid)
}

//Difference returns a new CellReferenceSlice that contains all of the cells in the
//receiver that are not also in the other.
func (self CellReferenceSlice) Difference(other CellReferenceSlice) CellReferenceSlice {
	if len(self) == 0 {
		return nil
	}
	return self.toCellSet().difference(other.toCellSet()).toReferenceSlice()
}

//Union returns a new CellSlice that contains all of the cells that are in
//either the receiver or the other CellSlice.
func (self CellSlice) Union(other CellSlice) CellSlice {
	if len(self) == 0 {
		return nil
	}
	grid := self[0].Grid()
	return self.toCellSet().union(other.toCellSet()).toSlice(grid)
}

//Union returns a new CellSlice that contains all of the cells that are in
//either the receiver or the other CellSlice.
func (self MutableCellSlice) Union(other CellSlice) MutableCellSlice {
	if len(self) == 0 {
		return nil
	}
	grid := self[0].MutableGrid()
	return self.toCellSet().union(other.toCellSet()).toMutableSlice(grid)
}

//Union returns a new CellSlice that contains all of the cells that are in
//either the receiver or the other CellSlice.
func (self CellReferenceSlice) Union(other CellReferenceSlice) CellReferenceSlice {
	if len(self) == 0 {
		return nil
	}
	return self.toCellSet().union(other.toCellSet()).toReferenceSlice()
}
