/*

sudoku is a package for generating, solving, and rating the difficulty of sudoku puzzles.
Notably, it is able to solve a puzzle like a human would, and aspires to provide
high-quality difficulty ratings for puzzles based on a model trained on hundreds
of thousands of solves by real users.

*/
package sudoku

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

const _NUM_NEIGHBORS = (DIM-1)*3 - (BLOCK_DIM-1)*2

type SymmetryType int

const (
	SYMMETRY_NONE SymmetryType = iota
	SYMMETRY_ANY
	SYMMETRY_HORIZONTAL
	SYMMETRY_VERTICAL
	SYMMETRY_BOTH
)

//Cell represents a single cell within a grid. It maintains information about
//the number that is filled, the numbers that are currently legal given the
//filled status of its neighbors, and whether any possibilities have been
//explicitly excluded by solve techniques. Cells should not be constructed on
//their own; create a Grid and grab references to the cells from there. Cell
//does not contain methods to mutate the Cell. See MutableCell for that.
type Cell interface {
	//Row returns the cell's row in its parent grid.
	Row() int

	//Col returns the cell's column in its parent grid.
	Col() int

	//Block returns the cell's block in its parent grid.
	Block() int

	//InGrid returns a reference to a cell in the provided grid that has the same
	//row/column as this cell. Effectively, this cell's analogue in the other
	//grid.
	InGrid(grid *Grid) Cell

	//MutableInGrid is like InGrid, but will only work on grids that are mutable.
	MutableInGrid(grid *Grid) MutableCell

	//Number returns the number the cell is currently set to.
	Number() int

	//Mark reads out whether the given mark has been set for this cell. See
	//SetMark for a description of what marks represent.
	Mark(number int) bool

	//Marks returns an IntSlice with each mark, in ascending order.
	Marks() IntSlice

	//Possible returns whether or not a given number is legal to fill via
	//SetNumber, given the state of the grid (specifically, the cell's neighbors)
	//and the numbers the cell was told to explicitly exclude via SetExclude. If
	//the cell is already filled with a number, it will return false for all
	//numbers.
	Possible(number int) bool

	//Possibilities returns a list of all current possibilities for this cell: all
	//numbers for which cell.Possible returns true.
	Possibilities() IntSlice

	//Invalid returns true if the cell has no valid possibilities to fill in,
	//implying that the grid is in an invalid state because this cell cannot be
	//filled with a number without violating a constraint.
	Invalid() bool

	//Locked returns whether or not the cell is locked. See Lock for more
	//information on the concept of locking.
	Locked() bool

	//SymmetricalPartner returns the cell's partner in the grid, based on the type
	//of symmetry requested.
	SymmetricalPartner(symmetry SymmetryType) Cell

	//Neighbors returns a CellSlice of all of the cell's neighbors--the other
	//cells in its row, column, and block. The set of neighbors is the set of
	//cells that this cell's number must not conflict with.
	Neighbors() CellSlice

	//String returns a debug-friendly summary of the Cell.
	String() string

	//DiagramExtents returns the top, left, height, and width coordinate in
	//grid.Diagram's output that  corresponds to the contents of this cell. The
	//top left corner is 0,0
	DiagramExtents() (top, left, height, width int)

	//Mutable will return a MutableCell underlying this one, or nil if that's
	//not possible. This should succeed if you're calling it from a method
	//that was called on a MutableGrid, but otherwise expect it to fail. In
	//general it is always safer to use grid.MutableCell(r,c),
	//cell.MutableInGrid(grid), etc wherever available because thos will fail
	//at compile time instead of run-time.
	Mutable() MutableCell

	//The following are methods that are only internal. Some of them are
	//nasty.
	ref() cellRef
	grid() *Grid
	diagramRows(showMarks bool) []string
	rank() int
	implicitNumber() int
}

