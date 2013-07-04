package sudoku

import (
	"math/rand"
)

type instructionType int

const (
	INSERT = iota
	GET
	DECREMENT_ACTIVE
)

type instruction struct {
	result          chan interface{}
	instructionType instructionType
	item            interface{}
	probability     float32
}

type stackItem struct {
	item interface{}
	next *stackItem
}

type ChanSyncedStack struct {
	*SyncedStack
	DefaultProbability float32
	Output             chan interface{}
}

type SyncedStack struct {
	instructions   chan instruction
	numItems       int
	numActiveItems int
	firstItem      *stackItem
}

func NewChanSyncedStack() *ChanSyncedStack {
	result := &ChanSyncedStack{&SyncedStack{make(chan instruction), 0, 0, nil}, 1.0, make(chan interface{}, 1)}
	go result.workLoop()
	return result
}

func NewSyncedStack() *SyncedStack {
	stack := &SyncedStack{make(chan instruction), 0, 0, nil}
	go stack.workLoop()
	return stack
}

func (self *SyncedStack) Length() int {
	return self.numItems
}

func (self *SyncedStack) NumActiveItems() int {
	return self.numActiveItems
}

func (self *SyncedStack) IsDone() bool {
	return self.numItems == 0 && self.numActiveItems == 0
}

func (self *SyncedStack) Dispose() {
	close(self.instructions)
}

func (self *SyncedStack) workLoop() {
	for {
		instruction := <-self.instructions
		self.processInstruction(instruction)
	}
}

func (self *ChanSyncedStack) workLoop() {
	var instruction instruction
	for {
		wrappedItem, previous := self.doSelect(self.DefaultProbability)
		if wrappedItem != nil {
			select {
			case self.Output <- wrappedItem.item:
				self.doExtract(wrappedItem, previous)
			case instruction := <-self.instructions:
				self.processInstruction(instruction)
			}
		} else {
			instruction = <-self.instructions
			self.processInstruction(instruction)
		}
	}
}

func (self *SyncedStack) processInstruction(instruction instruction) {
	switch instruction.instructionType {
	case INSERT:
		self.doInsert(instruction.item)
		instruction.result <- nil
	case GET:
		wrappedItem, previous := self.doSelect(instruction.probability)
		instruction.result <- self.doExtract(wrappedItem, previous)
	case DECREMENT_ACTIVE:
		self.doDecrementActive()
		instruction.result <- nil
	}
	//Drop other instructions on the floor for now.
}

func (self *SyncedStack) ItemDone() {
	result := make(chan interface{})
	self.instructions <- instruction{result, DECREMENT_ACTIVE, nil, 0.0}
	<-result
	return
}

func (self *SyncedStack) Insert(item interface{}) {
	result := make(chan interface{})
	self.instructions <- instruction{result, INSERT, item, 0.0}
	<-result
	return
}

func (self *ChanSyncedStack) Pop() interface{} {
	//You must use output for a ChanSyncedStack
	return nil
}

func (self *SyncedStack) Pop() interface{} {
	//Gets the last item on the stack.
	return self.Get(1.0)
}

func (self *ChanSyncedStack) Get(probability float32) interface{} {
	//You must use output for a ChanSyncedStack
	return nil
}

func (self *SyncedStack) Get(probability float32) interface{} {
	//Working from the back, will take each item with probability probability, else move to the next item in the stack.
	result := make(chan interface{})
	self.instructions <- instruction{result, GET, nil, probability}
	return <-result
}

func (self *SyncedStack) doDecrementActive() {
	//May only be called from workLoop.
	if self.numActiveItems > 0 {
		self.numActiveItems--
	}
}

func (self *SyncedStack) doInsert(item interface{}) {
	//May only be called from workLoop
	wrappedItem := &stackItem{item, self.firstItem}
	self.firstItem = wrappedItem
	self.numItems++
}

func (self *SyncedStack) doSelect(probability float32) (item *stackItem, previous *stackItem) {

	var lastItem *stackItem
	var lastLastItem *stackItem

	wrappedItem := self.firstItem

	for wrappedItem != nil {
		if rand.Float32() < probability {
			return wrappedItem, lastItem
		}
		lastLastItem = lastItem
		lastItem = wrappedItem
		wrappedItem = wrappedItem.next
	}

	return lastItem, lastLastItem
}

func (self *SyncedStack) doExtract(item *stackItem, previous *stackItem) interface{} {
	//may only be called from within workLoop
	//Called when we've decided we ware going to take the item.
	if item == nil {
		return nil
	}
	self.numActiveItems++
	self.numItems--
	if previous != nil {
		previous.next = item.next
	} else {
		self.firstItem = item.next
	}
	return item.item
}
