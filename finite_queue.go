package dokugen

type RankedObject interface {
	Rank() int
}

type FiniteQueue struct {
	min     int
	max     int
	objects [][]RankedObject
}

//Returns a new queue that will work for items with a rank as low as min or as high as max (inclusive)
func NewFiniteQueue(min int, max int) *FiniteQueue {
	return &FiniteQueue{min, max, make([][]RankedObject, max-min+1)}
}
