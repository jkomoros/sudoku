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
	objects []*finiteQueueList
}

type finiteQueueList struct {
	objects []RankedObject
	numNils int
	rank    int
}

type SyncedFiniteQueue struct {
	queue FiniteQueue
	In    chan RankedObject
	Out   chan RankedObject
	Exit  chan bool
}

//Returns a new queue that will work for items with a rank as low as min or as high as max (inclusive)
func NewFiniteQueue(min int, max int) *FiniteQueue {
	result := FiniteQueue{min, max, make([]*finiteQueueList, max-min+1)}
	for i := 0; i < max-min+1; i++ {
		result.objects[i] = &finiteQueueList{make([]RankedObject, 1), 0, i + result.min}
	}
	return &result
}

func NewSyncedFiniteQueue(min int, max int) *SyncedFiniteQueue {
	result := &SyncedFiniteQueue{*NewFiniteQueue(min, max), make(chan RankedObject), make(chan RankedObject), make(chan bool, 1)}
	go result.workLoop()
	return result
}

func (self *finiteQueueList) trimNils() {
	//TODO: if the numNils is greater than some proportion of len, do a full reallocate.
	for len(self.objects) > 0 && self.objects[0] == nil {
		self.objects = self.objects[1:]
		self.numNils--
	}
}

func (self *finiteQueueList) setNil(index int) {
	self.objects[index] = nil
	self.numNils++
}

func (self *finiteQueueList) getItem() RankedObject {
	self.trimNils()
	for len(self.objects) > 0 {
		index := rand.Intn(len(self.objects))
		obj := self.objects[index]
		if obj == nil {
			continue
		}
		if obj.Rank() != self.rank {
			self.setNil(index)
		}
		self.trimNils()
	}
	return nil
}

func (self *finiteQueueList) addItem(item RankedObject) {
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

func (self *FiniteQueue) getList(rank int) (*finiteQueueList, bool) {
	if !self.legalRank(rank) {
		return nil, false
	}
	return self.objects[rank-self.min], true
}
