package dokugen

const DIM = 9

type Cell struct {
	data string
}

type Grid struct {
	data  string
	cells [DIM * DIM]Cell
}
