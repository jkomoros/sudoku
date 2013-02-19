package sudoku

type CellList []*Cell

type intList []int

func getRow(cell *Cell) int {
	return cell.Row
}

func (self CellList) SameRow() bool {
	return self.CollectNums(getRow).Same()

}

func (self CellList) CollectNums(fetcher func(*Cell) int) intList {
	var result intList
	for _, cell := range self {
		result = append(result, fetcher(cell))
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
