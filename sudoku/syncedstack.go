package sudoku

import (
	"math/rand"
)

type instructionType int

const (
	INSERT = iota
	GET
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

type SyncedStack struct {
	instructions chan instruction
	numItems     int
	firstItem    *stackItem
}

func NewSyncedStack() *SyncedStack {
	stack := &SyncedStack{make(chan instruction), 0, nil}
	go stack.workLoop()
	return stack
}

func (self *SyncedStack) Length() int {
	return self.numItems
}

func (self *SyncedStack) Dispose() {
	close(self.instructions)
}

func (self *SyncedStack) workLoop() {
	for {
		instruction := <-self.instructions
		switch instruction.instructionType {
		case INSERT:
			self.doInsert(instruction.item)
			instruction.result <- nil
		case GET:
			instruction.result <- self.doGet(instruction.probability)
		}
		//Drop other instructions on the floor for now.
	}
}

func (self *SyncedStack) Insert(item interface{}) {
	result := make(chan interface{})
	self.instructions <- instruction{result, INSERT, item, 0.0}
	<-result
	return
}

func (self *SyncedStack) Pop() interface{} {
	//Gets the last item on the stack.
	return self.Get(1.0)
}

func (self *SyncedStack) Get(probability float32) interface{} {
	//Working from the back, will take each item with probability probability, else move to the next item in the stack.
	result := make(chan interface{})
	self.instructions <- instruction{result, GET, nil, probability}
	return <-result
}

func (self *SyncedStack) doInsert(item interface{}) {
	//May only be called from workLoop
	wrappedItem := &stackItem{item, self.firstItem}
	self.firstItem = wrappedItem
	self.numItems++
}

func (self *SyncedStack) doGet(probability float32) interface{} {
	//May only be called from workLoop
	wrappedItem := self.firstItem
	var lastItem *stackItem
	var lastLastItem *stackItem
	for wrappedItem != nil {
		if rand.Float32() < probability {
			//Found it!
			self.numItems--
			//Mend it
			if lastItem == nil {
				//It must have been the first item.
				self.firstItem = wrappedItem.next
			} else {
				lastItem.next = wrappedItem.next
			}
			return wrappedItem.item
		}
		lastLastItem = lastItem
		lastItem = wrappedItem
		wrappedItem = wrappedItem.next
	}
	//if we got to here, just return the lastItem.
	if lastItem == nil {
		return nil
	}
	self.numItems--
	if lastLastItem == nil {
		self.firstItem = nil
	} else {
		lastLastItem.next = nil
	}
	return lastItem.item
}
