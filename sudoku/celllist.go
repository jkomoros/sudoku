package sudoku

type CellList []*Cell

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
