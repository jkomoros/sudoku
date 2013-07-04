package sudoku

type instructionType int

const (
	INSERT = iota
)

type instruction struct {
	result          chan interface{}
	instructionType instructionType
	item            interface{}
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
			wrappedItem := &stackItem{instruction.item, self.firstItem}
			self.firstItem = wrappedItem
			self.numItems++
			instruction.result <- nil
		}
		//Drop other instructions on the floor for now.
	}
}

func (self *SyncedStack) Insert(item interface{}) {
	result := make(chan interface{})
	self.instructions <- instruction{result, INSERT, item}
	<-result
	return
}
