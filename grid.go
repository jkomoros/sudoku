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
type Grid interface {
	//Done marks the grid as ready to be used by another consumer of it. This potentially allows
	//grids to be reused (but not currently).
	Done()

	//LoadSDK loads a puzzle in SDK format. Unlike Load, LoadSDK "locks" the cells
	//that are filled. See cell.Lock for more on the concept of locking.
	LoadSDK(data string)

	//Load takes the string data and parses it into the puzzle. The format is the
	//'sdk' format: a `.` marks an empty cell, a number denotes a filled cell, and
	//an (optional) newline marks a new row. Load also accepts other variations on
	//the sdk format, including one with a `|` between each cell. For other sudoku
	//formats see the sdkconverter subpackage.
	Load(data string)

	//LoadSDKFromFile is a simple convenience wrapper around LoadSDK that loads a grid based on the contents
	//of the file at the given path.
	LoadSDKFromFile(path string) bool

	//Copy returns a new grid that has all of the same numbers and excludes filled in it.
	Copy() Grid

	//ResetExcludes calls ResetExcludes on all cells in the grid. See
	//Cell.SetExcluded for more about excludes.
	ResetExcludes()

	//ResetMarks calls ResetMarks on all cells in the grid. See Cell.SetMark for
	//more about marks.
	ResetMarks()

	//ResetUnlockedCells clears out numbers, marks, and excludes from each cell
	//that is unlocked. In general a locked cell represents a number present in
	//the original puzzle, so this method effectively clears all user
	//modifications back to the start of the puzzle.
	ResetUnlockedCells()

	//UnlockCells unlocks all cells. See cell.Lock for more information on the
	//concept of locking.
	UnlockCells()

	//LockFilledCells locks all cells in the grid that have a number set.
	LockFilledCells()

	//Cells returns a CellSlice with pointers to every cell in the grid,
	//from left to right and top to bottom.
	Cells() CellSlice

	//MutableCells returns a MutableCellSlice with pointers to every cell in the
	//grid, from left to right and top to bottom.
	MutableCells() MutableCellSlice

	//Row returns a CellSlice containing all of the cells in the given row (0
	//indexed), in order from left to right.
	Row(index int) CellSlice

	//MutableRow returns a MutableCellSlice containing all of the cells in the
	//girven row (0 indexed), in order from left to right.
	MutableRow(index int) MutableCellSlice

	//Col returns a CellSlice containing all of the cells in the given column (0
	//indexed), in order from top to bottom.
	Col(index int) CellSlice

	//MutableCol returns a MutableCellSlice containing all of the cells in the
	//given column (0 indexed), in order from top to bottom.
	MutableCol(index int) MutableCellSlice

	//Block returns a CellSlice containing all of the cells in the given block (0
	//indexed), in order from left to right, top to bottom.
	Block(index int) CellSlice

	//MutableBlock returns a MutableCellSlice containing all of the cells in the
	//given block (0 indexed), in order from left to right, top to bottom.
	MutableBlock(index int) MutableCellSlice

	//MutableCell returns a mutable cell. This is a safer operation than
	//grid.Cell(a,b).Mutable() because if you call it on a read-only grid it will
	//fail at compile time as opposed to run time.
	MutableCell(row, col int) MutableCell

	//Cell returns a reference to a specific cell (zero-indexed) in the grid.
	Cell(row, col int) Cell

	//Solved returns true if all cells are filled without violating any
	//constraints; that is, the puzzle is solved.
	Solved() bool

	//Invalid returns true if any numbers are set in the grid that conflict with
	//numbers set in neighborhing cells; when a valid solution cannot be arrived
	//at by continuing to fill additional cells.
	Invalid() bool

	//Empty returns true if none of the grid's cells are filled.
	Empty() bool

	//DataString represents the serialized format of the grid (not including
	//excludes) in canonical sdk format; the output is valid to pass to
	//Grid.Load(). If you want other formats, see the sdkconverter subpackage.
	DataString() string

	//String returns a concise representation of the grid appropriate for printing
	//to the screen. Currently simply an alias for DataString.
	String() string

	//Diagram returns a verbose visual representation of a grid, representing not
	//just filled numbers but also what numbers in a cell are possible. If
	//showMarks is true, instead of printing the possibles, it will print only the
	//activley added marks.
	Diagram(showMarks bool) string

	//Fill will find a random filling of the puzzle such that every cell is filled
	//and no cells conflict with their neighbors. If it cannot find one,  it will
	//return false and leave the grid as it found it. Generally you would only
	//want to call this on grids that have more than one solution (e.g. a fully
	//blank grid). Fill provides a good starting point for generated puzzles.
	Fill() bool

	//HumanSolution returns the SolveDirections that represent how a human would
	//solve this puzzle. It does not mutate the grid. If options is nil, will use
	//reasonable defaults.
	HumanSolution(options *HumanSolveOptions) *SolveDirections

	//HumanSolve is the workhorse of the package. It solves the puzzle much like a
	//human would, applying complex logic techniques iteratively to find a
	//sequence of steps that a reasonable human might apply to solve the puzzle.
	//HumanSolve is an expensive operation because at each step it identifies all
	//of the valid logic rules it could apply and then selects between them based
	//on various weightings. HumanSolve endeavors to find the most realistic human
	//solution it can by using a large number of possible techniques with
	//realistic weights, as well as by doing things like being more likely to pick
	//a cell that is in the same row/cell/block as the last filled cell. Returns
	//nil if the puzzle does not have a single valid solution. If options is nil,
	//will use reasonable defaults. Mutates the grid.
	HumanSolve(options *HumanSolveOptions) *SolveDirections

	//Hint returns a SolveDirections with precisely one CompoundSolveStep that is
	//a reasonable next step to move the puzzle towards being completed. It is
	//effectively a hint to the user about what Fill step to do next, and why it's
	//logically implied; the truncated return value of HumanSolve. Returns nil if
	//the puzzle has multiple solutions or is otherwise invalid. If options is
	//nil, will use reasonable defaults. optionalPreviousSteps, if provided,
	//serves to help the algorithm pick the most realistic next steps. Does not
	//mutate the grid.
	Hint(options *HumanSolveOptions, optionalPreviousSteps []*CompoundSolveStep) *SolveDirections

	//Difficulty returns a value between 0.0 and 1.0, representing how hard the
	//puzzle would be for a human to solve. :This is an EXTREMELY expensive method
	//(although repeated calls without mutating the grid return a cached value
	//quickly). It human solves the puzzle, extracts signals out of the
	//solveDirections, and then passes those signals into a machine-learned model
	//that was trained on hundreds of thousands of solves by real users in order
	//to generate a candidate difficulty. It then repeats the process multiple
	//times until the difficultly number begins to converge to an average.
	Difficulty() float64

	//Solve searches for a solution to the puzzle as it currently exists
	//without unfilling any cells. If one exists, it will fill in all cells to
	//fit that solution and return true. If there are no solutions the grid
	//will remain untouched and it will return false. If multiple solutions
	//exist, Solve will pick one at random.
	Solve() bool

	//NumSolutions returns the total number of solutions found in the grid when it
	//is solved forward from this point. A valid Sudoku puzzle has only one
	//solution. Does not mutate the grid.
	NumSolutions() int

	//HasSolution returns true if the grid has at least one solution. Does not
	//mutate the grid.
	HasSolution() bool

	//HasMultipleSolutions returns true if the grid has more than one solution.
	//Does not mutate the grid.
	HasMultipleSolutions() bool

	//Solutions returns a slice of grids that represent possible solutions if you
	//were to solve forward this grid. The current grid is not modified. If there
	//are no solutions forward from this location it will return a slice with
	//len() 0. Does not mutate the grid.
	Solutions() []Grid

	//HumanSolvePossibleSteps returns a list of CompoundSolveSteps that could
	//apply at this state, along with the probability distribution that a human
	//would pick each one. The optional previousSteps argument is the list of
	//CompoundSolveSteps that have been applied to the grid so far, and is used
	//primarily to tweak the probability distribution and make, for example, it
	//more likely to pick cells in the same block as the cell that was just
	//filled. This method is the workhorse at the core of HumanSolve() and is
	//exposed here primarily so users of this library can get a peek at which
	//possibilites exist at each step. cmd/i-sudoku is one user of this method.
	HumanSolvePossibleSteps(options *HumanSolveOptions, previousSteps []*CompoundSolveStep) (steps []*CompoundSolveStep, distribution ProbabilityDistribution)

	//The rest of these are private methods
	queue() *finiteQueue
	numFilledCells() int
	replace(other Grid)
	cachedSolutionsLock() *sync.RWMutex
	cachedSolutions() []Grid
	cachedSolutionsRequestedLength() int
	blockExtents(index int) (topRow int, topCol int, bottomRow int, bottomCol int)
	rank() int
	searchSolutions(queue *syncedFiniteQueue, isFirstRun bool, numSoughtSolutions int) Grid
	//ONLY TO BE USED FOR TESTING!
	impl() *gridImpl
}

