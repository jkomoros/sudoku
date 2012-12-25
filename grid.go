package dokugen

const DIM = 9

type Cell struct {
	Number int
	Row    int
	Col    int
}

type CellList interface {
	All() chan *Cell
	Without(cell *Cell) chan *Cell
}

type Grid struct {
	data  string
	cells *[DIM * DIM]Cell
}
