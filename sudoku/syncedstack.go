package sudoku

type instructionType int

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
