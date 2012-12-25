package dokugen

type CellList interface {
	All() chan *Cell
	Without(cell *Cell) chan *Cell
}

type simpleCellList struct {
	grid   *Grid
	start  int
	end    int
	stride int
	cache  []*Cell
}

func (self *simpleCellList) All() chan *Cell {
	return self.Without(nil)
}

func (self *simpleCellList) Without(exclude *Cell) chan *Cell {
	result := make(chan *Cell, DIM*DIM)
	if self.cache == nil {
		self.buildCache()
	}
	go func() {
		for _, cell := range self.cache {
			if cell == exclude {
				continue
			}
			result <- cell
		}
	}()
	return result
}

func (self *simpleCellList) buildCache() {
	if self.grid == nil {
		panic("Grid is nil!")
	}
	i := self.start
	if self.stride == 0 {
		self.stride = 1
	}
	for i < self.end {
		self.cache = append(self.cache, &self.grid.cells[i])
		i = i + self.stride
	}
}
