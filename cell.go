package sudoku

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
)

const NUM_NEIGHBORS = (DIM-1)*3 - (BLOCK_DIM-1)*2

type SymmetryType int

const (
	SYMMETRY_NONE SymmetryType = iota
	SYMMETRY_ANY
	SYMMETRY_HORIZONTAL
	SYMMETRY_VERTICAL
	SYMMETRY_BOTH
)

//Cell represents a single cell within a grid. It maintains information about the number that is filled, the numbers that are
//currently legal given the filled status of its neighbors, and whether any possibilities have been
//explicitly excluded by solve techniques. Cells should not be constructed on their own; create a Grid
//and grab references to the cells from there.
type Cell struct {
	grid *Grid
	//The number if it's explicitly set. Number() will return it if it's explicitly or implicitly set.
	number      int
	Row         int
	Col         int
	Block       int
	neighbors   CellList
	impossibles [DIM]int
	excluded    [DIM]bool
}

func newCell(grid *Grid, row int, col int) Cell {
	//TODO: we should not set the number until neighbors are initialized.
	return Cell{grid: grid, Row: row, Col: col, Block: grid.blockForCell(row, col)}
}

//InGrid returns a reference to a cell in the provided grid that has the same row/column as this cell.
//Effectively, this cell's analogue in the other grid.
func (self *Cell) InGrid(grid *Grid) *Cell {
	//Returns our analogue in the given grid.
	if grid == nil {
		return nil
	}
	return grid.Cell(self.Row, self.Col)
}

func (self *Cell) load(data string) {
	//Format, for now, is just the number itself, or 0 if no number.
	data = strings.Replace(data, ALT_0, "0", -1)
	num, _ := strconv.Atoi(data)
	self.SetNumber(num)
}

//Number returns the number the cell is currently set to.
func (self *Cell) Number() int {
	//A layer of indirection since number needs to be used from the Setter.
	return self.number
}

//SetNumber explicitly sets the number of the cell. This operation could cause the grid to become
//invalid if it conflicts with its neighbors' numbers. This operation will affect the Possiblities()
//of its neighbor cells.
func (self *Cell) SetNumber(number int) {
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
	if self.grid != nil {
		self.grid.cellModified(self)
		if (oldNumber > 0 && number == 0) || (oldNumber == 0 && number > 0) {
			//Our rank will have changed.
			//TODO: figure out how to test this.
			self.grid.cellRankChanged(self)
		}
	}
}

func (self *Cell) alertNeighbors(number int, possible bool) {
	for _, cell := range self.Neighbors() {
		if possible {
			cell.setPossible(number)
		} else {
			cell.setImpossible(number)
		}
	}
}

func (self *Cell) setPossible(number int) {
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
	if self.impossibles[number] == 0 && self.grid != nil {
		//TODO: should we check exclusion to save work?
		//Our rank will have changed.
		self.grid.cellRankChanged(self)
		//We may have just become valid.
		self.checkInvalid()
	}

}

func (self *Cell) setImpossible(number int) {
	//Number is 1 indexed, but we store it as 0-indexed
	number--
	if number < 0 || number >= DIM {
		return
	}
	self.impossibles[number]++
	if self.impossibles[number] == 1 && self.grid != nil {
		//TODO: should we check exclusion to save work?
		//Our rank will have changed.
		self.grid.cellRankChanged(self)
		//We may have just become invalid.
		self.checkInvalid()
	}
}

//SetExcluded defines whether a possibility is considered not feasible, even if not directly precluded
//by the Number()s of the cell's neighbors. This is used by advanced HumanSolve techniques that
//cull possibilities that are logically excluded by the state of the grid, in a non-direct way.
//The state of Excluded bits will affect the results of this cell's Possibilities() list.
func (self *Cell) SetExcluded(number int, excluded bool) {
	number--
	if number < 0 || number >= DIM {
		return
	}
	self.excluded[number] = excluded
	//Our rank may have changed.
	//TODO: should we check if we're invalid already?
	if self.grid != nil {
		self.grid.cellRankChanged(self)
		self.checkInvalid()
	}
}

//ResetExcludes sets all excluded bits to false, so that Possibilities() will be based purely
//on direct implications of the Number()s of neighbors. See also SetExcluded.
func (self *Cell) ResetExcludes() {
	for i := 0; i < DIM; i++ {
		self.excluded[i] = false
	}
	//Our rank may have changed.
	//TODO: should we check if we're invalid already?
	if self.grid != nil {
		self.grid.cellRankChanged(self)
		self.checkInvalid()
	}
}

