package sudokustate

type digest struct {
	Puzzle string
	Moves  []digestMove
}

type digestMove struct {
	Type   string
	Marks  map[int]bool
	Time   int
	Number int
	Group  digestGroup
}

type digestGroup struct {
	Type string
	ID   int
}
