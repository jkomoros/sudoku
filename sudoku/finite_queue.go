package sudoku

import (
	"math/rand"
)

type RankedObject interface {
	Rank() int
}

type FiniteQueue struct {
	min     int
	max     int
	objects []*finiteQueueBucket
}

type finiteQueueBucket struct {
	objects []RankedObject
	numNils int
	rank    int
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
	result := FiniteQueue{min, max, make([]*finiteQueueBucket, max-min+1)}
	for i := 0; i < max-min+1; i++ {
		result.objects[i] = &finiteQueueBucket{make([]RankedObject, 1), 0, i + result.min}
	}
	return &result
}

func NewSyncedFiniteQueue(min int, max int) *SyncedFiniteQueue {
	result := &SyncedFiniteQueue{*NewFiniteQueue(min, max), make(chan RankedObject), make(chan RankedObject), make(chan bool, 1)}
	go result.workLoop()
	return result
}

func (self *finiteQueueBucket) removeNils() {
	result := make([]RankedObject, len(self.objects)-self.numNils)
	targetIndex := 0
	for i := 0; i < len(self.objects); i++ {
		if self.objects[i] == nil {
			continue
		}
		result[targetIndex] = self.objects[i]
		targetIndex++
	}
	self.objects = result
	self.numNils = 0
}

func (self *finiteQueueBucket) trimNils() {
	if float64(self.numNils)/float64(len(self.objects)) > _REALLOCATE_PROPORTION {
		self.removeNils()
	}
	for len(self.objects) > 0 && self.objects[0] == nil {
		self.objects = self.objects[1:]
		self.numNils--
	}

}

func (self *finiteQueueBucket) setNil(index int) {
	if self.objects[index] == nil {
		//We don't want to double-count nils.
		return
	}
	self.objects[index] = nil
	self.numNils++
}

func (self *finiteQueueBucket) getItem() RankedObject {
	self.trimNils()
	for len(self.objects) > 0 {
		index := rand.Intn(len(self.objects))
		obj := self.objects[index]
		//Mark this object as gathered.
		self.setNil(index)
		if obj != nil && obj.Rank() == self.rank {
			return obj
		}
		self.trimNils()
	}
	return nil
}

func (self *finiteQueueBucket) addItem(item RankedObject) {
	//Scrub the list for this item.
	for _, obj := range self.objects {
		//Structs will compare equal if all of their fields are the same.
		if item == obj {
			//It's already there, just return.
			return
		}
	}
	self.objects = append(self.objects, item)
}

func (self *finiteQueueBucket) compact() {

	//First, find any i's that don't have the right rank anymore and mark as nils.
	for i, obj := range self.objects {
		if obj == nil {
			continue
		}
		if obj.Rank() != self.rank {
			self.setNil(i)
		}
	}
	self.removeNils()
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
}

func (self *FiniteQueue) Get() RankedObject {
	return self.GetSmallerThan(self.max + 1)
}

func (self *FiniteQueue) GetSmallerThan(max int) RankedObject {
	return self.getSmallerThan(max, make(map[RankedObject]int))
}

func (self *FiniteQueue) getSmallerThan(max int, ignoredObjects map[RankedObject]int) RankedObject {
	//TOOD: actually respect ignoredObjects
	for i, list := range self.objects {
		if i+self.min >= max {
			return nil
		}
		result := list.getItem()
		if result != nil {
			return result
		}
	}
	return nil
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
