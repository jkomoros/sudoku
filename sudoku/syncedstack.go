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

type ChanSyncedStack struct {
	*SyncedStack
	DefaultProbability float32
	Output             chan interface{}
	doneChan           chan bool
}

type SyncedStack struct {
	closed         bool
	instructions   chan instruction
	numItems       int
	numActiveItems int
	firstItem      *stackItem
}

func NewChanSyncedStack(doneChan chan bool) *ChanSyncedStack {
	//WARNING: doneChan should be a buffered channel or else you cloud get deadlock!
	result := &ChanSyncedStack{&SyncedStack{false, make(chan instruction), 0, 0, nil}, 1.0, make(chan interface{}, 1), doneChan}
	go result.workLoop()
	return result
}

func NewSyncedStack() *SyncedStack {
	stack := &SyncedStack{false, make(chan instruction), 0, 0, nil}
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
	self.instructions <- instruction{nil, DISPOSE, nil, 0.0}
	//Purposefully don't block.
}

func (self *SyncedStack) workLoop() {
	for {
		instruction, ok := <-self.instructions
		if !ok {
			return
		}
		self.processInstruction(instruction)
	}
}

func (self *ChanSyncedStack) workLoop() {

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

func (self *ChanSyncedStack) processInstruction(instruction instruction) {
	if instruction.instructionType == DISPOSE {
		self.doDispose()
	} else {
		self.SyncedStack.processInstruction(instruction)
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
	case DISPOSE:
		//disposes don't have a channel result.
		self.doDispose()
	}
	//Drop other instructions on the floor for now.
}

func (self *ChanSyncedStack) ItemDone() {
	self.SyncedStack.ItemDone()
	if self.IsDone() {
		self.Dispose()
	}
}

func (self *SyncedStack) ItemDone() {
	if self.closed {
		return
	}
	result := make(chan interface{})
	self.instructions <- instruction{result, DECREMENT_ACTIVE, nil, 0.0}
	<-result
	return
}

func (self *SyncedStack) Insert(item interface{}) {
	if self.closed {
		return
	}
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
	if self.closed {
		return nil
	}
	//Working from the back, will take each item with probability probability, else move to the next item in the stack.
	result := make(chan interface{})
	self.instructions <- instruction{result, GET, nil, probability}
	return <-result
}

func (self *ChanSyncedStack) doDispose() {
	self.SyncedStack.doDispose()
	close(self.Output)
	self.doneChan <- true
}

func (self *SyncedStack) doDispose() {
	//TODO: do we need to close out anything else here?
	//TODO: the work loops, when they notice the channel is closed, should probably exit.
	self.closed = true
	close(self.instructions)
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
