package sudokustate

import (
	"github.com/jkomoros/sudoku"
)

//Digest is an object representing the state of the model. Suitable for being
//saved as json.
type Digest struct {
	Puzzle string
	Moves  []MoveDigest
}

//MoveDigest is the record of a single move captured within a Digest.
type MoveDigest struct {
	Type   string
	Cell   sudoku.CellRef
	Marks  map[int]bool `json:",omitempty"`
	Time   int
	Number *int       `json:",omitempty"`
	Group  *groupInfo `json:",omitempty"`
}

//TODO: implement model.LoadDigest([]byte)

//Digest returns a Digest object representing the state of this model.
func (m *Model) Digest() Digest {
	return Digest{
		Puzzle: m.snapshot,
		Moves:  m.makeMovesDigest(),
	}
}

func (m *Model) makeMovesDigest() []MoveDigest {
	var result []MoveDigest

	//Move command cursor to the very first item in the linked list.
	currentCommand := m.commands
	if currentCommand == nil {
		return nil
	}
	for currentCommand.prev != nil {
		currentCommand = currentCommand.prev
	}

	for currentCommand != nil {

		command := currentCommand.c

		for _, subCommand := range command.SubCommands() {
			result = append(result, MoveDigest{
				Type: subCommand.Type(),
				//TODO: this is a hack, we just happen to know that there's only one item
				Cell:   subCommand.ModifiedCells(m)[0],
				Group:  command.GroupInfo(),
				Marks:  subCommand.Marks(),
				Number: subCommand.Number(),
			})
		}

		currentCommand = currentCommand.next
	}

	return result
}
