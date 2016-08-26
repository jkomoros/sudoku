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
//be acted on in various ways. Grid is read-only. For mutator methods, see MutableGrid.
type Grid interface {

	//Copy returns a new grid that has all of the same numbers and excludes filled in it.
	Copy() Grid

	//MutableCopy returns a new, mutable grid that has all of the same numbers
	//and excludes filled in it.
	MutableCopy() MutableGrid

	//CopyWithModifications returns a new Grid that has the given
	//modifications applied.
	CopyWithModifications(modifications GridModifcation) Grid

	//Cells returns a CellSlice with pointers to every cell in the grid,
	//from left to right and top to bottom.
	Cells() CellSlice

	//Row returns a CellSlice containing all of the cells in the given row (0
	//indexed), in order from left to right.
	Row(index int) CellSlice

	//Col returns a CellSlice containing all of the cells in the given column (0
	//indexed), in order from top to bottom.
	Col(index int) CellSlice

	//Block returns a CellSlice containing all of the cells in the given block (0
	//indexed), in order from left to right, top to bottom.
	Block(index int) CellSlice

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

	//HumanSolution returns the SolveDirections that represent how a human
	//would solve this puzzle. If options is nil, will use reasonable
	//defaults.
	HumanSolution(options *HumanSolveOptions) *SolveDirections

	//Hint returns a SolveDirections with precisely one CompoundSolveStep that is
	//a reasonable next step to move the puzzle towards being completed. It is
	//effectively a hint to the user about what Fill step to do next, and why it's
	//logically implied; the truncated return value of HumanSolve. Returns nil if
	//the puzzle has multiple solutions or is otherwise invalid. If options is
	//nil, will use reasonable defaults. optionalPreviousSteps, if provided,
	//serves to help the algorithm pick the most realistic next steps.
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

	//NumSolutions returns the total number of solutions found in the grid when it
	//is solved forward from this point. A valid Sudoku puzzle has only one
	//solution.
	NumSolutions() int

	//HasSolution returns true if the grid has at least one solution.
	HasSolution() bool

	//HasMultipleSolutions returns true if the grid has more than one solution.
	HasMultipleSolutions() bool

	//Solutions returns a slice of grids that represent possible solutions if you
	//were to solve forward this grid. The current grid is not modified. If there
	//are no solutions forward from this location it will return a slice with
	//len() 0.
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
	queue() queue
	numFilledCells() int
	blockForCell(row int, col int) int
	blockExtents(index int) (topRow int, topCol int, bottomRow int, bottomCol int)
	rank() int
	//TODO: this seems like it should be a top-level function that takes a
	//Grid as its first argument.
	searchSolutions(queue *syncedFiniteQueue, isFirstRun bool, numSoughtSolutions int) Grid
}

//TODO: before pushing this, check performance delta. It's bad!

