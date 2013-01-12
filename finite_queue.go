package dokugen

import (
	"math/rand"
)

type RankedObject interface {
	Rank() int
}

type FiniteQueue struct {
	min     int
	max     int
	objects [][]RankedObject
}

type SyncedFiniteQueue struct {
	queue FiniteQueue
	In    chan RankedObject
	Out   chan RankedObject
	Exit  chan bool
}

//Returns a new queue that will work for items with a rank as low as min or as high as max (inclusive)
func NewFiniteQueue(min int, max int) *FiniteQueue {
	return &FiniteQueue{min, max, make([][]RankedObject, max-min+1)}
}

func NewSyncedFiniteQueue(min int, max int) *SyncedFiniteQueue {
	result := &SyncedFiniteQueue{*NewFiniteQueue(min, max), make(chan RankedObject), make(chan RankedObject), make(chan bool, 1)}
	go result.workLoop()
	return result
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
	//Scrub the list for this item.
	for _, item := range list {
		//Structs will compare equal if all of their fields are the same.
		if item == obj {
			//It's already there, just return.
			return
		}
	}
	//If we get to here we need to add the item.
	self.setList(rank, append(list, obj))
}

func (self *FiniteQueue) Get() RankedObject {
	return self.GetSmallerThan(self.max + 1)
}

func (self *FiniteQueue) GetSmallerThan(max int) RankedObject {

	for i, list := range self.objects {
		if i+self.min >= max {
			return nil
		}
	ListLoop:
		for len(list) > 0 {

			//Clear off nils from the front.
			for list[0] == nil {
				if len(list) == 1 {
					self.objects[i] = nil
					list = nil
					//There are no more in this list; move to the next.
					continue ListLoop
				} else {
					self.objects[i] = list[1:]
					list = list[1:]
				}
			}
			//Pick one at random
			index := rand.Intn(len(list))
			obj := list[index]
			//Mark its old location as emptied.
			list[index] = nil

			//Meh, try again from this list.
			if obj == nil {
				continue
			}

			//Does this object still have the rank it did when it was inserted?
			if obj.Rank() == i+self.min {
				return obj
			}
			//otherwise, keep looking.
		}
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