//MutableCell is a Cell that also has methods that allow mutation of the cell.
type MutableCell interface {
	//MutableCell contains all of Cell's (read-only) methods.
	Cell

	//MutableNeighbors returns a MutableCellSlice of all of the cell's
	//neighbors--the other cells in its row, column, and block. The set of
	//neighbors is the set of cells that this cell's number must not conflict
	//with.
	MutableNeighbors() MutableCellSlice

	//MutableSymmetricalPartner returns the cell's mutable partner in the
	//grid, based on the type of symmetry requested.
	MutableSymmetricalPartner(symmetry SymmetryType) MutableCell

	//SetNumber explicitly sets the number of the cell. This operation could cause
	//the grid to become invalid if it conflicts with its neighbors' numbers. This
	//operation will affect the Possiblities() of its neighbor cells.
	SetNumber(number int)

	//SetExcluded defines whether a possibility is considered not feasible, even
	//if not directly precluded by the Number()s of the cell's neighbors. This is
	//used by advanced HumanSolve techniques that cull possibilities that are
	//logically excluded by the state of the grid, in a non-direct way. The state
	//of Excluded bits will affect the results of this cell's Possibilities()
	//list.
	SetExcluded(number int, excluded bool)

	//ResetExcludes sets all excluded bits to false, so that Possibilities() will
	//be based purely on direct implications of the Number()s of neighbors. See
	//also SetExcluded.
	ResetExcludes()

	//SetMark sets the mark at the given index to true. Marks represent number
	//marks proactively added to a cell by a user. They have no effect on the
	//solver or human solver; they only are visible when Diagram(true) is called.
	SetMark(number int, mark bool)

	//ResetMarks removes all marks. See SetMark for a description of what marks
	//represent.
	ResetMarks()

	//Lock 'locks' the cell. Locking represents the concept of cells that are set
	//at the beginning of the puzzle and that users may not modify. Locking does
	//not change whether calls to SetNumber or SetMark will fail; it only impacts
	//Diagram().
	Lock()

	//Unlock 'unlocks' the cell. See Lock for more information on the concept of
	//locking.
	Unlock()

	//The following are private methods
	setPossible(number int)
	setImpossible(number int)
	excludedLock() *sync.RWMutex
	setExcludedBulk(other [DIM]bool)
	excludedBulk() [DIM]bool
	setMarksBulk(other [DIM]bool)
	marksBulk() [DIM]bool
	//To be used only for testing!!!!
	impl() *cellImpl
}

type cellImpl struct {
	gridRef *Grid
	//The number if it's explicitly set. Number() will return it if it's explicitly or implicitly set.
	number          int
	row             int
	col             int
	block           int
	neighborsLock   sync.RWMutex
	neighbors       CellSlice
	impossibles     [DIM]int
	excludedLockRef sync.RWMutex
	excluded        [DIM]bool
	//TODO: do we need a marks lock?
	marks  [DIM]bool
	locked bool
}

func newCell(grid *Grid, row int, col int) cellImpl {
	//TODO: we should not set the number until neighbors are initialized.
	return cellImpl{gridRef: grid, row: row, col: col, block: grid.blockForCell(row, col)}
}

func (self *cellImpl) grid() *Grid {
	return self.gridRef
}

func (self *cellImpl) impl() *cellImpl {
	return self
}

func (self *cellImpl) excludedLock() *sync.RWMutex {
	return &self.excludedLockRef
}

func (self *cellImpl) setExcludedBulk(other [DIM]bool) {
	//Assumes someone else already holds the lock
	for i := 0; i < DIM; i++ {
		self.excluded[i] = other[i]
	}
}

func (self *cellImpl) excludedBulk() [DIM]bool {
	//Assumes someone else already holds the lock
	return self.excluded
}

func (self *cellImpl) setMarksBulk(other [DIM]bool) {
	//Assumes someone else already holds the lock for us
	for i := 0; i < DIM; i++ {
		self.marks[i] = other[i]
	}
}

func (self *cellImpl) marksBulk() [DIM]bool {
	//Assumes someone else already holds the lock
	return self.marks
}

func (self *cellImpl) Mutable() MutableCell {
	return self
}

func (self *cellImpl) Row() int {
	return self.row
}

func (self *cellImpl) Col() int {
	return self.col
}

func (self *cellImpl) Block() int {
	return self.block
}

func (self *cellImpl) MutableInGrid(grid *Grid) MutableCell {
	if grid == nil {
		return nil
	}
	return grid.MutableCell(self.Row(), self.Col())
}

