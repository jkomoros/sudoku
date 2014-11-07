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
	//Version monotonically increases as changes are made to the underlying queue.
	//This is how getters ensure that they stay in sync with their underlying queue.
	version       int
	defaultGetter *FiniteQueueGetter
}

type finiteQueueBucket struct {
	objects  []RankedObject
	rank     int
	shuffled bool
}

const _REALLOCATE_PROPORTION = 0.20

type SyncedFiniteQueue struct {
	queue FiniteQueue
	//TODO: should these counts actually be on the basic FiniteQueue?
	items       int
	activeItems int
	In          chan RankedObject
	Out         chan RankedObject
	ItemDone    chan bool
	Exit        chan bool
	//We'll send a true to this every time ItemDone() causes us to be IsDone()
	//Note: it should be buffered!
	done chan bool
}

type FiniteQueueGetter struct {
	queue            *FiniteQueue
	dispensedObjects map[RankedObject]int
	currentBucket    *finiteQueueBucket
	baseVersion      int
}

//Returns a new queue that will work for items with a rank as low as min or as high as max (inclusive)
func NewFiniteQueue(min int, max int) *FiniteQueue {
	result := FiniteQueue{min,
		max,
		make([]*finiteQueueBucket, max-min+1),
		nil,
		0,
		nil}
	for i := 0; i < max-min+1; i++ {
		result.objects[i] = &finiteQueueBucket{make([]RankedObject, 0), i + result.min, true}
	}
	return &result
}

func NewSyncedFiniteQueue(min int, max int, done chan bool) *SyncedFiniteQueue {
	result := &SyncedFiniteQueue{*NewFiniteQueue(min, max),
		0,
		0,
		make(chan RankedObject),
		make(chan RankedObject),
		make(chan bool),
		make(chan bool, 1),
		done}
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

func (self *SyncedFiniteQueue) IsDone() bool {
	return self.activeItems == 0 && self.items == 0
}

func (self *SyncedFiniteQueue) workLoop() {

	//We use this pattern to avoid duplicating code
	when := func(condition bool, c chan RankedObject) chan RankedObject {
		if condition {
			return c
		}
		return nil
	}

	exiting := false

	for {
		firstItem := self.queue.Get()
		itemSent := false

		//We can take in new things, send out smallest one, or exit.
		select {
		case <-self.Exit:
			exiting = true
			close(self.Out)
			if self.activeItems == 0 {
				//TODO: should we drain all of the incoming ones?
				return
			}
		case incoming := <-self.In:
			self.queue.Insert(incoming)
			self.items++
		case <-self.ItemDone:
			self.activeItems--
			if self.IsDone() {
				self.done <- true

			}
			if self.activeItems == 0 && exiting {
				return
			}
		case when(firstItem != nil && !exiting, self.Out) <- firstItem:
			itemSent = true
			self.items--
			self.activeItems++
		}
		//If we didn't send the item out, we need to put it back in.
		if firstItem != nil && !itemSent {
			self.queue.Insert(firstItem)
		}
	}

}

func (self *FiniteQueue) NewGetter() *FiniteQueueGetter {
	list, _ := self.getBucket(self.min)
	return &FiniteQueueGetter{self, make(map[RankedObject]int), list, 0}
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
	self.version++
}

func (self *FiniteQueue) Get() RankedObject {
	return self.GetSmallerThan(self.max + 1)
}

func (self *FiniteQueue) GetSmallerThan(max int) RankedObject {
	return self.getSmallerThan(max)
}

func (self *FiniteQueue) getSmallerThan(max int) RankedObject {

	//Very similar logic exists in finiteQueueBucket.getSmallerThan.
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

	self.version++
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

func (self *FiniteQueue) DefaultGetter() *FiniteQueueGetter {
	if self.defaultGetter == nil {
		self.defaultGetter = self.NewGetter()
	}
	return self.defaultGetter
}

func (self *FiniteQueue) ResetDefaultGetter() {
	self.defaultGetter = nil
}

func (self *FiniteQueueGetter) Get() RankedObject {
	if self.queue == nil {
		return nil
	}
	return self.getSmallerThan(self.queue.max + 1)
}

func (self *FiniteQueueGetter) GetSmallerThan(max int) RankedObject {
	return self.getSmallerThan(max)
}

func (self *FiniteQueueGetter) getSmallerThan(max int) RankedObject {
	//Very similar logic exists in finiteQueue.getSmallerThan.

	//Sanity check
	if self.queue == nil {
		return nil
	}

	if self.baseVersion != self.queue.version {
		self.currentBucket = nil
		self.baseVersion = self.queue.version
	}

	if self.currentBucket == nil {
		newBucket, _ := self.queue.getBucket(self.queue.min)
		if newBucket == nil {
			return nil
		}
		self.currentBucket = newBucket.copy()
	}

	var item RankedObject

	for {
		item = self.currentBucket.getItem()
		for item == nil {
			if self.currentBucket.empty() {
				newBucket, _ := self.queue.getBucket(self.currentBucket.rank + 1)
				if newBucket == nil || newBucket.rank >= max {
					//Got to the end
					return nil
				}
				self.currentBucket = newBucket.copy()
			}
			item = self.currentBucket.getItem()
		}
		//We have AN item. Is it one we've already dispensed?
		if rank, dispensed := self.dispensedObjects[item]; !dispensed || rank != item.Rank() {
			//It is new. Break out of the loop.
			break
		}
		//Otherwise, loop around again.
	}

	//Keep track of the fact we dispensed this item and keep track of its rank when it was dispensed.
	self.dispensedObjects[item] = item.Rank()

	return item
}
