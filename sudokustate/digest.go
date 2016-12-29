package sudokustate

import (
	"github.com/jkomoros/sudoku"
	"time"
)

//Digest is an object representing the state of the model. Consists primarily
//of a list of MoveGroupDigests. Suitable for being saved as json.
type Digest struct {
	//Puzzle is the puzzle, encoded as a string in doku format. (See
	//sdkconverter package)
	Puzzle     string
	MoveGroups []MoveGroupDigest
}

//MoveGroupDigest is the record of a group of moves that should all be applied
//at once. Most MoveGroups have a single move, but some have multiple.
type MoveGroupDigest struct {
	Moves []MoveDigest
	//How many nanoseconds elapsed since when the puzzle was reset to this
	//move.
	TimeOffset time.Duration `json:",omitempty"`
	//The description of the group (if it has one). Simple groups that contain
	//a single marks or number move often do not have group names.
	Description string `json:",omitempty"`
}

//MoveDigest is the record of a single move captured within a Digest, either
//setting a number or setting a set of marks on a cell (but never both).
type MoveDigest struct {
	Cell sudoku.CellRef
	//Which marks should be modified. Numbers that have a true should be set,
	//and numbers that have a false should be unset. All other marks will be
	//left the same. Either one or the other of Marks or Number may be set,
	//but never both.
	Marks map[int]bool `json:",omitempty"`
	//Which number to set. Either one or the other of Marks or Number may be
	//set, but never both.
	Number *int `json:",omitempty"`
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
			TimeOffset:  command.time,
		})

		currentCommand = currentCommand.next
	}

	return result
}
