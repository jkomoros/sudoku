package sudoku

type CellList []*Cell

type intList []int

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

func (self CellList) FilterByPossible(possible int) CellList {
	//TODO: test this
	filter := func(cell *Cell) bool {
		return cell.Possible(possible)
	}
	return self.Filter(filter)
}

func (self CellList) CollectNums(fetcher func(*Cell) int) intList {
	var result intList
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

func (self intList) Same() bool {
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
