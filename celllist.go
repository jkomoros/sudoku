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
	//TODO: cache the stream the first time through. Don't store in the cache value until it's done, and don't stomp if someone beat us.
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
	go func() {
		i := self.start
		if self.stride == 0 {
			self.stride = 1
		}
		for i < self.end {
			result <- &self.grid.cells[i]
			i = i + self.stride
		}
		close(result)
	}()
	return result
}
