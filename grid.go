package sudoku

import (
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

//BLOCK_DIM is the height and width of each block within the grid.
const BLOCK_DIM = 3

//DIM is the dimension of the grid (the height and width)
const DIM = BLOCK_DIM * BLOCK_DIM

//Constants for important aspects of the accepted format
const (
	ALT_0       = "."
	ROW_SEP     = "\n"
	COL_SEP     = "|"
	ALT_COL_SEP = "||"
)

//Constants for how important parts of the diagram are printed out
const (
	DIAGRAM_IMPOSSIBLE = " "
	DIAGRAM_RIGHT      = "|"
	DIAGRAM_BOTTOM     = "-"
	DIAGRAM_CORNER     = "+"
	DIAGRAM_NUMBER     = "•"
	DIAGRAM_LOCKED     = "X"
)

//Grid is the primary type in the package. It represents a DIMxDIM sudoku puzzle that can
//be acted on in various ways.
type Grid struct {
	initalized bool
	//This is the internal representation only. Having it be a fixed array
	//helps with memory locality and performance. However, iterating over the
	//cells means  that you get a copy, and have to be careful not to try
	//modifying it because the modifications won't work.
	cells               [DIM * DIM]Cell
	rows                [DIM]CellSlice
	cols                [DIM]CellSlice
	blocks              [DIM]CellSlice
	queueGetterLock     sync.RWMutex
	theQueue            *finiteQueue
	numFilledCells      int
	invalidCells        map[*Cell]bool
	cachedSolutionsLock sync.RWMutex
	cachedSolutions     []*Grid
	//The number of solutions that we REQUESTED when we got back
	//this list of cachedSolutions. This helps us avoid extra work
	//in cases where there's only one solution but in the past we'd
	//asked for more.
	cachedSolutionsRequestedLength int
	cachedDifficulty               float64
}

var gridCache chan *Grid

const _MAX_GRIDS = 100

//TODO:Allow num solver threads to be set at runtime
const _NUM_SOLVER_THREADS = 4

func init() {
	gridCache = make(chan *Grid, _MAX_GRIDS)
}

func getGrid() *Grid {
	select {
	case grid := <-gridCache:
		return grid
	default:
		return NewGrid()
	}
	return nil
}

func dropGrids() {
	for {
		select {
		case <-gridCache:
			//Keep on going
		default:
			return
		}
	}
}

func returnGrid(grid *Grid) {
	grid.ResetExcludes()
	grid.ResetMarks()
	select {
	case gridCache <- grid:
		//Returned it to the queue.
	default:
		//Drop it on the floor.
	}
}

//NewGrid creates a new, blank grid with all of its cells unfilled.
func NewGrid() *Grid {
	result := &Grid{}

	result.invalidCells = make(map[*Cell]bool)

	i := 0
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			result.cells[i] = newCell(result, r, c)
			//The cell can't insert itself because it doesn't know where it will actually live in memory yet.
			i++
		}
	}

	for index := 0; index < DIM; index++ {
		result.rows[index] = result.cellSlice(index, 0, index, DIM-1)
		result.cols[index] = result.cellSlice(0, index, DIM-1, index)
		result.blocks[index] = result.cellSlice(result.blockExtents(index))
	}

	result.cachedSolutionsRequestedLength = -1

	result.initalized = true
	return result
}

func (self *Grid) queue() *finiteQueue {

	self.queueGetterLock.RLock()
	queue := self.theQueue
	self.queueGetterLock.RUnlock()

	if queue == nil {
		self.queueGetterLock.Lock()
		self.theQueue = newFiniteQueue(1, DIM)
		for i := range self.cells {
			//If we did i, cell, cell would just be the temp variable. So we'll grab it via the index.
			self.theQueue.Insert(&self.cells[i])
		}
		queue = self.theQueue
		self.queueGetterLock.Unlock()
	}
	return queue
}

//Done marks the grid as ready to be used by another consumer of it. This potentially allows
//grids to be reused (but not currently).
func (self *Grid) Done() {
	//We're done using this grid; it's okay use to use it again.
	returnGrid(self)
}

