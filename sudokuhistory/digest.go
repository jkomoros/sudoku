package sudokuhistory

import (
	"errors"
	"github.com/jkomoros/sudoku"
	"github.com/jkomoros/sudoku/sdkconverter"
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

//valid will return nil if it is valid, an error otherwise.
func (d *Digest) valid() error {

	//Check that the puzzle exists.
	if d.Puzzle == "" {
		return errors.New("No puzzle")
	}

	//Check if the puzzle is a valid puzzle in doku format.
	converter := sdkconverter.Converters[sdkconverter.DokuFormat]

	if !converter.Valid(d.Puzzle) {
		return errors.New("Puzzle snapshot not doku format")
	}

	var lastTime time.Duration

	for _, group := range d.MoveGroups {
		if group.TimeOffset < 0 {
			return errors.New("Invalid time offset")
		}
		if group.TimeOffset < lastTime {
			return errors.New("TimeOffsets did not monotonically increase")
		}
		lastTime = group.TimeOffset

		if len(group.Moves) == 0 {
			return errors.New("One of the groups had no moves")
		}

		for _, move := range group.Moves {
			if move.Cell.Row < 0 || move.Cell.Row >= sudoku.DIM {
				return errors.New("Invalid row in cellref")
			}
			if move.Cell.Col < 0 || move.Cell.Col >= sudoku.DIM {
				return errors.New("Invalid col in cellref")
			}
			if move.Marks == nil && move.Number == nil {
				return errors.New("Neither marks nor number provided in move")
			}
			if move.Marks != nil && move.Number != nil {
				return errors.New("Both marks and number provided")
			}
		}
	}

	//Everything's OK!
	return nil
}

//LoadDigest takes in a Digest produced by model.Digest() and sets the
//internal state appropriately.
func (m *Model) LoadDigest(d Digest) error {

	if err := d.valid(); err != nil {
		return err
	}

	m.SetGrid(sdkconverter.Load(d.Puzzle))

	for _, group := range d.MoveGroups {
		m.StartGroup(group.Description)

		for _, move := range group.Moves {
			if move.Marks != nil {
				m.SetMarks(move.Cell, move.Marks)
			} else {
				m.SetNumber(move.Cell, *move.Number)
			}
		}

		m.FinishGroupAndExecute()

		m.currentCommand.c.time = group.TimeOffset
	}

	return nil

}

//Digest returns a Digest object representing the state of this model.
//Suitable for serializing as json.
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
