package sudoku

import (
	"strconv"
	"strings"
)

/*
 * Converters convert to and from non-default file formats for sudoku puzzles.
 */

type SudokuPuzzleConverter interface {
	Load(grid *Grid, puzzle string)
	DataString(grid *Grid) string
}

var Converters map[string]SudokuPuzzleConverter

func init() {
	Converters = make(map[string]SudokuPuzzleConverter)
	Converters["komo"] = &komoConverter{}
}

type komoConverter struct {
}

func (c *komoConverter) Load(grid *Grid, puzzle string) {
	//TODO: also handle odd things like user-provided marks and other things.

	var result string

	rows := strings.Split(puzzle, ";")
	for _, row := range rows {
		cols := strings.Split(row, ",")
		for _, col := range cols {
			if strings.Contains(col, "!") {
				result += strings.TrimSuffix(col, "!")
			} else {
				result += "."
			}
		}
		result += "\n"
	}

	//We added an extra \n in the last runthrough, remove it.
	result = strings.TrimSuffix(result, "\n")

	grid.Load(result)
}

func (c *komoConverter) DataString(grid *Grid) string {
	//The komo puzzle format fills all cells and marks which ones are 'locked',
	//whereas the default sdk format simply leaves non-'locked' cells as blank.
	//So we need to solve the puzzle.
	solvedGrid := grid.Copy()

	if !solvedGrid.Solve() {
		//Hmm, puzzle wasn't valid. We can't represent it in this format.
		return ""
	}
	result := ""
	for r := 0; r < DIM; r++ {
		for c := 0; c < DIM; c++ {
			cell := solvedGrid.Cell(r, c)
			result += strconv.Itoa(cell.Number())
			if grid.Cell(r, c).Number() != 0 {
				result += "!"
			}
			if c != DIM-1 {
				result += ","
			}
		}
		if r != DIM-1 {
			result += ";"
		}
	}
	return result
}
