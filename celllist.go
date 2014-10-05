package sudoku

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type CellList []*Cell

type IntSlice []int

type stringSlice []string

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

func (self CellList) CollectNums(fetcher func(*Cell) int) IntSlice {
	var result IntSlice
	for _, cell := range self {
		result = append(result, fetcher(cell))
	}
	return result
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

func (self CellList) Description() string {
	strings := make(stringSlice, len(self))

	for i, cell := range self {
		strings[i] = fmt.Sprintf("(%d,%d)", cell.Row, cell.Col)
	}

	return strings.description()
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