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

func (self *FiniteQueue) Insert(obj RankedObject) {
	rank := obj.Rank()
	list := self.getList(rank)
	if list == nil {
		//Apparently rank wasn't legal.
		return
	}
	//Scrub the list for this item.
	for _, item := range list {
		if item == obj {
			//It's already there, just return.
			return
		}
	}
	//If we get to here we need to add the item.
	self.setList(rank, append(list, obj))
}

func (self *FiniteQueue) legalRank(rank int) bool {
	return rank >= self.min && rank <= self.max
}

func (self *FiniteQueue) getList(rank int) []RankedObject {
	if !self.legalRank(rank) {
		return nil
	}
	return self.objects[rank-self.min]
}

func (self *FiniteQueue) setList(rank int, list []RankedObject) {
	if !self.legalRank(rank) {
		return
	}
	self.objects[rank-self.min] = list
}
