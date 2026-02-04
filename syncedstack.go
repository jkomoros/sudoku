package sudoku

import (
	"math/rand"
)

type instructionType int

const (
	_INSERT instructionType = iota
	_GET
	_DECREMENT_ACTIVE
	_DISPOSE
)

type instruction[T any] struct {
	result          chan T
	instructionType instructionType
	item            T
	probability     float32
}

type stackItem[T any] struct {
	item T
	next *stackItem[T]
}

type chanSyncedStack[T any] struct {
	*syncedStack[T]
	DefaultProbability float32
	Output             chan T
	doneChan           chan bool
}

type syncedStack[T any] struct {
	closed         bool
	instructions   chan instruction[T]
	numItems       int
	numActiveItems int
	firstItem      *stackItem[T]
}

func newChanSyncedStack[T any](doneChan chan bool) *chanSyncedStack[T] {
	//WARNING: doneChan should be a buffered channel or else you cloud get deadlock!
	result := &chanSyncedStack[T]{&syncedStack[T]{false, make(chan instruction[T]), 0, 0, nil}, 1.0, make(chan T, 1), doneChan}
	go result.workLoop()
	return result
}

func newSyncedStack[T any]() *syncedStack[T] {
	stack := &syncedStack[T]{false, make(chan instruction[T]), 0, 0, nil}
	go stack.workLoop()
	return stack
}

func (self *syncedStack[T]) Length() int {
	return self.numItems
}

func (self *syncedStack[T]) NumActiveItems() int {
	return self.numActiveItems
}

func (self *syncedStack[T]) IsDone() bool {
	return self.numItems == 0 && self.numActiveItems == 0
}

func (self *syncedStack[T]) Dispose() {
	if self.closed {
		return
	}
	var zero T
	self.instructions <- instruction[T]{nil, _DISPOSE, zero, 0.0}
	//Purposefully don't block.
}

func (self *syncedStack[T]) workLoop() {
	for {
		instruction, ok := <-self.instructions
		if !ok {
			return
		}
		self.processInstruction(instruction)
	}
}

func (self *chanSyncedStack[T]) workLoop() {

	//This workloop is complicated.
	//If we ahve an item and there's room in the Output channel, we ALWAYS want to do that.
	// But if there's not room in the channel, we don't want to just wait for an instruction--
	// because it's possible that the Output is emptied before we get another instruction.
	// So try first to fill output, then try either.

	var instruction instruction[T]
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

func (self *chanSyncedStack[T]) processInstruction(instruction instruction[T]) {
	if instruction.instructionType == _DISPOSE {
		self.doDispose()
	} else {
		self.syncedStack.processInstruction(instruction)
	}
}

func (self *syncedStack[T]) processInstruction(instruction instruction[T]) {
	var zero T
	switch instruction.instructionType {
	case _INSERT:
		self.doInsert(instruction.item)
		instruction.result <- zero
	case _GET:
		wrappedItem, previous := self.doSelect(instruction.probability)
		instruction.result <- self.doExtract(wrappedItem, previous)
	case _DECREMENT_ACTIVE:
		self.doDecrementActive()
		instruction.result <- zero
	case _DISPOSE:
		//disposes don't have a channel result.
		self.doDispose()
	}
	//Drop other instructions on the floor for now.
}

func (self *chanSyncedStack[T]) ItemDone() {
	self.syncedStack.ItemDone()
	if self.IsDone() {
		self.Dispose()
	}
}

func (self *syncedStack[T]) ItemDone() {
	if self.closed {
		return
	}
	var zero T
	result := make(chan T)
	self.instructions <- instruction[T]{result, _DECREMENT_ACTIVE, zero, 0.0}
	<-result
	return
}

func (self *syncedStack[T]) Insert(item T) {
	if self.closed {
		return
	}
	result := make(chan T)
	self.instructions <- instruction[T]{result, _INSERT, item, 0.0}
	<-result
	return
}

func (self *chanSyncedStack[T]) Pop() T {
	//You must use output for a ChanSyncedStack
	var zero T
	return zero
}

func (self *syncedStack[T]) Pop() T {
	//Gets the last item on the stack.
	return self.Get(1.0)
}

func (self *chanSyncedStack[T]) Get(probability float32) T {
	//You must use output for a ChanSyncedStack
	var zero T
	return zero
}

func (self *syncedStack[T]) Get(probability float32) T {
	if self.closed {
		var zero T
		return zero
	}
	//Working from the back, will take each item with probability probability, else move to the next item in the stack.
	var zero T
	result := make(chan T)
	self.instructions <- instruction[T]{result, _GET, zero, probability}
	return <-result
}

func (self *chanSyncedStack[T]) doDispose() {
	self.syncedStack.doDispose()
	close(self.Output)
	self.doneChan <- true
}

func (self *syncedStack[T]) doDispose() {
	//TODO: do we need to close out anything else here?
	//TODO: the work loops, when they notice the channel is closed, should probably exit.
	self.closed = true
	//TODO: ... Wait, why can we not close this? If we do close this, we get "send on closed channel"
	//close(self.instructions)
}

func (self *syncedStack[T]) doDecrementActive() {
	//May only be called from workLoop.
	if self.numActiveItems > 0 {
		self.numActiveItems--
	}
}

func (self *syncedStack[T]) doInsert(item T) {
	//May only be called from workLoop
	wrappedItem := &stackItem[T]{item, self.firstItem}
	self.firstItem = wrappedItem
	self.numItems++
}

func (self *syncedStack[T]) doSelect(probability float32) (item *stackItem[T], previous *stackItem[T]) {

	var lastItem *stackItem[T]
	var lastLastItem *stackItem[T]

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

func (self *syncedStack[T]) doExtract(item *stackItem[T], previous *stackItem[T]) T {
	//may only be called from within workLoop
	//Called when we've decided we ware going to take the item.
	var zero T
	if item == nil {
		return zero
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
