package sudokustate

import (
	"github.com/jkomoros/sudoku"
	"time"
)

//Digest is an object representing the state of the model. Consists primarily
//of a list of MoveGroupDigests. Suitable for being saved as json.
type Digest struct {
	Puzzle     string
	MoveGroups []MoveGroupDigest
}

//MoveGroupDigest is the record of a group of moves that should all be applied
//at once. Most MoveGroups have a single move, but some have multiple.
type MoveGroupDigest struct {
	Moves       []MoveDigest
	Time        time.Duration `json:",omitempty"`
	Description string        `json:",omitempty"`
}

//MoveDigest is the record of a single move captured within a Digest, either
//setting a number or setting a set of marks on a cell (but never both). The
//Marks field will have a true for every number that should be set, and a
//false for every mark that should be unset. All other marks will be left the
//same.
type MoveDigest struct {
	Cell   sudoku.CellRef
	Marks  map[int]bool `json:",omitempty"`
	Number *int         `json:",omitempty"`
}

//TODO: implement model.LoadDigest([]byte)

//Digest returns a Digest object representing the state of this model.
func (m *Model) Digest() Digest {
	return Digest{
		Puzzle:     m.snapshot,
		MoveGroups: m.makeMoveGroupsDigest(),
	}
}

func (m *Model) makeMoveGroupsDigest() []MoveGroupDigest {
	var result []MoveGroupDigest

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

		var moves []MoveDigest

		for _, subCommand := range command.subCommands {
			moves = append(moves, MoveDigest{
				//TODO: this is a hack, we just happen to know that there's only one item
				Cell:   subCommand.ModifiedCells(m)[0],
				Marks:  subCommand.Marks(),
				Number: subCommand.Number(),
			})
		}

		result = append(result, MoveGroupDigest{
			Moves:       moves,
			Description: command.description,
			Time:        command.time,
		})

		currentCommand = currentCommand.next
	}

	return result
}