//MutableGrid is a sudoku Grid that can be mutated.
type MutableGrid interface {
	//MutableGrid contains all of Grid's (read-only) methods.
	Grid

	//LoadSDK loads a puzzle in SDK format. Unlike Load, LoadSDK "locks" the cells
	//that are filled. See cell.Lock for more on the concept of locking.
	LoadSDK(data string)
	//TODO: all of these Load methos should be top-level functions that return
	//a Grid

	//Load takes the string data and parses it into the puzzle. The format is the
	//'sdk' format: a `.` marks an empty cell, a number denotes a filled cell, and
	//an (optional) newline marks a new row. Load also accepts other variations on
	//the sdk format, including one with a `|` between each cell. For other sudoku
	//formats see the sdkconverter subpackage.
	Load(data string)

	//LoadSDKFromFile is a simple convenience wrapper around LoadSDK that loads a grid based on the contents
	//of the file at the given path.
	LoadSDKFromFile(path string) bool

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

	//MutableCells returns a MutableCellSlice with pointers to every cell in the
	//grid, from left to right and top to bottom.
	MutableCells() MutableCellSlice

	//MutableRow returns a MutableCellSlice containing all of the cells in the
	//girven row (0 indexed), in order from left to right.
	MutableRow(index int) MutableCellSlice

	//MutableCol returns a MutableCellSlice containing all of the cells in the
	//given column (0 indexed), in order from top to bottom.
	MutableCol(index int) MutableCellSlice

	//MutableBlock returns a MutableCellSlice containing all of the cells in the
	//given block (0 indexed), in order from left to right, top to bottom.
	MutableBlock(index int) MutableCellSlice

	//MutableCell returns a mutable cell. This is a safer operation than
	//grid.Cell(a,b).Mutable() because if you call it on a read-only grid it will
	//fail at compile time as opposed to run time.
	MutableCell(row, col int) MutableCell

	//Fill will find a random filling of the puzzle such that every cell is filled
	//and no cells conflict with their neighbors. If it cannot find one,  it will
	//return false and leave the grid as it found it. Generally you would only
	//want to call this on grids that have more than one solution (e.g. a fully
	//blank grid). Fill provides a good starting point for generated puzzles.
	Fill() bool

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

	//Solve searches for a solution to the puzzle as it currently exists
	//without unfilling any cells. If one exists, it will fill in all cells to
	//fit that solution and return true. If there are no solutions the grid
	//will remain untouched and it will return false. If multiple solutions
	//exist, Solve will pick one at random.
	Solve() bool

	//Private methods
	replace(other MutableGrid)
	initalized() bool
	cellRankChanged(cell MutableCell)
	cellModified(cell MutableCell, oldNumber int)
	cellIsValid(cell MutableCell)
	cellIsInvalid(cell MutableCell)
	cachedSolutionsLock() *sync.RWMutex
	cachedSolutions() []Grid
	cachedSolutionsRequestedLength() int
	//ONLY TO BE USED FOR TESTING!
	impl() *mutableGridImpl
}

//mutableGridImpl is the default implementation of MutableGrid
type mutableGridImpl struct {
	isInitalized bool
	//This is the internal representation only. Having it be a fixed array
	//helps with memory locality and performance. However, iterating over the
	//cells means  that you get a copy, and have to be careful not to try
	//modifying it because the modifications won't work.
	cells                  [DIM * DIM]mutableCellImpl
	rows                   [DIM]CellSlice
	cols                   [DIM]CellSlice
	blocks                 [DIM]CellSlice
	queueGetterLock        sync.RWMutex
	theQueue               *finiteQueue
	numFilledCellsCounter  int
	invalidCells           map[MutableCell]bool
	cachedSolutionsLockRef sync.RWMutex
	cachedSolutionsRef     []Grid
	//The number of solutions that we REQUESTED when we got back
	//this list of cachedSolutions. This helps us avoid extra work
	//in cases where there's only one solution but in the past we'd
	//asked for more.
	cachedSolutionsRequestedLengthRef int
	cachedDifficulty                  float64
}

//gridImpl is the default implementation of Grid.
type gridImpl struct {
	//This structure is designed to be easy to just use copy() and minor fix
	//ups to get a valid copy very quickly--so no pointers.
	cells    [DIM * DIM]cellImpl
	theQueue readOnlyCellQueue
	//TODO: consider whether we should have rows, cols, and blocks cached. On
	//the downside it makes Grid.CopyWithModifications much slower (way more
	//fix up). On the other hand, those might be accessed pretty often...
	filledCellsCount int
	invalid          bool
	solved           bool
}

//TODO:Allow num solver threads to be set at runtime
const _NUM_SOLVER_THREADS = 4

//NewGrid creates a new, blank grid with all of its cells unfilled.
func NewGrid() MutableGrid {
	result := &mutableGridImpl{}

	result.invalidCells = make(map[MutableCell]bool)

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

	result.isInitalized = true
	return result
}

