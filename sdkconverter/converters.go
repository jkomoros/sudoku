/*
Package sdkconverter provides a set of converters to and from sudoku's default sdk format.

It supports three file types: 'sdk', 'doku', and 'komo'. To help ensure you're
using a supported format, pass FooFormat instead of the direct string.
*/
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

//Format is a type of puzzle format this package understands. For safety, use
//the FooFormat constants instead of strings.
type Format string

/*
SDK is the default file format of the main library, and many other sudoku
tools. Rows are delimited by new lines, and columns are delimited by a
single pipe character ('|'). Optionally the column delimiter may be replaced
with two pipes, or no characters. Each cell is represented by a single
character, 0-9 or '.'. '0' and '.' denote an unfilled cell. The sdk format
does not support marks, locks, or user-filled numbers. The SDK format is the
default format primarily because it is simple and understood by many other
tools.
*/
const SDKFormat Format = "sdk"

/*
The doku format's rows are delimited by line breaks and columns are
delimited by "|". Each cell consists of (in order):

* Either a space, period, or number 0-9, denoting the number that is
currently filled in the space.

* An optional "!" if the cell's number is locked

* An optional list of 1 or more marks, contained in "(" and ")" and
separated by ","

* 0 or more spaces or tabs (useful for lining up the file's
columns for clarity)

If there are no locked cells, marks, or extra whitespace in any cells, then
either only the column delimiters OR both the column and row delimiters may be
omitted. Thus, every valid SDK file is a valid doku file. Doku is the
recommended format for any uses that include user modifications of the cell.
*/
const DokuFormat Format = "doku"

/*
komo is a legacy format that is used in certain online sudoku games. Like doku
it can store user modifications to the grid. Unlike all of the other formats,
it  requires that each cell store the solution number for that cell, meaning
it is unable to store grids that have no valid solution.
*/
const KomoFormat Format = "komo"

const _KOMO_CELL_RE = `\d!?(:\d)?(\[(\d\.)*\d\])?`
const _DOKU_CELL_RE = `( |\.|\d)!?(\((\d,)*\d\))?( |\t)*`

type SudokuPuzzleConverter interface {
	//Load loads the puzzle defined by `puzzle`, in the format tied to this
	//partcular converter, into the provided grid. Returns false if the puzzle
	//couldn't be loaded (generally because it's invalid)
	Load(grid sudoku.MutableGrid, puzzle string) bool
	//DataString returns the serialization of the provided grid in the format provided by this converter.
	DataString(grid sudoku.Grid) string
	//Valid returns true if the provided string will successfully deserialize
	//with the given converter to a grid. For example, if the string contains
	//data only 50 cells, this should return false.
	Valid(puzzle string) bool
}

type komoConverter struct {
}

type dokuConverter struct {
}

//This one is a total pass through, just for convenience.
type sdkConverter struct {
}

//Converters is a list of the provided converters for direct access. It's
//better to use one of the convenience methods unless you need raw access.
var Converters map[Format]SudokuPuzzleConverter

func init() {
	Converters = make(map[Format]SudokuPuzzleConverter)
	Converters["komo"] = &komoConverter{}
	Converters["sdk"] = &sdkConverter{}
	Converters["doku"] = &dokuConverter{}
}

//DataString is a convenience method that returns the given grid's string
//representation in the given format. If the converter does not exist, it will
//return a zero-length string.
func DataString(format Format, grid sudoku.Grid) string {

	converter := Converters[format]

	if converter == nil {
		return ""
	}

	return converter.DataString(grid)

}

