package sudoku

import (
	"math/rand"
)

type instructionType int

const (
	INSERT = iota
	GET
	DECREMENT_ACTIVE
	DISPOSE
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

type chanSyncedStack struct {
	*syncedStack
	DefaultProbability float32
	Output             chan interface{}
	doneChan           chan bool
}

type syncedStack struct {
	closed         bool
	instructions   chan instruction
	numItems       int
	numActiveItems int
	firstItem      *stackItem
}

func newChanSyncedStack(doneChan chan bool) *chanSyncedStack {
	//WARNING: doneChan should be a buffered channel or else you cloud get deadlock!
	result := &chanSyncedStack{&syncedStack{false, make(chan instruction), 0, 0, nil}, 1.0, make(chan interface{}, 1), doneChan}
	go result.workLoop()
	return result
}

func newSyncedStack() *syncedStack {
	stack := &syncedStack{false, make(chan instruction), 0, 0, nil}
	go stack.workLoop()
	return stack
}

func (self *syncedStack) Length() int {
	return self.numItems
}

func (self *syncedStack) NumActiveItems() int {
	return self.numActiveItems
}

func (self *syncedStack) IsDone() bool {
	return self.numItems == 0 && self.numActiveItems == 0
}

func (self *syncedStack) Dispose() {
	if self.closed {
		return
	}
	self.instructions <- instruction{nil, DISPOSE, nil, 0.0}
	//Purposefully don't block.
}

func (self *syncedStack) workLoop() {
	for {
		instruction, ok := <-self.instructions
		if !ok {
			return
		}
		self.processInstruction(instruction)
	}
}

func (self *chanSyncedStack) workLoop() {

	//This workloop is complicated.
	//If we ahve an item and there's room in the Output channel, we ALWAYS want to do that.
	// But if there's not room in the channel, we don't want to just wait for an instruction--
	// because it's possible that the Output is emptied before we get another instruction.
	// So try first to fill output, then try either.

	var instruction instruction
	var ok bool
	for {
		if self.closed {
			return
		}
		if self.numItems > 0 {
			wrappedItem, previous := self.doSelect(self.DefaultProbability)
			select {
			case self.Output <- wrappedItem.item:
				self.doExtract(wrappedItem, previous)
			default:
				select {
				case self.Output <- wrappedItem.item:
					self.doExtract(wrappedItem, previous)
				case instruction, ok = <-self.instructions:
					if !ok {
						return
					}
					self.processInstruction(instruction)
				}
			}
		} else {
			instruction, ok = <-self.instructions
			if !ok {
				return
			}
			self.processInstruction(instruction)
		}
	}
}

func (self *chanSyncedStack) processInstruction(instruction instruction) {
	if instruction.instructionType == DISPOSE {
		self.doDispose()
	} else {
		self.syncedStack.processInstruction(instruction)
	}
}

func (self *syncedStack) processInstruction(instruction instruction) {
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
	case DISPOSE:
		//disposes don't have a channel result.
		self.doDispose()
	}
	//Drop other instructions on the floor for now.
}

func (self *chanSyncedStack) ItemDone() {
	self.syncedStack.ItemDone()
	if self.IsDone() {
		self.Dispose()
	}
}

func (self *syncedStack) ItemDone() {
	if self.closed {
		return
	}
	result := make(chan interface{})
	self.instructions <- instruction{result, DECREMENT_ACTIVE, nil, 0.0}
	<-result
	return
}

func (self *syncedStack) Insert(item interface{}) {
	if self.closed {
		return
	}
	result := make(chan interface{})
	self.instructions <- instruction{result, INSERT, item, 0.0}
	<-result
	return
}

func (self *chanSyncedStack) Pop() interface{} {
	//You must use output for a ChanSyncedStack
	return nil
}

func (self *syncedStack) Pop() interface{} {
	//Gets the last item on the stack.
	return self.Get(1.0)
}

func (self *chanSyncedStack) Get(probability float32) interface{} {
	//You must use output for a ChanSyncedStack
	return nil
}

func (self *syncedStack) Get(probability float32) interface{} {
	if self.closed {
		return nil
	}
	//Working from the back, will take each item with probability probability, else move to the next item in the stack.
	result := make(chan interface{})
	self.instructions <- instruction{result, GET, nil, probability}
	return <-result
}

func (self *chanSyncedStack) doDispose() {
	self.syncedStack.doDispose()
	close(self.Output)
	self.doneChan <- true
}

func (self *syncedStack) doDispose() {
	//TODO: do we need to close out anything else here?
	//TODO: the work loops, when they notice the channel is closed, should probably exit.
	self.closed = true
	//TODO: ... Wait, why can we not close this? If we do close this, we get "send on closed channel"
	//close(self.instructions)
}

func (self *syncedStack) doDecrementActive() {
	//May only be called from workLoop.
	if self.numActiveItems > 0 {
		self.numActiveItems--
	}
}

func (self *syncedStack) doInsert(item interface{}) {
	//May only be called from workLoop
	wrappedItem := &stackItem{item, self.firstItem}
	self.firstItem = wrappedItem
	self.numItems++
}

func (self *syncedStack) doSelect(probability float32) (item *stackItem, previous *stackItem) {

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

func (self *syncedStack) doExtract(item *stackItem, previous *stackItem) interface{} {
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