//newStarterGrid is the underlying implementation to create a *gridImpl based
//on a source MutableGrid. The reason it accepts a MutableGrid and not a Grid
//is to reinforce that if you have a Grid and want another Grid you should
//either use Copy() or CopyWithModifications, which are much faster.
func newStarterGrid(grid MutableGrid) *gridImpl {

	//TODO: test this once it actually knows what it's doing!

	result := &gridImpl{
		filledCellsCount: grid.numFilledCells(),
		invalid:          grid.Invalid(),
		solved:           grid.Solved(),
	}

	var cells [DIM * DIM]cellImpl

	for i, sourceCell := range grid.Cells() {
		cells[i] = cellImpl{
			gridRef: result,
			number:  sourceCell.Number(),
			row:     sourceCell.Row(),
			col:     sourceCell.Col(),
			block:   sourceCell.Block(),
			//TODO: actually copy in impossibles, excluded, marks.
			locked: sourceCell.Locked(),
		}
	}

	result.theQueue.grid = result

	i := 0
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			result.theQueue.cellRefs[i] = cellRef{r, c}
			i++
		}
	}

	result.theQueue.fix()

	return result

}

func (self *mutableGridImpl) initalized() bool {
	return self.isInitalized
}

func (self *mutableGridImpl) impl() *mutableGridImpl {
	return self
}

func (self *gridImpl) numFilledCells() int {
	return self.filledCellsCount
}

func (self *mutableGridImpl) numFilledCells() int {
	return self.numFilledCellsCounter
}

func (self *mutableGridImpl) cachedSolutions() []Grid {
	//Assumes someone else holds the lock
	return self.cachedSolutionsRef
}

func (self *mutableGridImpl) cachedSolutionsRequestedLength() int {
	//Assumes someone else holds the lock for us
	return self.cachedSolutionsRequestedLengthRef
}

func (self *gridImpl) queue() queue {
	return &self.theQueue
}

func (self *mutableGridImpl) queue() queue {

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

func (self *mutableGridImpl) cachedSolutionsLock() *sync.RWMutex {
	return &self.cachedSolutionsLockRef
}

func (self *mutableGridImpl) LoadSDK(data string) {
	self.UnlockCells()
	self.Load(data)
	self.LockFilledCells()
}

func (self *mutableGridImpl) Load(data string) {
	//All col separators are basically just to make it easier to read. Remove them.
	data = strings.Replace(data, ALT_COL_SEP, COL_SEP, -1)
	data = strings.Replace(data, COL_SEP, "", -1)
	//TODO: shouldn't we have more error checking, like for wrong dimensions?
	for r, row := range strings.Split(data, ROW_SEP) {
		for c, data := range strings.Split(row, "") {
			cell := self.mutableCellImpl(r, c)
			cell.load(data)
		}
	}
}

func (self *mutableGridImpl) LoadSDKFromFile(path string) bool {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	self.LoadSDK(string(data))
	return true
}

func (self *gridImpl) MutableCopy() MutableGrid {
	//TODO: implement this!
	return nil
}

func (self *mutableGridImpl) MutableCopy() MutableGrid {
	result := NewGrid()
	result.replace(self)
	return result
}

func (self *gridImpl) Copy() Grid {
	//Since it's read-only, no need to actually make a copy.
	return self
}

func (self *mutableGridImpl) Copy() Grid {
	return self.MutableCopy()
}

//Copies the state of the other grid into self, so they look the same.
func (self *mutableGridImpl) replace(other MutableGrid) {
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

func (self *mutableGridImpl) transpose() MutableGrid {
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

func (self *mutableGridImpl) ResetExcludes() {
	for i := range self.cells {
		self.cells[i].ResetExcludes()
	}
}

func (self *mutableGridImpl) ResetMarks() {
	for i := range self.cells {
		self.cells[i].ResetMarks()
	}
}

func (self *mutableGridImpl) ResetUnlockedCells() {
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

func (self *mutableGridImpl) UnlockCells() {
	for i := range self.cells {
		self.cells[i].Unlock()
	}
}

func (self *mutableGridImpl) LockFilledCells() {
	for i := range self.cells {
		cell := &self.cells[i]
		if cell.Number() != 0 {
			cell.Lock()
		}
	}
}

func (self *gridImpl) Cells() CellSlice {
	result := make(CellSlice, len(self.cells))
	for i := range self.cells {
		result[i] = &self.cells[i]
	}
	//TODO: consider caching this?
	return result
}

func (self *mutableGridImpl) Cells() CellSlice {
	return self.MutableCells().cellSlice()
}

func (self *mutableGridImpl) MutableCells() MutableCellSlice {
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
		return nil
	}
	return self.cellSlice(index, 0, index, DIM-1)
}

func (self *mutableGridImpl) Row(index int) CellSlice {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Row: ", index)
		return nil
	}
	return self.rows[index]
}

func (self *mutableGridImpl) MutableRow(index int) MutableCellSlice {
	result := self.Row(index)
	if result == nil {
		return nil
	}
	return result.mutableCellSlice()
}

func (self *gridImpl) Col(index int) CellSlice {
	if index < 0 || index >= DIM {
		return nil
	}
	return self.cellSlice(0, index, DIM-1, index)
}

func (self *mutableGridImpl) Col(index int) CellSlice {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Col: ", index)
		return nil
	}
	return self.cols[index]
}

func (self *mutableGridImpl) MutableCol(index int) MutableCellSlice {
	result := self.Col(index)
	if result == nil {
		return nil
	}
	return result.mutableCellSlice()
}

func (self *gridImpl) Block(index int) CellSlice {
	if index < 0 || index >= DIM {
		return nil
	}
	return self.cellSlice(self.blockExtents(index))
}

func (self *mutableGridImpl) Block(index int) CellSlice {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Block: ", index)
		return nil
	}
	return self.blocks[index]
}

