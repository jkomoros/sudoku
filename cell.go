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
	//A layer of indirection since number needs to be used from the Setter.
	return self.number
}

func (self *Cell) SetNumber(number int) {
	//Sets the explicit number. This will affect its neighbors possibles list.
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
	return strconv.Itoa(self.Number())
}

func (self *Cell) String() string {
	return "Cell[" + strconv.Itoa(self.Row) + "][" + strconv.Itoa(self.Col) + "]:" + strconv.Itoa(self.Number()) + "\n"
}
