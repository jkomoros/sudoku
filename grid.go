package dokugen

const DIM = 9

type Grid struct {
	data  string
	cells *[DIM * DIM]Cell
}
