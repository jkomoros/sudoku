package sudoku

import (
	"math/rand"
)

type RankedObject interface {
	Rank() int
}

type FiniteQueue struct {
	min           int
	max           int
	objects       []*finiteQueueBucket
	currentBucket *finiteQueueBucket
}

type finiteQueueBucket struct {
	objects  []RankedObject
	rank     int
	shuffled bool
}

const _REALLOCATE_PROPORTION = 0.20

type SyncedFiniteQueue struct {
	queue FiniteQueue
	In    chan RankedObject
	Out   chan RankedObject
	Exit  chan bool
}

type FiniteQueueGetter struct {
	queue         *FiniteQueue
	ignoreObjects map[RankedObject]bool
	currentList   *finiteQueueBucket
}

//Returns a new queue that will work for items with a rank as low as min or as high as max (inclusive)
func NewFiniteQueue(min int, max int) *FiniteQueue {
	result := FiniteQueue{min, max, make([]*finiteQueueBucket, max-min+1), nil}
	for i := 0; i < max-min+1; i++ {
		result.objects[i] = &finiteQueueBucket{make([]RankedObject, 0), i + result.min, true}
	}
	return &result
}

func NewSyncedFiniteQueue(min int, max int) *SyncedFiniteQueue {
	result := &SyncedFiniteQueue{*NewFiniteQueue(min, max), make(chan RankedObject), make(chan RankedObject), make(chan bool, 1)}
	go result.workLoop()
	return result
}

func (self *finiteQueueBucket) getItem() RankedObject {

	if !self.shuffled {
		self.shuffle()
	}

	for len(self.objects) > 0 {
		item := self.objects[0]
		self.objects = self.objects[1:]
		if item.Rank() == self.rank {
			return item
		}
	}
	return nil
}

func (self *finiteQueueBucket) empty() bool {
	return len(self.objects) == 0
}

func (self *finiteQueueBucket) addItem(item RankedObject) {
	if item == nil {
		return
	}
	//Scrub the list for this item.
	for _, obj := range self.objects {
		//Structs will compare equal if all of their fields are the same.
		if item == obj {
			//It's already there, just return.
			return
		}
	}
	self.objects = append(self.objects, item)
	self.shuffled = false
}

func (self *finiteQueueBucket) copy() *finiteQueueBucket {
	//TODO: test this
	newObjects := make([]RankedObject, len(self.objects))
	copy(newObjects, self.objects)
	//We set shuffled to false because right now we copy the shuffle order of our copy.
	return &finiteQueueBucket{newObjects, self.rank, false}
}

func (self *finiteQueueBucket) shuffle() {
	//TODO: test this.
	//Shuffles the items in place.
	newPositions := rand.Perm(len(self.objects))
	newObjects := make([]RankedObject, len(self.objects))
	for i, j := range newPositions {
		newObjects[j] = self.objects[i]
	}
	self.objects = newObjects
	self.shuffled = true
}

func (self *SyncedFiniteQueue) workLoop() {
	for {
		firstItem := self.queue.Get()
		itemSent := false
		if firstItem == nil {
			//We can take in new things or accept an exit.
			select {
			case <-self.Exit:
				return
			case incoming := <-self.In:
				self.queue.Insert(incoming)
			}
		} else {
			//We can take in new things, send out smallest one, or exit.
			select {
			case <-self.Exit:
				return
			case incoming := <-self.In:
				self.queue.Insert(incoming)
			case self.Out <- firstItem:
				itemSent = true
			}
			//If we didn't send the item out, we need to put it back in.
			if !itemSent {
				self.queue.Insert(firstItem)
			}
		}
	}
}

func (self *FiniteQueue) NewGetter() *FiniteQueueGetter {
	list, _ := self.getBucket(self.min)
	return &FiniteQueueGetter{self, make(map[RankedObject]bool), list}
}

func (self *FiniteQueue) Min() int {
	return self.min
}

func (self *FiniteQueue) Max() int {
	return self.max
}

func (self *FiniteQueue) Insert(obj RankedObject) {
	rank := obj.Rank()
	list, ok := self.getBucket(rank)
	if !ok {
		//Apparently rank wasn't legal.
		return
	}
	list.addItem(obj)
	self.currentBucket = nil
}

func (self *FiniteQueue) Get() RankedObject {
	return self.GetSmallerThan(self.max + 1)
}

func (self *FiniteQueue) GetSmallerThan(max int) RankedObject {
	return self.getSmallerThan(max, make(map[RankedObject]int))
}

func (self *FiniteQueue) getSmallerThan(max int, ignoredObjects map[RankedObject]int) RankedObject {

	//TODO: remove ignoredObjects as an argument.

	//TODO: test that if an item is inserted while we're walking through we return it.

	if self.currentBucket == nil {
		self.currentBucket, _ = self.getBucket(self.min)
		if self.currentBucket == nil {
			return nil
		}
	}

	item := self.currentBucket.getItem()

	for item == nil {
		if self.currentBucket.empty() {
			self.currentBucket, _ = self.getBucket(self.currentBucket.rank + 1)
			if self.currentBucket == nil || self.currentBucket.rank >= max {
				//Got to the end
				return nil
			}
		}
		item = self.currentBucket.getItem()
	}

	return item

}

func (self *FiniteQueue) legalRank(rank int) bool {
	return rank >= self.min && rank <= self.max
}

func (self *FiniteQueue) getBucket(rank int) (*finiteQueueBucket, bool) {
	if !self.legalRank(rank) {
		return nil, false
	}
	return self.objects[rank-self.min], true
}
