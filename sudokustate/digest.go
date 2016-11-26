package sudokustate

import (
	"github.com/jkomoros/sudoku"
)

type digest struct {
	Puzzle string
	Moves  []digestMove
}

type digestMove struct {
	Type   string
	Cell   sudoku.CellRef
	Marks  map[int]bool
	Time   int
	Number int
	Group  groupInfo
}

func (m *Model) makeDigest() digest {
	//TODO: test this
	return digest{
		Puzzle: m.snapshot,
		Moves:  m.makeMovesDigest(),
	}
}

func (m *Model) makeMovesDigest() []digestMove {
	var result []digestMove

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

		groupInfo := command.GroupInfo()

		for _, subCommand := range command.SubCommands() {
			result = append(result, digestMove{
				Type: subCommand.Type(),
				//TODO: this is a hack, we just happen to know that there's only one item
				Cell:  subCommand.ModifiedCells(m)[0],
				Group: *groupInfo,
				Marks: subCommand.Marks(),
			})
		}

		currentCommand = currentCommand.next
	}

	return result
}