func (self *cellImpl) InGrid(grid *Grid) Cell {
	//Returns our analogue in the given grid.
	if grid == nil {
		return nil
	}
	return grid.Cell(self.Row(), self.Col())
}

func (self *cellImpl) load(data string) {
	//Format, for now, is just the number itself, or 0 if no number.
	data = strings.Replace(data, ALT_0, "0", -1)
	num, _ := strconv.Atoi(data)
	self.SetNumber(num)
}

func (self *cellImpl) Number() int {
	//A layer of indirection since number needs to be used from the Setter.
	return self.number
}

func (self *cellImpl) SetNumber(number int) {
	//Sets the explicit number. This will affect its neighbors possibles list.
	if self.number == number {
		//No work to do now.
		return
	}
	oldNumber := self.number
	self.number = number
	if oldNumber > 0 {
		for i := 1; i <= DIM; i++ {
			if i == oldNumber {
				continue
			}
			self.setPossible(i)
		}
		self.alertNeighbors(oldNumber, true)
	}
	if number > 0 {
		for i := 1; i <= DIM; i++ {
			if i == number {
				continue
			}
			self.setImpossible(i)
		}
		self.alertNeighbors(number, false)
	}
	if self.gridRef != nil {
		self.gridRef.cellModified(self, oldNumber)
		if (oldNumber > 0 && number == 0) || (oldNumber == 0 && number > 0) {
			//Our rank will have changed.
			//TODO: figure out how to test this.
			self.gridRef.cellRankChanged(self)
		}
	}
}

func (self *cellImpl) alertNeighbors(number int, possible bool) {
	for _, cell := range self.MutableNeighbors() {
		if possible {
			cell.setPossible(number)
		} else {
			cell.setImpossible(number)
		}
	}
}

func (self *cellImpl) setPossible(number int) {
	//Number is 1 indexed, but we store it as 0-indexed
	number--
	if number < 0 || number >= DIM {
		return
	}
	if self.impossibles[number] == 0 {
		log.Println("We were told to mark something that was already possible to possible.")
		return
	}
	self.impossibles[number]--
	if self.impossibles[number] == 0 && self.gridRef != nil {
		//TODO: should we check exclusion to save work?
		//Our rank will have changed.
		self.gridRef.cellRankChanged(self)
		//We may have just become valid.
		self.checkInvalid()
	}

}

func (self *cellImpl) setImpossible(number int) {
	//Number is 1 indexed, but we store it as 0-indexed
	number--
	if number < 0 || number >= DIM {
		return
	}
	self.impossibles[number]++
	if self.impossibles[number] == 1 && self.gridRef != nil {
		//TODO: should we check exclusion to save work?
		//Our rank will have changed.
		self.gridRef.cellRankChanged(self)
		//We may have just become invalid.
		self.checkInvalid()
	}
}

func (self *cellImpl) SetExcluded(number int, excluded bool) {
	number--
	if number < 0 || number >= DIM {
		return
	}
	self.excludedLockRef.Lock()
	self.excluded[number] = excluded
	self.excludedLockRef.Unlock()
	//Our rank may have changed.
	//TODO: should we check if we're invalid already?
	if self.gridRef != nil {
		self.gridRef.cellRankChanged(self)
		self.checkInvalid()
	}
}

func (self *cellImpl) ResetExcludes() {
	self.excludedLockRef.Lock()
	for i := 0; i < DIM; i++ {
		self.excluded[i] = false
	}
	self.excludedLockRef.Unlock()
	//Our rank may have changed.
	//TODO: should we check if we're invalid already?
	if self.gridRef != nil {
		self.gridRef.cellRankChanged(self)
		self.checkInvalid()
	}
}

func (self *cellImpl) SetMark(number int, marked bool) {
	number--
	if number < 0 || number >= DIM {
		return
	}
	self.marks[number] = marked
}

func (self *cellImpl) Mark(number int) bool {
	number--
	if number < 0 || number >= DIM {
		return false
	}
	return self.marks[number]
}

func (self *cellImpl) Marks() IntSlice {
	var result IntSlice
	for i := 0; i < DIM; i++ {
		if self.marks[i] {
			result = append(result, i+1)
		}
	}
	return result
}

