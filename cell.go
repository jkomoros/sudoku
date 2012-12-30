package dokugen

import (
	"log"
	"strconv"
	"strings"
)

const ALT_0 = "."
const DIAGRAM_IMPOSSIBLE = " "
const DIAGRAM_RIGHT = "|"
const DIAGRAM_BOTTOM = "-"
const DIAGRAM_CORNER = "+"
const DIAGRAM_NUMBER = "•"

type Cell struct {
	grid *Grid
	//The number if it's explicitly set. Number() will return it if it's explicitly or implicitly set.
	number      int
	Row         int
	Col         int
	Block       int
	neighbors   []*Cell
	impossibles [DIM]int
}

func NewCell(grid *Grid, row int, col int) Cell {
	//TODO: we should not set the number until neighbors are initialized.
	return Cell{grid: grid, Row: row, Col: col, Block: grid.blockForCell(row, col)}
}

func (self *Cell) Load(data string) {
	//Format, for now, is just the number itself, or 0 if no number.
	data = strings.Replace(data, ALT_0, "0", -1)
	num, _ := strconv.Atoi(data)
	self.SetNumber(num)
}

func (self *Cell) Number() int {
	//A layer of indirection since number needs to be used from the Setter.
	return self.number
}

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
		//Our rank will have changed.
		self.grid.queue.Insert(self)
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
		//Our rank will have changed.
		self.grid.queue.Insert(self)
	}
}

func (self *Cell) Possible(number int) bool {
	//Number is 1 indexed, but we store it as 0-indexed
	number--
	if number < 0 || number >= DIM {
		return false
	}
	return self.impossibles[number] == 0
}

func (self *Cell) Invalid() bool {
	//Returns true if no numbers are possible.
	//TODO: figure out a way to send this back up to the solver when it happens.
	for _, counter := range self.impossibles {
		if counter == 0 {
			return false
		}
	}
	return true
}

func (self *Cell) Rank() int {
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

func (self *Cell) Neighbors() []*Cell {
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
				if self.impossibles[current] == 0 {
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
