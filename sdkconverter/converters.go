//Package sdkconverter provides a set of converters to and from sudoku's default sdk format.
package sdkconverter

import (
	"github.com/jkomoros/sudoku"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

/*
 * Converters convert to and from non-default file formats for sudoku puzzles.
 */

const _KOMO_CELL_RE = `\d!?:?(\[(\d\.)*\d\])?`

type SudokuPuzzleConverter interface {
	//Load loads the puzzle defined by `puzzle`, in the format tied to this
	//partcular converter, into the provided grid. Returns false if the puzzle
	//couldn't be loaded (generally because it's invalid)
	Load(grid *sudoku.Grid, puzzle string) bool
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

	grid.LoadSDK(sdk)

	return converter.DataString(grid)
}

//Format returns the most likely format type for the provided puzzle string,
//or "" if none are valid.
func Format(puzzle string) string {
	for format, converter := range Converters {
		if converter.Valid(puzzle) {
			return format
		}
	}
	return ""
}

//Load returns a new grid that is loaded up from the provided puzzle string.
//The format is guessed from the input string. If no valid format is guessed,
//an empty (non-nil) grid is returned.
func Load(puzzle string) *sudoku.Grid {
	grid := sudoku.NewGrid()
	LoadInto(grid, puzzle)
	return grid
}

//LoadInto loads the given puzzle state into the given grid. The format of the
//puzzle string is guessed. If no valid format can be detected, the grid won't
//be modified.
func LoadInto(grid *sudoku.Grid, puzzle string) {
	formatString := Format(puzzle)
	converter := Converters[formatString]
	if converter == nil {
		return
	}
	converter.Load(grid, puzzle)
}

type cellInfo struct {
	number int
	locked bool
	marks  sudoku.IntSlice
}

func (i cellInfo) fillCell(cell *sudoku.Cell) {
	cell.SetNumber(i.number)
	if i.locked {
		cell.Lock()
	} else {
		cell.Unlock()
	}
	for _, num := range i.marks {
		cell.SetMark(num, true)
	}
}

func parseKomoCell(data string) cellInfo {
	//How many characters into the string we are. We can't use the index from
	//range, because that's byte offset, which we don't care about.
	i := 0

	result := cellInfo{}

	solutionNumber := 0

	//TODO: having this hand-rolled parser feels really brittle
	inMarkSection := false

	for _, ch := range data {

		if i == 0 {
			//The first char must be the solutionNumber.
			solutionNumber, _ = strconv.Atoi(string(ch))
		} else if inMarkSection {
			switch ch {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				num, _ := strconv.Atoi(string(ch))
				result.marks = append(result.marks, num)
			case '.':
				//Mark delim. OK
			case ']':
				inMarkSection = false
			}
		} else {
			switch ch {
			case '!':
				result.locked = true
				result.number = solutionNumber
			//TODO: better way of checking for it being an int, tied to DIM
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				//User filled number since not in postion 0
				result.number, _ = strconv.Atoi(string(ch))
			case '[':
				inMarkSection = true
			default:
				//Unknown
				panic(ch)
			}
		}
		i++
	}
	return result
}

func (c *komoConverter) Load(grid *sudoku.Grid, puzzle string) bool {
	//TODO: also handle odd things like user-provided marks and other things.

	if !c.Valid(puzzle) {
		return false
	}

	//TODO: reset grid?

	rows := strings.Split(puzzle, ";")
	for r, row := range rows {
		cols := strings.Split(row, ",")
		for c, col := range cols {
			parseKomoCell(col).fillCell(grid.Cell(r, c))
		}
	}

	return true
}

func (c *komoConverter) DataString(grid *sudoku.Grid) string {

	//TODO: understand marks, user-filled numbers.
	//TODO: actually lock locked cells.

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

	rows := strings.Split(puzzle, ";")

	if len(rows) != sudoku.DIM {
		//Too few rows!
		return false
	}

	cellRE := regexp.MustCompile(_KOMO_CELL_RE)

	for _, row := range rows {
		cols := strings.Split(row, ",")
		if len(cols) != sudoku.DIM {
			//Too few cols!
			return false
		}
		for _, col := range cols {
			if !cellRE.MatchString(col) {
				return false
			}
			//TODO: have this validate spaces that have:
			//* branchNumbers
			//* branchMarks
			//* customBox number
		}
	}
	return true
}

func (c *sdkConverter) Load(grid *sudoku.Grid, puzzle string) bool {
	if !c.Valid(puzzle) {
		return false
	}
	grid.LoadSDK(puzzle)
	return true
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