func (self *cellImpl) ResetMarks() {
	for i := 0; i < DIM; i++ {
		self.marks[i] = false
	}
}

func (self *cellImpl) Possible(number int) bool {
	//Number is 1 indexed, but we store it as 0-indexed
	number--
	if number < 0 || number >= DIM {
		return false
	}
	self.excludedLockRef.RLock()
	isExcluded := self.excluded[number]
	self.excludedLockRef.RUnlock()
	return self.impossibles[number] == 0 && !isExcluded
}

func (self *cellImpl) Possibilities() IntSlice {
	//TODO: performance improvement would be to not need to grab the lock DIM
	//times in Possible and lift it out into a PossibleImpl that has assumes
	//the lock is already grabbed.
	var result IntSlice
	if self.number != 0 {
		return nil
	}
	for i := 1; i <= DIM; i++ {
		if self.Possible(i) {
			result = append(result, i)
		}
	}
	return result
}

func (self *cellImpl) checkInvalid() {
	if self.grid == nil {
		return
	}
	if self.Invalid() {
		self.gridRef.cellIsInvalid(self)
	} else {
		self.gridRef.cellIsValid(self)
	}
}

func (self *cellImpl) Invalid() bool {
	//Returns true if no numbers are possible.
	//TODO: figure out a way to send this back up to the solver when it happens.
	//TODO: shouldn't this always return true if there is a number set?
	for i, counter := range self.impossibles {
		self.excludedLockRef.RLock()
		excluded := self.excluded[i]
		self.excludedLockRef.RUnlock()
		if counter == 0 && !excluded {
			return false
		}
	}
	return true
}

func (self *cellImpl) Lock() {
	self.locked = true
}

func (self *cellImpl) Unlock() {
	self.locked = false
}

func (self *cellImpl) Locked() bool {
	return self.locked
}

func (self *cellImpl) rank() int {
	if self.number != 0 {
		return 0
	}
	count := 0
	for _, counter := range self.impossibles {
		if counter == 0 {
			count++
		}
	}
	return count
}

func (self *cellImpl) ref() cellRef {
	return cellRef{self.Row(), self.Col()}
}

//Sets ourselves to a random one of our possibilities.
func (self *cellImpl) pickRandom() {
	possibilities := self.Possibilities()
	choice := possibilities[rand.Intn(len(possibilities))]
	self.SetNumber(choice)
}

func (self *cellImpl) implicitNumber() int {
	//Impossibles is in 0-index space, but represents nubmers in 1-indexed space.
	result := -1
	for i, counter := range self.impossibles {
		if counter == 0 {
			//Is there someone else competing for this? If so there's no implicit number
			if result != -1 {
				return 0
			}
			result = i
		}
	}
	//convert from 0-indexed to 1-indexed
	return result + 1
}

func (self *cellImpl) MutableSymmetricalPartner(symmetry SymmetryType) MutableCell {
	result := self.SymmetricalPartner(symmetry)

	if result == nil {
		return nil
	}

	return result.Mutable()
}

func (self *cellImpl) SymmetricalPartner(symmetry SymmetryType) Cell {

	if symmetry == SYMMETRY_ANY {
		//TODO: don't chose a type of smmetry that doesn't have a partner
		typesOfSymmetry := []SymmetryType{SYMMETRY_BOTH, SYMMETRY_HORIZONTAL, SYMMETRY_HORIZONTAL, SYMMETRY_VERTICAL}
		symmetry = typesOfSymmetry[rand.Intn(len(typesOfSymmetry))]
	}

	switch symmetry {
	case SYMMETRY_BOTH:
		if cell := self.gridRef.cellImpl(DIM-self.Row()-1, DIM-self.Col()-1); cell != self {
			return cell
		}
	case SYMMETRY_HORIZONTAL:
		if cell := self.gridRef.cellImpl(DIM-self.Row()-1, self.Col()); cell != self {
			return cell
		}
	case SYMMETRY_VERTICAL:
		if cell := self.gridRef.cellImpl(self.Row(), DIM-self.Col()-1); cell != self {
			return cell
		}
	}

	//If the cell was the same as self, or SYMMETRY_NONE
	return nil
}