//Load takes the string data and parses it into the puzzle. The format is the 'sdk' format:
//a `.` marks an empty cell, a number denotes a filled cell, and a newline marks a new row.
//Load also accepts other variations on the sdk format, including one with a `|` between each cell.
//For other sudoku formats see the sdkconverter subpackage.
func (self *Grid) Load(data string) {
	//All col separators are basically just to make it easier to read. Remove them.
	data = strings.Replace(data, ALT_COL_SEP, COL_SEP, -1)
	data = strings.Replace(data, COL_SEP, "", -1)
	//TODO: shouldn't we have more error checking, like for wrong dimensions?
	for r, row := range strings.Split(data, ROW_SEP) {
		for c, data := range strings.Split(row, "") {
			cell := self.Cell(r, c)
			cell.load(data)
		}
	}
}

//LoadFromFile is a simple convenience wrapper around Load that loads a grid based on the contents
//of the file at the given path.
func (self *Grid) LoadFromFile(path string) bool {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	self.Load(string(data))
	return true
}

//Copy returns a new grid that has all of the same numbers and excludes filled in it.
func (self *Grid) Copy() *Grid {
	//TODO: ideally we'd have some kind of smart SparseGrid or something that we can return.
	result := NewGrid()
	result.replace(self)
	return result
}

//Copies the state of the other grid into self, so they look the same.
func (self *Grid) replace(other *Grid) {
	self.Load(other.DataString())
	//Also set excludes
	for index, _ := range other.cells {
		otherCell := &other.cells[index]
		selfCell := otherCell.InGrid(self)
		//TODO: the fact that I'm reaching into Cell's excludeLock outside of Cell is a Smell.
		selfCell.excludedLock.Lock()
		otherCell.excludedLock.RLock()
		for i := 0; i < DIM; i++ {
			selfCell.excluded[i] = otherCell.excluded[i]
		}
		otherCell.excludedLock.RUnlock()
		selfCell.excludedLock.Unlock()
	}
	self.cachedSolutionsLock.Lock()
	other.cachedSolutionsLock.RLock()
	self.cachedSolutionsRequestedLength = other.cachedSolutionsRequestedLength
	self.cachedSolutions = other.cachedSolutions
	other.cachedSolutionsLock.RUnlock()
	self.cachedSolutionsLock.Unlock()
}

func (self *Grid) transpose() *Grid {
	//Returns a new grid that is the same as this grid (ignoring overrides, which are nulled), but with rows and cols swapped.
	result := NewGrid()
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			original := self.Cell(r, c)
			copy := result.Cell(c, r)
			copy.SetNumber(original.Number())
			//This should be a copy since it's an array, right?
			copy.excluded = original.excluded
		}
	}
	return result
}

//ResetExcludes calls ResetExcludes on all cells in the grid. See Cell.SetExcluded for more about excludes.
func (self *Grid) ResetExcludes() {
	for i := range self.cells {
		self.cells[i].ResetExcludes()
	}
}

//ResetMarks calls ResetMarks on all cells in the grid. See Cell.SetMark for
//more about marks.
func (self *Grid) ResetMarks() {
	for i := range self.cells {
		self.cells[i].ResetMarks()
	}
}

//Cells returns a CellSlice with pointers to every cell in the grid,
//from left to right and top to bottom.
func (self *Grid) Cells() CellSlice {
	//Returns a CellSlice of all of the cells in order.
	result := make(CellSlice, len(self.cells))
	for i, _ := range self.cells {
		//We don't use the second argument of range because that would be a copy of the cell, not the real one.
		result[i] = &self.cells[i]
	}
	//TODO: cache this result
	return result
}

//Row returns a CellSlice containing all of the cells in the given row (0 indexed), in order from left to right.
func (self *Grid) Row(index int) CellSlice {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Row: ", index)
		return nil
	}
	return self.rows[index]
}