//gridImpl is the default implementation of Grid
type gridImpl struct {
	initalized bool
	//This is the internal representation only. Having it be a fixed array
	//helps with memory locality and performance. However, iterating over the
	//cells means  that you get a copy, and have to be careful not to try
	//modifying it because the modifications won't work.
	cells                  [DIM * DIM]cellImpl
	rows                   [DIM]CellSlice
	cols                   [DIM]CellSlice
	blocks                 [DIM]CellSlice
	queueGetterLock        sync.RWMutex
	theQueue               *finiteQueue
	numFilledCellsCounter  int
	invalidCells           map[*cellImpl]bool
	cachedSolutionsLockRef sync.RWMutex
	cachedSolutionsRef     []Grid
	//The number of solutions that we REQUESTED when we got back
	//this list of cachedSolutions. This helps us avoid extra work
	//in cases where there's only one solution but in the past we'd
	//asked for more.
	cachedSolutionsRequestedLengthRef int
	cachedDifficulty                  float64
}

var gridCache chan Grid

const _MAX_GRIDS = 100

//TODO:Allow num solver threads to be set at runtime
const _NUM_SOLVER_THREADS = 4

func init() {
	gridCache = make(chan Grid, _MAX_GRIDS)
}

func getGrid() Grid {
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

func returnGrid(grid Grid) {
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
func NewGrid() Grid {
	result := &gridImpl{}

	result.invalidCells = make(map[*cellImpl]bool)

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

	result.cachedSolutionsRequestedLengthRef = -1

	result.initalized = true
	return result
}

func (self *gridImpl) impl() *gridImpl {
	return self
}

func (self *gridImpl) numFilledCells() int {
	return self.numFilledCellsCounter
}

func (self *gridImpl) cachedSolutions() []Grid {
	//Assumes someone else holds the lock
	return self.cachedSolutionsRef
}

func (self *gridImpl) cachedSolutionsRequestedLength() int {
	//Assumes someone else holds the lock for us
	return self.cachedSolutionsRequestedLengthRef
}

func (self *gridImpl) queue() *finiteQueue {

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

func (self *gridImpl) cachedSolutionsLock() *sync.RWMutex {
	return &self.cachedSolutionsLockRef
}

func (self *gridImpl) Done() {
	//We're done using this grid; it's okay use to use it again.
	returnGrid(self)
}

func (self *gridImpl) LoadSDK(data string) {
	self.UnlockCells()
	self.Load(data)
	self.LockFilledCells()
}

func (self *gridImpl) Load(data string) {
	//All col separators are basically just to make it easier to read. Remove them.
	data = strings.Replace(data, ALT_COL_SEP, COL_SEP, -1)
	data = strings.Replace(data, COL_SEP, "", -1)
	//TODO: shouldn't we have more error checking, like for wrong dimensions?
	for r, row := range strings.Split(data, ROW_SEP) {
		for c, data := range strings.Split(row, "") {
			cell := self.cellImpl(r, c)
			cell.load(data)
		}
	}
}

func (self *gridImpl) LoadSDKFromFile(path string) bool {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	self.LoadSDK(string(data))
	return true
}

func (self *gridImpl) Copy() Grid {
	//TODO: ideally we'd have some kind of smart SparseGrid or something that we can return.
	result := NewGrid()
	result.replace(self)
	return result
}

//Copies the state of the other grid into self, so they look the same.
func (self *gridImpl) replace(other Grid) {
	//Also set excludes
	for _, otherCell := range other.MutableCells() {
		selfCell := otherCell.MutableInGrid(self)

		selfCell.SetNumber(otherCell.Number())
		//TODO: the fact that I'm reaching into Cell's excludeLock outside of Cell is a Smell.
		selfCell.excludedLock().Lock()
		otherCell.excludedLock().RLock()

		//TODO: it's conceivable that this extra copying is what's taking 15%
		//longer.
		selfCell.setExcludedBulk(otherCell.excludedBulk())
		selfCell.setMarksBulk(otherCell.marksBulk())

		otherCell.excludedLock().RUnlock()
		selfCell.excludedLock().Unlock()
		if otherCell.Locked() {
			selfCell.Lock()
		} else {
			selfCell.Unlock()
		}
	}
	self.cachedSolutionsLock().Lock()
	other.cachedSolutionsLock().RLock()
	self.cachedSolutionsRequestedLengthRef = other.cachedSolutionsRequestedLength()
	self.cachedSolutionsRef = other.cachedSolutions()
	other.cachedSolutionsLock().RUnlock()
	self.cachedSolutionsLock().Unlock()
}

func (self *gridImpl) transpose() Grid {
	//Returns a new grid that is the same as this grid (ignoring overrides, which are nulled), but with rows and cols swapped.
	result := NewGrid()
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			original := self.MutableCell(r, c)
			copy := result.MutableCell(c, r)
			copy.SetNumber(original.Number())
			//TODO: shouldn't we have a lock here or something?
			copy.setExcludedBulk(original.excludedBulk())
		}
	}
	return result
}

func (self *gridImpl) ResetExcludes() {
	for i := range self.cells {
		self.cells[i].ResetExcludes()
	}
}

func (self *gridImpl) ResetMarks() {
	for i := range self.cells {
		self.cells[i].ResetMarks()
	}
}

func (self *gridImpl) ResetUnlockedCells() {
	for i := range self.cells {
		cell := &self.cells[i]
		if cell.Locked() {
			continue
		}
		cell.SetNumber(0)
		cell.ResetMarks()
		cell.ResetExcludes()
	}
}

func (self *gridImpl) UnlockCells() {
	for i := range self.cells {
		self.cells[i].Unlock()
	}
}

func (self *gridImpl) LockFilledCells() {
	for i := range self.cells {
		cell := &self.cells[i]
		if cell.Number() != 0 {
			cell.Lock()
		}
	}
}

func (self *gridImpl) Cells() CellSlice {
	return self.MutableCells().cellSlice()
}

func (self *gridImpl) MutableCells() MutableCellSlice {
	//Returns a CellSlice of all of the cells in order.
	result := make(MutableCellSlice, len(self.cells))
	for i := range self.cells {
		//We don't use the second argument of range because that would be a copy of the cell, not the real one.
		result[i] = &self.cells[i]
	}
	//TODO: cache this result
	return result
}

func (self *gridImpl) Row(index int) CellSlice {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Row: ", index)
		return nil
	}
	return self.rows[index]
}