func (self *mutableGridImpl) MutableBlock(index int) MutableCellSlice {
	result := self.Block(index)
	if result == nil {
		return nil
	}
	return result.mutableCellSlice()
}

func gridBlockExtentsImpl(grid Grid, index int) (topRow int, topCol int, bottomRow int, bottomCol int) {
	//Conceptually, we'll pretend like the grid is made up of blocks that are arrayed with row/column
	//Once we find the block r/c, we'll multiply by the actual dim to get the upper left corner.

	blockCol := index % BLOCK_DIM
	blockRow := (index - blockCol) / BLOCK_DIM

	col := blockCol * BLOCK_DIM
	row := blockRow * BLOCK_DIM

	return row, col, row + BLOCK_DIM - 1, col + BLOCK_DIM - 1
}

func (self *gridImpl) blockExtents(index int) (topRow int, topCol int, bottomRow int, bottomCol int) {
	return gridBlockExtentsImpl(self, index)
}

func (self *mutableGridImpl) blockExtents(index int) (topRow int, topCol int, bottomRow int, bottomCol int) {
	return gridBlockExtentsImpl(self, index)
}

func gridBlockForCellImpl(row, col int) int {
	blockCol := col / BLOCK_DIM
	blockRow := row / BLOCK_DIM
	return blockRow*BLOCK_DIM + blockCol
}

func (self *gridImpl) blockForCell(row int, col int) int {
	return gridBlockForCellImpl(row, col)
}

func (self *mutableGridImpl) blockForCell(row int, col int) int {
	return gridBlockForCellImpl(row, col)
}

func (self *mutableGridImpl) blockHasNeighbors(index int) (top bool, right bool, bottom bool, left bool) {
	topRow, topCol, bottomRow, bottomCol := self.blockExtents(index)
	top = topRow != 0
	bottom = bottomRow != DIM-1
	left = topCol != 0
	right = bottomCol != DIM-1
	return
}

func (self *mutableGridImpl) mutableCellImpl(row int, col int) *mutableCellImpl {
	//TODO: can we get rid of this?
	index := row*DIM + col
	if index >= DIM*DIM || index < 0 {
		log.Println("Invalid row/col index passed to Cell: ", row, ", ", col)
		return nil
	}
	return &self.cells[index]
}

func (self *mutableGridImpl) MutableCell(row int, col int) MutableCell {
	return self.mutableCellImpl(row, col)
}

