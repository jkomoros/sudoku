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

func (self *FiniteQueue) Min() int {
	return self.min
}

func (self *FiniteQueue) Max() int {
	return self.max
}

func (self *FiniteQueue) Insert(obj RankedObject) {
	rank := obj.Rank()
	list, ok := self.getList(rank)
	if !ok {
		//Apparently rank wasn't legal.
		return
	}
	//Scrub the list for this item.
	for _, item := range list {
		//Structs will compare equal if all of their fields are the same.
		if &item == &obj {
			//It's already there, just return.
			return
		}
	}
	//If we get to here we need to add the item.
	self.setList(rank, append(list, obj))
}

func (self *FiniteQueue) Get() RankedObject {
	for i, list := range self.objects {
		if len(list) == 0 {
			continue
		}
		obj := list[0]
		if len(list) == 1 {
			self.objects[i] = nil
		} else {
			self.objects[i] = list[1:]
		}
		//Does this object still have the rank it did when it was inserted?
		if obj.Rank() == i+self.min {
			return obj
		}
		//otherwise, keep looking.
	}
	return nil
}

func (self *FiniteQueue) legalRank(rank int) bool {
	return rank >= self.min && rank <= self.max
}

func (self *FiniteQueue) getList(rank int) ([]RankedObject, bool) {
	if !self.legalRank(rank) {
		return nil, false
	}
	return self.objects[rank-self.min], true
}

func (self *FiniteQueue) setList(rank int, list []RankedObject) {
	if !self.legalRank(rank) {
		return
	}
	self.objects[rank-self.min] = list
}
