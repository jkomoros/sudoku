package sudoku

import (
	"math/rand"
	"sort"
	"sync"
)

type rankedObject interface {
	rank() int
}

type queue interface {
	NewGetter() queueGetter
}

type queueGetter interface {
	Get() rankedObject
	GetSmallerThan(max int) rankedObject
}

//readOnlyCellQueue is a special queue that fits queue and queue getter
//interfaces and is optimized for use as grid.queue when you know that grid
//isn't mutable. It's also designed to be easy to bootstrap by copying the
//value of a previous queue, then calling fix().
type readOnlyCellQueue struct {
	grid     Grid
	cellRefs [DIM * DIM]cellRef
}

type readOnlyCellQueueGetter struct {
	queue   *readOnlyCellQueue
	counter int
}

type finiteQueue struct {
	min           int
	max           int
	objects       []*finiteQueueBucket
	currentBucket *finiteQueueBucket
	//Version monotonically increases as changes are made to the underlying queue.
	//This is how getters ensure that they stay in sync with their underlying queue.
	versionLock sync.RWMutex
	version     int
}

type finiteQueueBucket struct {
	objects  []rankedObject
	rank     int
	shuffled bool
}

const _REALLOCATE_PROPORTION = 0.20

type syncedFiniteQueue struct {
	queue finiteQueue
	lock  *sync.RWMutex
	//TODO: should these counts actually be on the basic FiniteQueue?
	items       int
	activeItems int
	In          chan rankedObject
	Out         chan rankedObject
	ItemDone    chan bool
	Exit        chan bool
	//We'll send a true to this every time ItemDone() causes us to be IsDone()
	//Note: it should be buffered!
	done chan bool
}

type finiteQueueGetter struct {
	queue            *finiteQueue
	dispensedObjects map[rankedObject]int
	currentBucket    *finiteQueueBucket
	baseVersion      int
}

//Returns a new queue that will work for items with a rank as low as min or as high as max (inclusive)
func newFiniteQueue(min int, max int) *finiteQueue {
	result := finiteQueue{min,
		max,
		make([]*finiteQueueBucket, max-min+1),
		nil,
		sync.RWMutex{},
		0}
	for i := 0; i < max-min+1; i++ {
		result.objects[i] = &finiteQueueBucket{make([]rankedObject, 0), i + result.min, true}
	}
	return &result
}

func newSyncedFiniteQueue(min int, max int, done chan bool) *syncedFiniteQueue {
	result := &syncedFiniteQueue{*newFiniteQueue(min, max),
		&sync.RWMutex{},
		0,
		0,
		make(chan rankedObject),
		make(chan rankedObject),
		make(chan bool),
		make(chan bool, 1),
		done}
	go result.workLoop()
	return result
}

//defaultRefs should be called to initalize the object to have default refs.
//No need to call this if you're copying in an earlier state.
func (r *readOnlyCellQueue) defaultRefs() {

	counter := 0
	for row := 0; row < DIM; row++ {
		for col := 0; col < DIM; col++ {
			r.cellRefs[counter] = cellRef{row, col}
			counter++
		}
	}
}

//Fix should be called after all of the items are in place and before any
//Getters have been vended.
func (r *readOnlyCellQueue) fix() {
	sort.Sort(r)
}

func (r *readOnlyCellQueue) Len() int {
	return DIM * DIM
}

func (r *readOnlyCellQueue) Less(i, j int) bool {
	firstCellRank := r.cellRefs[i].Cell(r.grid).rank()
	secondCellRank := r.cellRefs[j].Cell(r.grid).rank()
	return firstCellRank < secondCellRank
}

func (r *readOnlyCellQueue) Swap(i, j int) {
	r.cellRefs[i], r.cellRefs[j] = r.cellRefs[j], r.cellRefs[i]
}

func (r *readOnlyCellQueue) NewGetter() queueGetter {
	return &readOnlyCellQueueGetter{queue: r}
}

func (r *readOnlyCellQueueGetter) Get() rankedObject {
	//This will never return an item with a rank less than 0, which is the
	//behavior of normal finiteQueues in Grids because of how they're
	//configured. It feels kind of weird that we bake that constraint in here;
	//although I guess it's appropriate given that these
	//readOnlyCellQueueGetters are so specialized anyway.
	for {
		if r.counter >= r.queue.Len() {
			return nil
		}
		result := r.queue.cellRefs[r.counter].Cell(r.queue.grid)
		r.counter++
		if result.rank() > 0 {
			return result
		}
	}
}

func (r *readOnlyCellQueueGetter) GetSmallerThan(max int) rankedObject {
	item := r.Get()
	if item == nil {
		return nil
	}
	if item.rank() >= max {
		//Put it back!
		r.counter--
		return nil
	}
	return item
}