func (self *gridImpl) Cell(row int, col int) Cell {
	index := row*DIM + col
	if index >= DIM*DIM || index < 0 {
		return nil
	}
	return &self.cells[index]
}

func (self *mutableGridImpl) Cell(row int, col int) Cell {
	index := row*DIM + col
	if index >= DIM*DIM || index < 0 {
		log.Println("Invalid row/col index passed to Cell: ", row, ", ", col)
		return nil
	}
	//A first version of this just returned &self.cells[index], but that
	//caused lots of tests to fail. Hmmmm...
	return &self.cells[index].cellImpl
}

func gridCellSliceImpl(grid Grid, rowOne int, colOne int, rowTwo int, colTwo int) CellSlice {
	//both gridImpl and mutableGridImpl can use the same basic implementation
	length := (rowTwo - rowOne + 1) * (colTwo - colOne + 1)
	result := make(CellSlice, length)
	currentRow := rowOne
	currentCol := colOne
	for i := 0; i < length; i++ {
		result[i] = grid.Cell(currentRow, currentCol)
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

func (self *gridImpl) cellSlice(rowOne int, colOne int, rowTwo int, colTwo int) CellSlice {
	return gridCellSliceImpl(self, rowOne, colOne, rowTwo, colTwo)
}

func (self *mutableGridImpl) cellSlice(rowOne int, colOne int, rowTwo int, colTwo int) CellSlice {
	return gridCellSliceImpl(self, rowOne, colOne, rowTwo, colTwo)
}

func (self *gridImpl) Solved() bool {
	return self.solved
}

func (self *mutableGridImpl) Solved() bool {
	//TODO: use numFilledCells here.
	if self.numFilledCellsCounter != len(self.cells) {
		return false
	}
	return !self.Invalid()
}

//We separate this so that we can call it repeatedly within fillSimpleCells,
//and because we know we won't break the more expensive tests.
func (self *mutableGridImpl) cellsInvalid() bool {
	if len(self.invalidCells) > 0 {
		return true
	}
	return false
}

func (self *gridImpl) Invalid() bool {
	return self.invalid
}

func (self *mutableGridImpl) Invalid() bool {
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
	return self.filledCellsCount == 0
}

func (self *mutableGridImpl) Empty() bool {
	return self.numFilledCells() == 0
}

//Called by cells when they notice they are invalid and the grid might not know that.
func (self *mutableGridImpl) cellIsInvalid(cell MutableCell) {
	//Doesn't matter if it was already set.
	self.invalidCells[cell] = true
}

//Called by cells when they notice they are valid and think the grid might not know that.
func (self *mutableGridImpl) cellIsValid(cell MutableCell) {
	delete(self.invalidCells, cell)
}

func (self *mutableGridImpl) cellModified(cell MutableCell, oldNumber int) {
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

func (self *mutableGridImpl) cellRankChanged(cell MutableCell) {
	//We don't want to create the queue if it doesn't exist. But if it does exist we want to get the real one.
	self.queueGetterLock.RLock()
	queue := self.theQueue
	self.queueGetterLock.RUnlock()
	if queue != nil {
		queue.Insert(cell)
	}
}

func (self *gridImpl) rank() int {
	return len(self.cells) - self.filledCellsCount
}

func (self *mutableGridImpl) rank() int {
	return len(self.cells) - self.numFilledCellsCounter
}

func (self *gridImpl) DataString() string {
	//TODO: implement this!
	return ""
}

func (self *mutableGridImpl) DataString() string {
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

func (self *mutableGridImpl) String() string {
	return self.DataString()
}

func gridDiagramImpl(grid Grid, showMarks bool) string {
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
		tempRows = grid.Cell(r, 0).diagramRows(showMarks)
		for c := 1; c < DIM; c++ {
			cellRows := grid.Cell(r, c).diagramRows(showMarks)
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

func (self *gridImpl) Diagram(showMarks bool) string {
	return gridDiagramImpl(self, showMarks)
}

func (self *mutableGridImpl) Diagram(showMarks bool) string {
	return gridDiagramImpl(self, showMarks)
}