//Possible returns whether or not a given number is legal to fill via SetNumber, given the state of the grid (specifically,
//the cell's neighbors) and the numbers the cell was told to explicitly exclude via SetExclude.
//If the cell is already filled with a number, it will return false for all numbers.
func (self *Cell) Possible(number int) bool {
	//Number is 1 indexed, but we store it as 0-indexed
	number--
	if number < 0 || number >= DIM {
		return false
	}
	return self.impossibles[number] == 0 && !self.excluded[number]
}

//Possibilities returns a list of all current possibilities for this cell: all numbers for which cell.Possible
//returns true.
func (self *Cell) Possibilities() (result []int) {
	//tODO: shouldn't this return an intslice?
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

func (self *Cell) checkInvalid() {
	if self.grid == nil {
		return
	}
	if self.Invalid() {
		self.grid.cellIsInvalid(self)
	} else {
		self.grid.cellIsValid(self)
	}
}

func (self *Cell) Invalid() bool {
	//Returns true if no numbers are possible.
	//TODO: figure out a way to send this back up to the solver when it happens.
	for i, counter := range self.impossibles {
		if counter == 0 && !self.excluded[i] {
			return false
		}
	}
	return true
}

func (self *Cell) rank() int {
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

func (self *Cell) ref() cellRef {
	return cellRef{self.Row, self.Col}
}

//Sets ourselves to a random one of our possibilities.
func (self *Cell) pickRandom() {
	possibilities := self.Possibilities()
	choice := possibilities[rand.Intn(len(possibilities))]
	self.SetNumber(choice)
}

func (self *Cell) implicitNumber() int {
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

func (self *Cell) SymmetricalPartner(symmetry SymmetryType) *Cell {

	if symmetry == SYMMETRY_ANY {
		//TODO: don't chose a type of smmetry that doesn't have a partner
		typesOfSymmetry := []SymmetryType{SYMMETRY_BOTH, SYMMETRY_HORIZONTAL, SYMMETRY_HORIZONTAL, SYMMETRY_VERTICAL}
		symmetry = typesOfSymmetry[rand.Intn(len(typesOfSymmetry))]
	}

	switch symmetry {
	case SYMMETRY_BOTH:
		if cell := self.grid.Cell(DIM-self.Row-1, DIM-self.Col-1); cell != self {
			return cell
		}
	case SYMMETRY_HORIZONTAL:
		if cell := self.grid.Cell(DIM-self.Row-1, self.Col); cell != self {
			return cell
		}
	case SYMMETRY_VERTICAL:
		if cell := self.grid.Cell(self.Row, DIM-self.Col-1); cell != self {
			return cell
		}
	}

	//If the cell was the same as self, or SYMMETRY_NONE
	return nil
}

func (self *Cell) Neighbors() CellList {
	if self.grid == nil || !self.grid.initalized {
		return nil
	}
	if self.neighbors == nil {
		//We don't want duplicates, so we will collect in a map (used as a set) and then reduce.
		neighborsMap := make(map[*Cell]bool)
		for _, cell := range self.grid.Row(self.Row) {
			if cell == self {
				continue
			}
			neighborsMap[cell] = true
		}
		for _, cell := range self.grid.Col(self.Col) {
			if cell == self {
				continue
			}
			neighborsMap[cell] = true
		}
		for _, cell := range self.grid.Block(self.Block) {
			if cell == self {
				continue
			}
			neighborsMap[cell] = true
		}
		self.neighbors = make([]*Cell, len(neighborsMap))
		i := 0
		for cell, _ := range neighborsMap {
			self.neighbors[i] = cell
			i++
		}
	}
	return self.neighbors

}

func (self *Cell) DataString() string {
	result := strconv.Itoa(self.Number())
	return strings.Replace(result, "0", ALT_0, -1)
}

func (self *Cell) String() string {
	return "Cell[" + strconv.Itoa(self.Row) + "][" + strconv.Itoa(self.Col) + "]:" + strconv.Itoa(self.Number()) + "\n"
}

func (self *Cell) positionInBlock() (top, right, bottom, left bool) {
	if self.grid == nil {
		return
	}
	topRow, topCol, bottomRow, bottomCol := self.grid.blockExtents(self.Block)
	top = self.Row == topRow
	right = self.Col == bottomCol
	bottom = self.Row == bottomRow
	left = self.Col == topCol
	return
}

func (self *Cell) diagramRows() (rows []string) {
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
					row += DIAGRAM_NUMBER
				}
			} else {
				//Print the possibles.
				if self.Possible(current + 1) {
					row += strconv.Itoa(current + 1)
				} else {
					row += DIAGRAM_IMPOSSIBLE
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

func (self *Cell) Diagram() string {
	return strings.Join(self.diagramRows(), "\n")
}