//ToSDK is a convenience wrapper that takes the name of a format and the puzzle data
//and returns an sdk string.
func ToSDK(format Format, other string) (sdk string) {
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
func ToOther(format Format, sdk string) (other string) {

	converter := Converters[format]

	if converter == nil {
		return ""
	}

	grid := sudoku.LoadSDK(sdk)

	return converter.DataString(grid)
}

//PuzzleFormat returns the most likely format type for the provided puzzle
//string, or "" if none are valid.
func PuzzleFormat(puzzle string) Format {
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
func Load(puzzle string) sudoku.MutableGrid {
	grid := sudoku.NewGrid()
	LoadInto(grid, puzzle)
	return grid
}

//LoadInto loads the given puzzle state into the given grid. The format of the
//puzzle string is guessed. If no valid format can be detected, the grid won't
//be modified.
func LoadInto(grid sudoku.MutableGrid, puzzle string) {
	formatString := PuzzleFormat(puzzle)
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

func (i cellInfo) fillCell(cell sudoku.MutableCell) {
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
	expectingUserNumber := false

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
		} else if expectingUserNumber {
			switch ch {
			//TODO: better way of checking for it being an int, tied to DIM
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				//User filled number since not in postion 0
				result.number, _ = strconv.Atoi(string(ch))
			default:
				panic(ch)
			}
			expectingUserNumber = false
		} else {
			switch ch {
			case '!':
				result.locked = true
				result.number = solutionNumber
			case ':':
				expectingUserNumber = true
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

func (c *komoConverter) Load(grid sudoku.MutableGrid, puzzle string) bool {
	//TODO: also handle odd things like user-provided marks and other things.

	if !c.Valid(puzzle) {
		return false
	}

	//TODO: reset grid?

	rows := strings.Split(puzzle, ";")
	for r, row := range rows {
		cols := strings.Split(row, ",")
		for c, col := range cols {
			parseKomoCell(col).fillCell(grid.MutableCell(r, c))
		}
	}

	return true
}

func (c *komoConverter) DataString(grid sudoku.Grid) string {

	//TODO: understand marks, user-filled numbers.
	//TODO: actually lock locked cells.

	//The komo puzzle format fills all cells and marks which ones are 'locked',
	//whereas the default sdk format simply leaves non-'locked' cells as blank.
	//So we need to solve the puzzle.
	solvedGrid := grid.MutableCopy()

	//Solve the grid.
	if !solvedGrid.Solve() {
		//Hmm, puzzle wasn't valid. We can't represent it in this format.
		return ""
	}
	result := ""
	for r := 0; r < sudoku.DIM; r++ {
		for c := 0; c < sudoku.DIM; c++ {
			solutionCell := solvedGrid.Cell(r, c)
			cell := grid.Cell(r, c)
			result += strconv.Itoa(solutionCell.Number())
			if cell.Locked() {
				result += "!"
			} else {
				//If it's not locked, and a user has input a number, output
				//that.
				if cell.Number() != 0 {
					result += ":" + strconv.Itoa(cell.Number())
				}
			}
			if len(cell.Marks()) != 0 {
				result += "["
				var stringMarkList []string
				for _, mark := range cell.Marks() {
					stringMarkList = append(stringMarkList, strconv.Itoa(mark))
				}
				result += strings.Join(stringMarkList, ".")
				result += "]"
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

func (c *dokuConverter) DataString(grid sudoku.Grid) string {
	result := ""
	for r := 0; r < sudoku.DIM; r++ {
		for c := 0; c < sudoku.DIM; c++ {
			cell := grid.Cell(r, c)
			if cell.Number() == 0 {
				//TODO: for this (any pipes) we should use sudoku's constants,
				//right?
				result += "."
			} else {
				result += strconv.Itoa(cell.Number())
			}
			if cell.Locked() {
				result += "!"
			}
			if len(cell.Marks()) != 0 {
				result += "("
				var stringMarkList []string
				for _, mark := range cell.Marks() {
					stringMarkList = append(stringMarkList, strconv.Itoa(mark))
				}
				result += strings.Join(stringMarkList, ",")
				result += ")"
			}
			if c != sudoku.DIM-1 {
				result += "|"
			}
		}
		if r != sudoku.DIM-1 {
			result += "\n"
		}
	}
	return result
}

func (c *dokuConverter) Valid(puzzle string) bool {

	if strings.HasSuffix(puzzle, "\n") {
		puzzle = puzzle[:len(puzzle)-1]
	}

	//In the simplest case, a Doku reduces down to a valid SDK.
	if Converters["sdk"].Valid(puzzle) {
		return true
	}
	//OK, it's not the simple case, let's see if it's valid.
	rows := strings.Split(puzzle, "\n")

	if len(rows) != sudoku.DIM {
		return false
	}

	cellRE := regexp.MustCompile(_DOKU_CELL_RE)

	for _, row := range rows {
		cols := strings.Split(row, "|")
		if len(cols) != sudoku.DIM {
			//Too few cols!
			return false
		}
		for _, col := range cols {
			if !cellRE.MatchString(col) {
				return false
			}
		}
	}

	return true

}

func parseDokuCell(data string) cellInfo {
	//How many characters into the string we are. We can't use the index from
	//range, because that's byte offset, which we don't care about.
	i := 0

	result := cellInfo{}

	//TODO: having this hand-rolled parser feels really brittle
	inMarkSection := false

	for _, ch := range data {

		if i == 0 {
			//The first character is the filled number.
			switch ch {
			case '0', '.', ' ':
				result.number = 0
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				result.number, _ = strconv.Atoi(string(ch))
			}
		} else if inMarkSection {
			switch ch {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				num, _ := strconv.Atoi(string(ch))
				result.marks = append(result.marks, num)
			case ',':
				//Mark delim. OK
			case ')':
				inMarkSection = false
			}
		} else {
			switch ch {
			case '!':
				result.locked = true
			case '(':
				inMarkSection = true
			case ' ', '\t':
				//Whitespace, OK
			default:
				//Unknown
				panic(ch)
			}
		}
		i++
	}
	return result
}

func (c *dokuConverter) Load(grid sudoku.MutableGrid, puzzle string) bool {

	if !c.Valid(puzzle) {
		return false
	}

	sConverter := Converters["sdk"]
	if sConverter.Valid(puzzle) {
		//it's a valid sdk, just use that
		return sConverter.Load(grid, puzzle)
	}

	//It's a more complex doku.
	if strings.HasSuffix(puzzle, "\n") {
		puzzle = puzzle[:len(puzzle)-1]
	}

	rows := strings.Split(puzzle, "\n")
	for r, row := range rows {
		cols := strings.Split(row, "|")
		for c, col := range cols {
			parseDokuCell(col).fillCell(grid.MutableCell(r, c))
		}
	}

	return true
}

func (c *sdkConverter) Load(grid sudoku.MutableGrid, puzzle string) bool {
	if !c.Valid(puzzle) {
		return false
	}
	grid.Load(puzzle)
	return true
}

func (c *sdkConverter) DataString(grid sudoku.Grid) string {
	return grid.DataString()
}

func (c *sdkConverter) Valid(puzzle string) bool {

	//The SDK format is essentially just DIM * DIM characters that are either
	//0-DIM or a '.'. Col separators and newlines are stripped.

	//Strip out all newlines and col seps.
	puzzle = strings.Replace(puzzle, sudoku.COL_SEP, "", -1)
	puzzle = strings.Replace(puzzle, sudoku.ALT_COL_SEP, "", -1)
	puzzle = strings.Replace(puzzle, "\n", "", -1)
	puzzle = strings.Replace(puzzle, "\r", "", -1)

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
