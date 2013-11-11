package sudoku

import (
	"io/ioutil"
	"log"
	"strings"
)

//TODO: Support non-squared DIMS (logic in Block() would need updating)
const BLOCK_DIM = 3
const DIM = BLOCK_DIM * BLOCK_DIM
const ROW_SEP = "\n"
const COL_SEP = "|"
const ALT_COL_SEP = "||"

type Grid struct {
	initalized      bool
	cells           [DIM * DIM]Cell
	rows            [DIM]CellList
	cols            [DIM]CellList
	blocks          [DIM]CellList
	queue           *FiniteQueue
	numFilledCells  int
	invalidCells    map[*Cell]bool
	cachedSolutions []*Grid
}

var gridCache chan *Grid

const MAX_GRIDS = 100

const NUM_SOLVER_THREADS = 4

func init() {
	gridCache = make(chan *Grid, MAX_GRIDS)
}

func GetGrid() *Grid {
	select {
	case grid := <-gridCache:
		grid.ResetOverrides()
		return grid
	default:
		return NewGrid()
	}
	return nil
}

func DropGrids() {
	for {
		select {
		case <-gridCache:
			//Keep on going
		default:
			return
		}
	}
}

func ReturnGrid(grid *Grid) {
	select {
	case gridCache <- grid:
		//Returned it to the queue.
	default:
		//Drop it on the floor.
	}
}

func NewGrid() *Grid {
	result := &Grid{}
	result.queue = NewFiniteQueue(1, DIM)
	result.invalidCells = make(map[*Cell]bool)
	i := 0
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			result.cells[i] = NewCell(result, r, c)
			//The cell can't insert itself because it doesn't know where it will actually live in memory yet.
			result.queue.Insert(&result.cells[i])
			i++
		}
	}
	result.initalized = true
	return result
}

func (self *Grid) Load(data string) {
	//All col separators are basically just to make it easier to read. Remove them.
	data = strings.Replace(data, ALT_COL_SEP, COL_SEP, -1)
	data = strings.Replace(data, COL_SEP, "", -1)
	//TODO: shouldn't we have more error checking, like for wrong dimensions?
	for r, row := range strings.Split(data, ROW_SEP) {
		for c, data := range strings.Split(row, "") {
			cell := self.Cell(r, c)
			cell.Load(data)
		}
	}
}

func (self *Grid) LoadFromFile(path string) bool {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	self.Load(string(data))
	return true
}

//Returns a new grid that has exactly the same numbers placed as the original.
func (self *Grid) Copy() *Grid {
	//TODO: ideally we'd have some kind of smart SparseGrid or something that we can return.
	result := GetGrid()
	result.Load(self.DataString())
	return result
}

func (self *Grid) transpose() *Grid {
	//Returns a new grid that is the same as this grid (ignoring overrides, which are nulled), but with rows and cols swapped.
	result := GetGrid()
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			original := self.Cell(r, c)
			copy := result.Cell(c, r)
			copy.SetNumber(original.Number())
		}
	}
	return result
}

func (self *Grid) ResetOverrides() {
	for _, cell := range self.cells {
		cell.resetExcludes()
	}
}

func (self *Grid) Row(index int) CellList {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Row: ", index)
		return nil
	}
	if self.rows[index] == nil {
		self.rows[index] = self.cellList(index, 0, index, DIM-1)
	}
	return self.rows[index]
}

func (self *Grid) Col(index int) CellList {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Col: ", index)
		return nil
	}
	if self.cols[index] == nil {
		self.cols[index] = self.cellList(0, index, DIM-1, index)
	}
	return self.cols[index]
}

func (self *Grid) Block(index int) CellList {
	if index < 0 || index >= DIM {
		log.Println("Invalid index passed to Block: ", index)
		return nil
	}
	if self.blocks[index] == nil {
		topRow, topCol, bottomRow, bottomCol := self.blockExtents(index)
		self.blocks[index] = self.cellList(topRow, topCol, bottomRow, bottomCol)
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

func (self *Grid) Cell(row int, col int) *Cell {
	index := row*DIM + col
	if index >= DIM*DIM || index < 0 {
		log.Println("Invalid row/col index passed to Cell: ", row, ", ", col)
		return nil
	}
	return &self.cells[index]
}

func (self *Grid) cellList(rowOne int, colOne int, rowTwo int, colTwo int) CellList {
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
	return CellList(result)
}

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

//Grid will never be invalid based on moves made by the solver; it will detect times that
//someone called SetNumber with an impossible number after the fact, though.
func (self *Grid) Invalid() bool {
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
	self.cachedSolutions = nil
	if cell.Number() == 0 {
		self.numFilledCells--
	} else {
		self.numFilledCells++
	}
}

func (self *Grid) Rank() int {
	return len(self.cells) - self.numFilledCells
}

func (self *Grid) DataString() string {
	var rows []string
	for r := 0; r < DIM; r++ {
		var row []string
		for c := 0; c < DIM; c++ {
			row = append(row, self.cells[r*DIM+c].DataString())
		}
		rows = append(rows, strings.Join(row, COL_SEP))
	}
	return strings.Join(rows, ROW_SEP)
}

func (self *Grid) String() string {
	var rows []string
	for r := 0; r < DIM; r++ {
		var row []string
		for c := 0; c < DIM; c++ {
			row = append(row, self.cells[r*DIM+c].String())
		}
		rows = append(rows, strings.Join(row, COL_SEP))
	}
	return strings.Join(rows, ROW_SEP)
}

func (self *Grid) Diagram() string {
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
		tempRows = self.Cell(r, 0).diagramRows()
		for c := 1; c < DIM; c++ {
			cellRows := self.Cell(r, c).diagramRows()
			for i, _ := range tempRows {
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
