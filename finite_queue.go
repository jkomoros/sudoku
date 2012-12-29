package dokugen

type RankedObject interface {
	Rank() int
}

type FiniteQueue struct {
	min     int
	objects [][]RankedObject
}

func NewFiniteQueue(min int, max int) *FiniteQueue {
	return &FiniteQueue{min, make([][]RankedObject, max-min)}
}