func (self *gridImpl) MutableRow(index int) MutableCellSlice {
	result := self.Row(index)
	if result == nil {
		return nil
	}
	return result.mutableCellSlice()
}

func (self *gridImpl) Col(index int) CellSlice {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Col: ", index)
		return nil
	}
	return self.cols[index]
}

func (self *gridImpl) MutableCol(index int) MutableCellSlice {
	result := self.Col(index)
	if result == nil {
		return nil
	}
	return result.mutableCellSlice()
}

func (self *gridImpl) Block(index int) CellSlice {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Block: ", index)
		return nil
	}
	return self.blocks[index]
}

func (self *gridImpl) MutableBlock(index int) MutableCellSlice {
	result := self.Block(index)
	if result == nil {
		return nil
	}
	return result.mutableCellSlice()
}

func (self *gridImpl) blockExtents(index int) (topRow int, topCol int, bottomRow int, bottomCol int) {
	//Conceptually, we'll pretend like the grid is made up of blocks that are arrayed with row/column
	//Once we find the block r/c, we'll multiply by the actual dim to get the upper left corner.

	blockCol := index % BLOCK_DIM
	blockRow := (index - blockCol) / BLOCK_DIM

	col := blockCol * BLOCK_DIM
	row := blockRow * BLOCK_DIM

	return row, col, row + BLOCK_DIM - 1, col + BLOCK_DIM - 1
}

