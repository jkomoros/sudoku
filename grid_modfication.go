package sudoku

//GridModification is a series of CellModifications to apply to a Grid.
type GridModifcation []*CellModification

//CellModification represents a modification to be made to a given Cell in a
//grid.
type CellModification struct {
	//The cell representing the cell to modify. The cell's analog (at the same
	//row, col address) will be modified in the new grid.
	Cell Cell
	//The number to put in the cell. Negative numbers signify no changes.
	Number int
	//The excludes to proactively set. Invalid numbers will be ignored.
	//Indexes not listed will be left the same.
	ExcludesChanges map[int]bool
	//The marks to proactively set. Invalid numbers will be ignored.
	//Indexes not listed will be left the same.
	MarksChanges map[int]bool
}

//newCellModification returns a CellModification for the given cell that is a
//no-op.
func newCellModification(cell Cell) *CellModification {
	return &CellModification{
		Cell:            cell,
		Number:          -1,
		ExcludesChanges: make(map[int]bool),
		MarksChanges:    make(map[int]bool),
	}
}

func (self *gridImpl) CopyWithModifications(modifications GridModifcation) Grid {
	//TODO: when we have an honest-to-god readonly grid impl, optimize this.
	result := self.Copy()

	for _, modification := range modifications {
		cell := modification.Cell.MutableInGrid(result)

		if modification.Number >= 0 && modification.Number < DIM {
			cell.SetNumber(modification.Number)
		}

		for key, val := range modification.ExcludesChanges {
			//setExcluded will skip invalid entries
			cell.SetExcluded(key, val)
		}

		for key, val := range modification.MarksChanges {
			//SetMark will skip invalid numbers
			cell.SetMark(key, val)
		}
	}

	return result
}