//Col returns a CellSlice containing all of the cells in the given column (0 indexed), in order from top to bottom.
func (self *Grid) Col(index int) CellSlice {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Col: ", index)
		return nil
	}
	return self.cols[index]
}

//Block returns a CellSlice containing all of the cells in the given block (0 indexed), in order from left to right, top to bottom.
func (self *Grid) Block(index int) CellSlice {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Block: ", index)
		return nil
	}
	return self.blocks[index]
}

func (self *Grid) blockExtents(index int) (topRow int, topCol int, bottomRow int, bottomCol int) {
	//Conceptually, we'll pretend like the grid is made up of blocks that are arrayed with row/column
	//Once we find the block r/c, we'll multiply by the actual dim to get the upper left corner.

	blockCol := index % BLOCK_DIM
	blockRow := (index - blockCol) / BLOCK_DIM

	col := blockCol * BLOCK_DIM
	row := blockRow * BLOCK_DIM

	return row, col, row + BLOCK_DIM - 1, col + BLOCK_DIM - 1
}

func (self *Grid) blockForCell(row int, col int) int {
	blockCol := col / BLOCK_DIM
	blockRow := row / BLOCK_DIM
	return blockRow*BLOCK_DIM + blockCol
}

func (self *Grid) blockHasNeighbors(index int) (top bool, right bool, bottom bool, left bool) {
	topRow, topCol, bottomRow, bottomCol := self.blockExtents(index)
	top = topRow != 0
	bottom = bottomRow != DIM-1
	left = topCol != 0
	right = bottomCol != DIM-1
	return
}

//Cell returns a reference to a specific cell (zero-indexed) in the grid.
func (self *Grid) Cell(row int, col int) *Cell {
	index := row*DIM + col
	if index >= DIM*DIM || index < 0 {
		log.Println("Invalid row/col index passed to Cell: ", row, ", ", col)
		return nil
	}
	return &self.cells[index]
}

func (self *Grid) cellSlice(rowOne int, colOne int, rowTwo int, colTwo int) CellSlice {
	length := (rowTwo - rowOne + 1) * (colTwo - colOne + 1)
	result := make([]*Cell, length)
	currentRow := rowOne
	currentCol := colOne
	for i := 0; i < length; i++ {
		result[i] = self.Cell(currentRow, currentCol)
		if colTwo > currentCol {
			currentCol++
		} else {
			if rowTwo > currentRow {
				currentRow++
				currentCol = colOne
			} else {
				//This should only happen the last time through the loop.
			}
		}
	}
	return CellSlice(result)
}

//Solved returns true if all cells are filled without violating any constraints; that is, the puzzle is solved.
func (self *Grid) Solved() bool {
	//TODO: use numFilledCells here.
	if self.numFilledCells != len(self.cells) {
		return false
	}
	return !self.Invalid()
}

//We separate this so that we can call it repeatedly within fillSimpleCells, and because we know we won't break the more expensive tests.
func (self *Grid) cellsInvalid() bool {
	if len(self.invalidCells) > 0 {
		return true
	}
	return false
}

//Invalid returns true if any numbers are set in the grid that conflict with numbers set in neighborhing cells;
//when a valid solution cannot be arrived at by continuing to fill additional cells.
func (self *Grid) Invalid() bool {
	//Grid will never be invalid based on moves made by the solver; it will detect times that
	//someone called SetNumber with an impossible number after the fact, though.

	if self.cellsInvalid() {
		return true
	}
	for i := 0; i < DIM; i++ {
		row := self.Row(i)
		rowCheck := make(map[int]bool)
		for _, cell := range row {
			if cell.Number() == 0 {
				continue
			}
			if rowCheck[cell.Number()] {
				return true
			}
			rowCheck[cell.Number()] = true
		}
		col := self.Col(i)
		colCheck := make(map[int]bool)
		for _, cell := range col {
			if cell.Number() == 0 {
				continue
			}
			if colCheck[cell.Number()] {
				return true
			}
			colCheck[cell.Number()] = true
		}
		block := self.Block(i)
		blockCheck := make(map[int]bool)
		for _, cell := range block {
			if cell.Number() == 0 {
				continue
			}
			if blockCheck[cell.Number()] {
				return true
			}
			blockCheck[cell.Number()] = true
		}
	}
	return false
}

