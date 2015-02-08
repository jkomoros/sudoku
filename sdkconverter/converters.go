package sdkconverter

import (
	"dokugen"
	"strconv"
	"strings"
)

/*
 * Converters convert to and from non-default file formats for sudoku puzzles.
 */

type SudokuPuzzleConverter interface {
	Load(grid *sudoku.Grid, puzzle string)
	DataString(grid *sudoku.Grid) string
}

var Converters map[string]SudokuPuzzleConverter

func init() {
	Converters = make(map[string]SudokuPuzzleConverter)
	Converters["komo"] = &komoConverter{}
}

func ToSDK(format string, other string) (sdk string) {
	grid := sudoku.NewGrid()
	converter := Converters[format]

	if converter == nil {
		return ""
	}

	converter.Load(grid, other)

	return grid.DataString()

}

func ToOther(format string, sdk string) (other string) {
	grid := sudoku.NewGrid()
	converter := Converters[format]

	if converter == nil {
		return ""
	}

	grid.Load(sdk)

	return converter.DataString(grid)
}

type komoConverter struct {
}

func (c *komoConverter) Load(grid *sudoku.Grid, puzzle string) {
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

func (c *komoConverter) DataString(grid *sudoku.Grid) string {
	//The komo puzzle format fills all cells and marks which ones are 'locked',
	//whereas the default sdk format simply leaves non-'locked' cells as blank.
	//So we need to solve the puzzle.
	solvedGrid := grid.Copy()

	//TODO: when running the tests, this fails every so often, even though the puzzle is solveable. This implies
	//a BIG problem somewhere in solver.
	if !solvedGrid.Solve() {
		//Hmm, puzzle wasn't valid. We can't represent it in this format.
		return ""
	}
	result := ""
	for r := 0; r < sudoku.DIM; r++ {
		for c := 0; c < sudoku.DIM; c++ {
			cell := solvedGrid.Cell(r, c)
			result += strconv.Itoa(cell.Number())
			if grid.Cell(r, c).Number() != 0 {
				result += "!"
			}
			if c != sudoku.DIM-1 {
				result += ","
			}
		}
		if r != sudoku.DIM-1 {
			result += ";"
		}
	}
	return result
}