func (self *gridImpl) blockForCell(row int, col int) int {
	blockCol := col / BLOCK_DIM
	blockRow := row / BLOCK_DIM
	return blockRow*BLOCK_DIM + blockCol
}

func (self *gridImpl) blockHasNeighbors(index int) (top bool, right bool, bottom bool, left bool) {
	topRow, topCol, bottomRow, bottomCol := self.blockExtents(index)
	top = topRow != 0
	bottom = bottomRow != DIM-1
	left = topCol != 0
	right = bottomCol != DIM-1
	return
}

//cellImpl is required because some clients in the package need the actual
//underlying pointer for comparison.
func (self *gridImpl) cellImpl(row int, col int) *cellImpl {
	index := row*DIM + col
	if index >= DIM*DIM || index < 0 {
		log.Println("Invalid row/col index passed to Cell: ", row, ", ", col)
		return nil
	}
	return &self.cells[index]
}

func (self *gridImpl) MutableCell(row int, col int) MutableCell {
	return self.cellImpl(row, col)
}

func (self *gridImpl) Cell(row int, col int) Cell {
	return self.cellImpl(row, col)
}

func (self *gridImpl) cellSlice(rowOne int, colOne int, rowTwo int, colTwo int) CellSlice {
	length := (rowTwo - rowOne + 1) * (colTwo - colOne + 1)
	result := make(CellSlice, length)
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

func (self *gridImpl) Solved() bool {
	//TODO: use numFilledCells here.
	if self.numFilledCellsCounter != len(self.cells) {
		return false
	}
	return !self.Invalid()
}

//We separate this so that we can call it repeatedly within fillSimpleCells,
//and because we know we won't break the more expensive tests.
func (self *gridImpl) cellsInvalid() bool {
	if len(self.invalidCells) > 0 {
		return true
	}
	return false
}

func (self *gridImpl) Invalid() bool {
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

func (self *gridImpl) Empty() bool {
	return self.numFilledCells() == 0
}

//Called by cells when they notice they are invalid and the grid might not know that.
func (self *gridImpl) cellIsInvalid(cell *cellImpl) {
	//Doesn't matter if it was already set.
	self.invalidCells[cell] = true
}

//Called by cells when they notice they are valid and think the grid might not know that.
func (self *gridImpl) cellIsValid(cell *cellImpl) {
	delete(self.invalidCells, cell)
}

func (self *gridImpl) cellModified(cell *cellImpl, oldNumber int) {
	self.cachedSolutionsLock().Lock()
	self.cachedSolutionsRef = nil
	self.cachedSolutionsRequestedLengthRef = -1
	self.cachedSolutionsLock().Unlock()
	self.cachedDifficulty = 0.0

	if cell.Number() == 0 && oldNumber != 0 {
		self.numFilledCellsCounter--
	} else if cell.Number() != 0 && oldNumber == 0 {
		self.numFilledCellsCounter++
	}

}

func (self *gridImpl) cellRankChanged(cell *cellImpl) {
	//We don't want to create the queue if it doesn't exist. But if it does exist we want to get the real one.
	self.queueGetterLock.RLock()
	queue := self.theQueue
	self.queueGetterLock.RUnlock()
	if queue != nil {
		queue.Insert(cell)
	}
}

func (self *gridImpl) rank() int {
	return len(self.cells) - self.numFilledCellsCounter
}

func (self *gridImpl) DataString() string {
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

func (self *gridImpl) String() string {
	return self.DataString()
}

func (self *gridImpl) Diagram(showMarks bool) string {
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
