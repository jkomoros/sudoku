package dokugen

type CellList interface {
	All() CellStream
	Without(cell *Cell) CellStream
}

type CellStream chan *Cell

type simpleCellList struct {
	grid   *Grid
	start  int
	end    int
	stride int
	//TODO: mark if cache is valid or not, and build it up the first time we walk through it. Right now we don't actually need channels.
	cache []*Cell
}

func (self CellStream) Now() (result []*Cell) {
	if self == nil {
		return
	}
	for cell := range self.Chan() {
		result = append(result, cell)
	}
	self = nil
	return
}

func (self CellStream) Chan() chan *Cell {
	if self == nil {
		return nil
	}
	result := chan *Cell(self)
	self = nil
	return result
}

func (self *simpleCellList) All() CellStream {
	return self.Without(nil)
}

func (self *simpleCellList) Without(exclude *Cell) CellStream {
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
		close(result)
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