func (self *cellImpl) MutableNeighbors() MutableCellSlice {
	result := self.Neighbors()
	if result == nil {
		return nil
	}
	return result.mutableCellSlice()
}

func (self *cellImpl) Neighbors() CellSlice {
	if self.gridRef == nil || !self.gridRef.initalized {
		return nil
	}

	self.neighborsLock.RLock()
	neighbors := self.neighbors
	self.neighborsLock.RUnlock()

	if neighbors == nil {
		//We don't want duplicates, so we will collect in a map (used as a set) and then reduce.
		neighborsMap := make(map[Cell]bool)
		for _, cell := range self.gridRef.Row(self.Row()) {
			if cell == self {
				continue
			}
			neighborsMap[cell] = true
		}
		for _, cell := range self.gridRef.Col(self.Col()) {
			if cell == self {
				continue
			}
			neighborsMap[cell] = true
		}
		for _, cell := range self.gridRef.Block(self.Block()) {
			if cell == self {
				continue
			}
			neighborsMap[cell] = true
		}
		self.neighborsLock.Lock()
		self.neighbors = make(CellSlice, len(neighborsMap))
		i := 0
		for cell := range neighborsMap {
			self.neighbors[i] = cell
			i++
		}
		neighbors = self.neighbors
		self.neighborsLock.Unlock()
	}
	return neighbors

}

func (self *cellImpl) dataString() string {
	result := strconv.Itoa(self.Number())
	return strings.Replace(result, "0", ALT_0, -1)
}

func (self *cellImpl) String() string {
	return "Cell[" + strconv.Itoa(self.Row()) + "][" + strconv.Itoa(self.Col()) + "]:" + strconv.Itoa(self.Number()) + "\n"
}

func (self *cellImpl) positionInBlock() (top, right, bottom, left bool) {
	if self.grid == nil {
		return
	}
	topRow, topCol, bottomRow, bottomCol := self.gridRef.blockExtents(self.Block())
	top = self.Row() == topRow
	right = self.Col() == bottomCol
	bottom = self.Row() == bottomRow
	left = self.Col() == topCol
	return
}

func (self *cellImpl) DiagramExtents() (top, left, height, width int) {
	top = self.Col() * (BLOCK_DIM + 1)
	top += self.Col() / BLOCK_DIM

	left = self.Row() * (BLOCK_DIM + 1)
	left += self.Row() / BLOCK_DIM

	return top, left, BLOCK_DIM, BLOCK_DIM

}

func (self *cellImpl) diagramRows(showMarks bool) (rows []string) {
	//We'll only draw barriers at our bottom right edge.
	_, right, bottom, _ := self.positionInBlock()
	current := 0
	for r := 0; r < BLOCK_DIM; r++ {
		row := ""
		for c := 0; c < BLOCK_DIM; c++ {
			if self.number != 0 {
				//Print just the number.
				if r == BLOCK_DIM/2 && c == BLOCK_DIM/2 {
					row += strconv.Itoa(self.number)
				} else {
					if self.Locked() {
						row += DIAGRAM_LOCKED
					} else {
						row += DIAGRAM_NUMBER
					}
				}
			} else {
				//Print the possibles.
				if showMarks {
					if self.Mark(current + 1) {
						row += strconv.Itoa(current + 1)
					} else {
						row += DIAGRAM_IMPOSSIBLE
					}
				} else {
					if self.Possible(current + 1) {
						row += strconv.Itoa(current + 1)
					} else {
						row += DIAGRAM_IMPOSSIBLE
					}
				}
			}
			current++
		}
		rows = append(rows, row)
	}

	//Do we need to pad each row with | on the right?
	if !right {
		for i, data := range rows {
			rows[i] = data + DIAGRAM_RIGHT
		}
	}
	//Do we need an extra bottom row?
	if !bottom {
		rows = append(rows, strings.Repeat(DIAGRAM_BOTTOM, BLOCK_DIM))
		// Does it need a + at the end?
		if !right {
			rows[len(rows)-1] = rows[len(rows)-1] + DIAGRAM_CORNER
		}
	}

	return rows
}

func (self *cellImpl) diagram() string {
	return strings.Join(self.diagramRows(false), "\n")
}
