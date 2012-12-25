package dokugen

const DIM = 9

type Cell struct {
	grid   *Grid
	Number int
	Row    int
	Col    int
	simpleCellList
}

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

type Grid struct {
	data  string
	cells *[DIM * DIM]Cell
}

func (self *simpleCellList) All() chan *Cell {
	result := make(chan *Cell, DIM*DIM)
	if self.cache == nil {
		self.buildCache()
	}
	go func() {
		for _, cell := range self.cache {
			result <- cell
		}
	}()
	return result
}

func (self *simpleCellList) buildCache() {
	i := self.start
	if self.stride == 0 {
		self.stride = 1
	}
	for i < self.end {
		self.cache = append(self.cache, &self.grid.cells[i])
		i = i + self.stride
	}
}
