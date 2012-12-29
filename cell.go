package dokugen

import (
	"log"
	"strconv"
)

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
	num, _ := strconv.Atoi(data)
	self.SetNumber(num)
}

func (self *Cell) Number() int {
	//A layer of indirection since number could be set explicitly or implicitly.
	return self.number
	//TODO: return the number if there's only one that's possible
}

func (self *Cell) SetNumber(number int) {
	//Sets the explicit number. This will affect its neighbors possibles list (in the future).
	oldNumber := self.number
	self.number = number
	if oldNumber > 0 {
		for i := 1; i <= DIM; i++ {
			if i == oldNumber {
				continue
			}
			self.setPossible(i)
		}
	}
	if number > 0 {
		for i := 1; i <= DIM; i++ {
			if i == number {
				continue
			}
			self.setImpossible(i)
		}
	}
	//TODO: alert neighbors that it changed.
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
	//TODO: see if this allows us to have an un-set implicitly set number, and alert neighbors if so.

}

func (self *Cell) setImpossible(number int) {
	//Number is 1 indexed, but we store it as 0-indexed
	number--
	if number < 0 || number >= DIM {
		return
	}
	self.impossibles[number]++
	//TODO: see if this allows us to have an implicitly set number, and alert neighbors if so.
}

func (self *Cell) Possible(number int) bool {
	//Number is 1 indexed, but we store it as 0-indexed
	number--
	if number < 0 || number >= DIM {
		return false
	}
	return self.impossibles[number] == 0
}

func (self *Cell) Neighbors() []*Cell {
	if !self.grid.initalized {
		log.Println("Neighbors called before the grid had been finalized")
		return nil
	}
	if self.neighbors == nil {
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
	return strconv.Itoa(self.Number())
}

func (self *Cell) String() string {
	return "Cell[" + strconv.Itoa(self.Row) + "][" + strconv.Itoa(self.Col) + "]:" + strconv.Itoa(self.Number()) + "\n"
}
