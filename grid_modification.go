package sudoku

//GridModification is a series of CellModifications to apply to a Grid.
type GridModification []*CellModification

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

//TODO: audit all uses of step/compoundstep.Apply()

//TOOD: make readOnlyCellImpl. Test if neighbors should be derived or not.
//Everything should be actual contiguous memory, no pointers (except for
//grid). Likely should make cellImpl embed a readOnlyCellImpl and only
//override items it needs to.

//TODO: make readOnlyGridImpl. Two possible approaches: a version that is
//incredibly easy to copy and then do minor tweaks. Or a version that stores a
//dictionary of cell configs, and any time you grab a Cell we look it up in
//the dict or in the ancestors' dicts.

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

//equivalent returns true if the other grid modification is equivalent to this one.
func (m GridModification) equivalent(other GridModification) bool {
	if len(m) != len(other) {
		return false
	}
	for i, modification := range m {
		otherModification := other[i]
		if modification.Cell.ref().String() != otherModification.Cell.ref().String() {
			return false
		}
		if modification.Number != otherModification.Number {
			return false
		}

		if len(modification.ExcludesChanges) != len(otherModification.ExcludesChanges) {
			return false
		}

		for key, val := range modification.ExcludesChanges {
			otherVal, ok := otherModification.ExcludesChanges[key]
			if !ok {
				return false
			}
			if val != otherVal {
				return false
			}
		}

		if len(modification.MarksChanges) != len(otherModification.MarksChanges) {
			return false
		}

		for key, val := range modification.MarksChanges {
			otherVal, ok := otherModification.MarksChanges[key]
			if !ok {
				return false
			}
			if val != otherVal {
				return false
			}
		}
	}
	return true
}

func (self *gridImpl) CopyWithModifications(modifications GridModification) Grid {

	//TODO: test this implementation deeply! Lots of crazy stuff that could go
	//wrong.

	result := new(gridImpl)

	//Copy in everything
	*result = *self

	for i := 0; i < DIM*DIM; i++ {
		cell := &result.cells[i]
		cell.gridRef = result
	}

	cellNumberModified := false

	for _, modification := range modifications {
		cell := result.cellImpl(modification.Cell.Row(), modification.Cell.Col())

		if modification.Number >= 0 && modification.Number <= DIM {
			//cell.setNumber will handle setting all of the impossibles
			if cell.setNumber(modification.Number) {
				cellNumberModified = true
			}
		}

		for key, val := range modification.ExcludesChanges {
			//Key is 1-indexed
			key--
			cell.excluded[key] = val
		}

		for key, val := range modification.MarksChanges {
			//Key is 1-indexed
			key--
			cell.marks[key] = val
		}
	}

	if cellNumberModified {

		//At least one cell's number was modified, which means we need to fix
		//up the queue, numFilledCells, Invalid, Solved.

		filledCellsCount := 0

		for _, cell := range result.cells {
			if cell.number == 0 {
				continue
			}
			filledCellsCount++
		}

		result.filledCellsCount = filledCellsCount

		//Check if we're invalid.

		invalid := false

		for _, cell := range result.cells {
			//Make sure we have at least one possibility per cell
			foundPossibility := false
			for i := 0; i < DIM; i++ {
				if cell.impossibles[i] == 0 {
					foundPossibility = true
					break
				}
			}
			if !foundPossibility {
				invalid = true
				break
			}
		}

		if !invalid {
			//Let's do a deep check
			invalid = gridGroupsInvalid(result)
		}

		result.invalid = invalid

		if filledCellsCount == DIM*DIM && !result.invalid {
			//All cells are filled and it's not invalid, so it's solved!
			result.solved = true
		} else {
			//No way it's solved
			result.solved = false
		}

		result.theQueue.fix()
	}

	return result

}

func (self *mutableGridImpl) CopyWithModifications(modifications GridModification) Grid {
	//TODO: when we have an honest-to-god readonly grid impl, optimize this.
	result := self.MutableCopy()

	for _, modification := range modifications {
		cell := modification.Cell.MutableInGrid(result)

		if modification.Number >= 0 && modification.Number <= DIM {
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
