//Package sdkconverter provides a set of converters to and from sudoku's default sdk format.
package sdkconverter

import (
	"github.com/jkomoros/sudoku"
	"strconv"
	"strings"
	"unicode/utf8"
)

/*
 * Converters convert to and from non-default file formats for sudoku puzzles.
 */

type SudokuPuzzleConverter interface {
	//Load loads the puzzle defined by `puzzle`, in the format tied to this partcular converter,
	//into the provided grid.
	Load(grid *sudoku.Grid, puzzle string)
	//DataString returns the serialization of the provided grid in the format provided by this converter.
	DataString(grid *sudoku.Grid) string
	//Valid returns true if the provided string will successfully deserialize
	//with the given converter to a grid. For example, if the string contains
	//data only 50 cells, this should return false.
	Valid(puzzle string) bool
}

type komoConverter struct {
}

//This one is a total pass through, just for convenience.
type sdkConverter struct {
}

//Converters is a list of the provided converters. Currently only "komo" and "sdk" (a pass-through) are provided.
var Converters map[string]SudokuPuzzleConverter

func init() {
	Converters = make(map[string]SudokuPuzzleConverter)
	Converters["komo"] = &komoConverter{}
	Converters["sdk"] = &sdkConverter{}
}

//ToSDK is a convenience wrapper that takes the name of a format and the puzzle data
//and returns an sdk string.
func ToSDK(format string, other string) (sdk string) {
	grid := sudoku.NewGrid()
	converter := Converters[format]

	if converter == nil {
		return ""
	}

	converter.Load(grid, other)

	return grid.DataString()

}

//ToOther is a conenience wrapper that takes the name of a format and the sdk datastring
//and returns the DataString in the other format.
func ToOther(format string, sdk string) (other string) {
	grid := sudoku.NewGrid()
	converter := Converters[format]

	if converter == nil {
		return ""
	}

	grid.Load(sdk)

	return converter.DataString(grid)
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

	//Solve the grid.
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

func (c *komoConverter) Valid(puzzle string) bool {
	//TODO: implement this for real and test
	return true
}

func (c *sdkConverter) Load(grid *sudoku.Grid, puzzle string) {
	grid.Load(puzzle)
}

func (c *sdkConverter) DataString(grid *sudoku.Grid) string {
	return grid.DataString()
}

func (c *sdkConverter) Valid(puzzle string) bool {

	//The SDK format is essentially just DIM * DIM characters that are either
	//0-DIM or a '.'. Col separators and newlines are stripped.

	//Strip out all newlines and col seps.
	puzzle = strings.Replace(puzzle, sudoku.COL_SEP, "", -1)
	puzzle = strings.Replace(puzzle, sudoku.ALT_COL_SEP, "", -1)
	puzzle = strings.Replace(puzzle, "\n", "", -1)

	if len(puzzle) != sudoku.DIM*sudoku.DIM {
		//Wrong length!
		return false
	}

	//Make sure every character is either an int 0-DIM or a '.'

	//make map of legal chars

	legalChars := make(map[rune]bool)

	legalChars['.'] = true
	for i := 0; i <= sudoku.DIM; i++ {
		ch, _ := utf8.DecodeRuneInString(strconv.Itoa(i))
		legalChars[ch] = true
	}

	for _, ch := range puzzle {
		if !legalChars[ch] {
			//Illegal character!
			return false
		}
	}

	return true
}