func (self *finiteQueueBucket) getItem() rankedObject {

	if !self.shuffled {
		self.shuffle()
	}

	for len(self.objects) > 0 {
		item := self.objects[0]
		self.objects = self.objects[1:]
		if item.rank() == self.rank {
			return item
		}
	}
	return nil
}

func (self *finiteQueueBucket) empty() bool {
	return len(self.objects) == 0
}

func (self *finiteQueueBucket) addItem(item rankedObject) {
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
	newObjects := make([]rankedObject, len(self.objects))
	copy(newObjects, self.objects)
	//We set shuffled to false because right now we copy the shuffle order of our copy.
	return &finiteQueueBucket{newObjects, self.rank, false}
}

func (self *finiteQueueBucket) shuffle() {
	//TODO: test this.
	//Shuffles the items in place.
	newPositions := rand.Perm(len(self.objects))
	newObjects := make([]rankedObject, len(self.objects))
	for i, j := range newPositions {
		newObjects[j] = self.objects[i]
	}
	self.objects = newObjects
	self.shuffled = true
}

func (self *syncedFiniteQueue) IsDone() bool {

	self.lock.RLock()
	result := self.activeItems == 0 && self.items == 0
	self.lock.RUnlock()

	return result
}

func (self *syncedFiniteQueue) workLoop() {

	//We use this pattern to avoid duplicating code
	when := func(condition bool, c chan rankedObject) chan rankedObject {
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
			self.lock.Lock()
			self.items++
			self.lock.Unlock()
		case <-self.ItemDone:
			self.lock.Lock()
			self.activeItems--
			self.lock.Unlock()
			if self.IsDone() {
				self.done <- true

			}
			if self.activeItems == 0 && exiting {
				return
			}
		case when(firstItem != nil && !exiting, self.Out) <- firstItem:
			itemSent = true
			self.lock.Lock()
			self.items--
			self.activeItems++
			self.lock.Unlock()
		}
		//If we didn't send the item out, we need to put it back in.
		if firstItem != nil && !itemSent {
			self.queue.Insert(firstItem)
		}
	}

}

func (self *finiteQueue) NewGetter() queueGetter {
	list, _ := self.getBucket(self.min)
	return &finiteQueueGetter{self, make(map[rankedObject]int), list, 0}
}

func (self *finiteQueue) Min() int {
	return self.min
}

func (self *finiteQueue) Max() int {
	return self.max
}

func (self *finiteQueue) Insert(obj rankedObject) {
	rank := obj.rank()
	list, ok := self.getBucket(rank)
	if !ok {
		//Apparently rank wasn't legal.
		return
	}
	list.addItem(obj)
	self.currentBucket = nil
	self.versionLock.Lock()
	self.version++
	self.versionLock.Unlock()
}

func (self *finiteQueue) Get() rankedObject {
	return self.GetSmallerThan(self.max + 1)
}

func (self *finiteQueue) GetSmallerThan(max int) rankedObject {
	return self.getSmallerThan(max)
}

func (self *finiteQueue) getSmallerThan(max int) rankedObject {

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

	self.versionLock.Lock()
	self.version++
	self.versionLock.Unlock()
	return item

}

func (self *finiteQueue) legalRank(rank int) bool {
	return rank >= self.min && rank <= self.max
}

func (self *finiteQueue) getBucket(rank int) (*finiteQueueBucket, bool) {
	if !self.legalRank(rank) {
		return nil, false
	}
	return self.objects[rank-self.min], true
}

func (self *finiteQueueGetter) Get() rankedObject {
	if self.queue == nil {
		return nil
	}
	return self.getSmallerThan(self.queue.max + 1)
}

func (self *finiteQueueGetter) GetSmallerThan(max int) rankedObject {
	return self.getSmallerThan(max)
}

func (self *finiteQueueGetter) getSmallerThan(max int) rankedObject {
	//Very similar logic exists in finiteQueue.getSmallerThan.

	//Sanity check
	if self.queue == nil {
		return nil
	}

	self.queue.versionLock.RLock()
	queueVersion := self.queue.version

	if self.baseVersion != queueVersion {
		self.currentBucket = nil
		self.baseVersion = queueVersion
	}
	self.queue.versionLock.RUnlock()

	if self.currentBucket == nil {
		newBucket, _ := self.queue.getBucket(self.queue.min)
		if newBucket == nil {
			return nil
		}
		self.currentBucket = newBucket.copy()
	}

	var item rankedObject

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
		if rank, dispensed := self.dispensedObjects[item]; !dispensed || rank != item.rank() {
			//It is new. Break out of the loop.
			break
		}
		//Otherwise, loop around again.
	}

	//Keep track of the fact we dispensed this item and keep track of its rank when it was dispensed.
	self.dispensedObjects[item] = item.rank()

	return item
}