//Returns true if none of the grid's cells are filled.
func (self *Grid) Empty() bool {
	return self.numFilledCells == 0
}

//Called by cells when they notice they are invalid and the grid might not know that.
func (self *Grid) cellIsInvalid(cell *Cell) {
	//Doesn't matter if it was already set.
	self.invalidCells[cell] = true
}

//Called by cells when they notice they are valid and think the grid might not know that.
func (self *Grid) cellIsValid(cell *Cell) {
	delete(self.invalidCells, cell)
}

func (self *Grid) cellModified(cell *Cell) {
	self.cachedSolutionsLock.Lock()
	self.cachedSolutions = nil
	self.cachedSolutionsRequestedLength = -1
	self.cachedSolutionsLock.Unlock()
	self.cachedDifficulty = 0.0
	if cell.Number() == 0 {
		self.numFilledCells--
	} else {
		self.numFilledCells++
	}
}

func (self *Grid) cellRankChanged(cell *Cell) {
	//We don't want to create the queue if it doesn't exist. But if it does exist we want to get the real one.
	self.queueGetterLock.RLock()
	queue := self.theQueue
	self.queueGetterLock.RUnlock()
	if queue != nil {
		queue.Insert(cell)
	}
}

func (self *Grid) rank() int {
	return len(self.cells) - self.numFilledCells
}

//DataString represents the serialized format of the grid (not including excludes) in canonical sdk format; the output
//is valid to pass to Grid.Load(). If you want other formats, see the sdkconverter subpackage.
func (self *Grid) DataString() string {
	var rows []string
	for r := 0; r < DIM; r++ {
		var row []string
		for c := 0; c < DIM; c++ {
			row = append(row, self.cells[r*DIM+c].dataString())
		}
		rows = append(rows, strings.Join(row, COL_SEP))
	}
	return strings.Join(rows, ROW_SEP)
}

//String returns a concise representation of the grid appropriate for printing to the screen.
//Currently simply an alias for DataString.
func (self *Grid) String() string {
	return self.DataString()
}

//Diagram returns a verbose visual representation of a grid, representing not
//just filled numbers but also what numbers in a cell are possible. If
//showMarks is true, instead of printing the possibles, it will print only the
//activley added marks.
func (self *Grid) Diagram(showMarks bool) string {
	var rows []string

	//Generate a block boundary row to use later.
	var blockBoundaryRow string
	{
		var blockBoundarySegments []string
		blockBoundaryBlockPiece := strings.Repeat(DIAGRAM_BOTTOM, BLOCK_DIM*BLOCK_DIM+BLOCK_DIM-1)
		for i := 0; i < BLOCK_DIM; i++ {
			blockBoundarySegments = append(blockBoundarySegments, blockBoundaryBlockPiece)
		}
		blockBoundaryRow = strings.Join(blockBoundarySegments, DIAGRAM_CORNER+DIAGRAM_CORNER)
	}

	for r := 0; r < DIM; r++ {
		var tempRows []string
		tempRows = self.Cell(r, 0).diagramRows(showMarks)
		for c := 1; c < DIM; c++ {
			cellRows := self.Cell(r, c).diagramRows(showMarks)
			for i := range tempRows {
				tempRows[i] += cellRows[i]
				//Are we at a block boundary?
				if c%BLOCK_DIM == BLOCK_DIM-1 && c != DIM-1 {
					tempRows[i] += DIAGRAM_RIGHT + DIAGRAM_RIGHT
				}
			}

		}
		rows = append(rows, tempRows...)
		//Are we at a block boundary?
		if r%BLOCK_DIM == BLOCK_DIM-1 && r != DIM-1 {
			rows = append(rows, blockBoundaryRow, blockBoundaryRow)
		}
	}
	return strings.Join(rows, "\n")
}
