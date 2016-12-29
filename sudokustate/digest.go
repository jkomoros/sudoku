package sudokustate

import (
	"encoding/json"
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
	Number *int
	Group  groupInfo
}

//TODO: implement Model.Digest()[]byte

func (m *Model) Digest() []byte {
	obj := m.makeDigest()

	result, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return nil
	}
	return result
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

		groupInfoPtr := command.GroupInfo()

		var info groupInfo

		if groupInfoPtr == nil {
			info = groupInfo{}
		} else {
			info = *groupInfoPtr
		}

		for _, subCommand := range command.SubCommands() {
			result = append(result, digestMove{
				Type: subCommand.Type(),
				//TODO: this is a hack, we just happen to know that there's only one item
				Cell:   subCommand.ModifiedCells(m)[0],
				Group:  info,
				Marks:  subCommand.Marks(),
				Number: subCommand.Number(),
			})
		}

		currentCommand = currentCommand.next
	}

	return result
}
