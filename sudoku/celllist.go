package sudoku

type CellList []*Cell

type intList []int

func (self CellList) SameRow() bool {
	if len(self) == 0 {
		return true
	}
	row := self[0].Row
	for _, cell := range self {
		if cell.Row != row {
			return false
		}
	}
	return true
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
