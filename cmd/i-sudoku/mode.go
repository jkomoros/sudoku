package main

import (
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"strconv"
	"strings"
)

var (
	MODE_DEFAULT    = &defaultMode{}
	MODE_COMMAND    = &commandMode{}
	MODE_CONFIRM    = &confirmMode{}
	MODE_FILE_INPUT = &fileInputMode{}
)

const (
	MARKS_MODE_FAIL_LOCKED         = "Can't enter mark mode on a cell that's locked."
	MARKS_MODE_FAIL_NUMBER         = "Can't enter mark mode on a cell that has a filled number."
	DEFAULT_MODE_FAIL_LOCKED       = "Can't enter a number in a locked cell."
	FAST_MODE_NO_OPEN_CELLS        = "Can't fast move: no more open cells in that direction"
	SINGLE_FILL_MORE_THAN_ONE_MARK = "The cell does not have precisely one mark set."
	GRID_LOADED_MESSAGE            = "Grid successfully loaded."
	HELP_MESSAGE                   = `The following commands are also available on this screen:
{c} to enter command mode to do things like quit and load a new puzzle
{h} to get a hint
	{Ctrl-h} to get a debug hint print-out
{+} or {=} to set the selected cell's marks to all legal marks
{-} to remove all invalid marks from the selected cell
{<enter>} to set a cell to the number that is the only current mark
{u} to undo a move
{r} to redo a move
{m} to enter mark mode, so all numbers entered will toggle marks
{f} to toggle fast move mode, allowing you to skip over filled cells
{s} to quick-save to the last used filename`
	STATUS_DEFAULT = "{→,←,↓,↑} to move cells, {0-9} to enter number, {Shift + 0-9} to toggle marks, {?} to list other commands"
	STATUS_COMMAND = "COMMAND: {n}ew puzzle, {q}uit, {l}oad puzzle..., {s}ave puzzle as..., {r}eset puzzle, {ESC} cancel"
	STATUS_LOAD    = "Filename? {Enter} to commit, {Esc} to cancel:"
)

func runeIsNum(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func runeIsShiftedNum(ch rune) bool {
	if runeIsNum(ch) {
		return false
	}
	return runeIsNum(shiftedNumRuneToNum(ch))
}

func shiftedNumRuneToNum(ch rune) rune {
	//Note: this assumes an american keyboard layout
	//TODO: make this resilient to other keyboard layouts, perhaps with a way
	//for the user to 'train' it.
	switch ch {
	case '!':
		return '1'
	case '@':
		return '2'
	case '#':
		return '3'
	case '$':
		return '4'
	case '%':
		return '5'
	case '^':
		return '6'
	case '&':
		return '7'
	case '*':
		return '8'
	case '(':
		return '9'
	case ')':
		return '0'
	}
	return ch
}

type InputMode interface {
	//TODO: doesn't it feel weird that every method takes a main model?
	handleInput(c *mainController, evt termbox.Event)
	shouldEnter(c *mainController) bool
	statusLine(c *mainController) string
	newCellSelected(c *mainController)
	//Cursor location is always in the status bar; it's a matter of how far to
	//right in that bar it is. -1 renders it offscreen, effectively meaning
	//'no cursor'
	cursorLocation(c *mainController) (x int)
}

type baseMode struct{}

func (s *baseMode) handleInput(c *mainController, evt termbox.Event) {
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyCtrlC:
			confirmQuit(c)
		}
	}
}

func (s *baseMode) cursorLocation(c *mainController) int {
	//By default the cursor should be offscreen.
	return -1
}

func (s *baseMode) statusLine(c *mainController) string {
	return STATUS_DEFAULT
}

func (s *baseMode) newCellSelected(c *mainController) {
	//Do nothing by default.
}

func (s *baseMode) shouldEnter(c *mainController) bool {
	return true
}

type defaultMode struct {
	baseMode
}

func (s *defaultMode) handleInput(c *mainController, evt termbox.Event) {

	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyArrowDown:
			c.MoveSelectionDown(c.FastMode())
		case termbox.KeyArrowLeft:
			c.MoveSelectionLeft(c.FastMode())
		case termbox.KeyArrowRight:
			c.MoveSelectionRight(c.FastMode())
		case termbox.KeyArrowUp:
			c.MoveSelectionUp(c.FastMode())
		case termbox.KeyCtrlH:
			c.ShowDebugHint()
		case termbox.KeyEsc:
			c.ClearConsole()
		case termbox.KeyEnter:
			if c.lastShownHint != nil {
				c.EnterHint()
			} else {
				c.SetSelectedToOnlyMark()
			}
		default:
			handled = false
		}
		switch {
		case evt.Ch == 'h':
			c.ShowHint()
		case evt.Ch == 'f':
			c.ToggleFastMode()
		case evt.Ch == '?', evt.Ch == '/':
			c.SetConsoleMessage(HELP_MESSAGE, true)
		case evt.Ch == '+', evt.Ch == '=':
			c.FillSelectedWithLegalMarks()
		case evt.Ch == '-':
			c.RemoveInvalidMarksFromSelected()
		case evt.Ch == 'c':
			c.EnterMode(MODE_COMMAND)
		case evt.Ch == 'm':
			c.ToggleMarkMode()
		case evt.Ch == 's':
			c.SaveCommandIssued()
		//TODO: test that u/r undo and redo
		case evt.Ch == 'u':
			c.Undo()
		case evt.Ch == 'r':
			c.Redo()
		case runeIsShiftedNum(evt.Ch):
			//TODO: ideally Ctrl+Num would work to put in one mark. But termbox doesn't appear to let that work.
			num, err := strconv.Atoi(string(shiftedNumRuneToNum(evt.Ch)))
			if err != nil {
				panic(err)
			}
			c.ToggleSelectedMark(num)
		case runeIsNum(evt.Ch):
			//TODO: this is a seriously gross way of converting a rune to a string.
			num, err := strconv.Atoi(string(evt.Ch))
			if err != nil {
				panic(err)
			}
			if c.MarkMode() {
				c.ToggleSelectedMark(num)
			} else {
				c.SetSelectedNumber(num)
			}
		default:
			if !handled {
				//neither handler handled it; defer to base.
				s.baseMode.handleInput(c, evt)
			}
		}
	}
}

type commandMode struct {
	baseMode
}

func confirmQuit(c *mainController) {
	c.enterConfirmMode("Quit? Your progress will be lost.",
		DEFAULT_NO,
		func() {
			c.exitNow = true
		},
		func() {},
	)
}

func (s *commandMode) handleInput(c *mainController, evt termbox.Event) {
	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyEsc:
			c.EnterMode(MODE_DEFAULT)
		default:
			handled = false
		}
		switch {
		case evt.Ch == 'q':
			confirmQuit(c)
		case evt.Ch == 'n':
			c.enterConfirmMode("Replace grid with a new one? This cannot be undone.",
				DEFAULT_NO,
				func() {
					c.NewGrid()
					c.EnterMode(MODE_DEFAULT)
				},
				func() {
					c.EnterMode(MODE_DEFAULT)
				},
			)
		case evt.Ch == 'r':
			c.enterConfirmMode("Reset? Your progress will be lost. This cannot be undone.",
				DEFAULT_NO,
				func() {
					c.ResetGrid()
					c.EnterMode(MODE_DEFAULT)
				},
				func() {
					c.EnterMode(MODE_DEFAULT)
				},
			)
		case evt.Ch == 'l':
			//TODO: confirm they want to load a puzzle and blow away state
			c.enterFileInputMode(func(input string) {
				c.LoadGridFromFile(input)
				c.EnterMode(MODE_DEFAULT)
			})
		case evt.Ch == 's':
			c.SaveAsCommandIssued()
		default:
			if !handled {
				//Neither of us handled it so defer to base.
				s.baseMode.handleInput(c, evt)
			}
		}
	}
}

func (s *commandMode) statusLine(c *mainController) string {
	return STATUS_COMMAND
}

type defaultOption int

const (
	DEFAULT_YES defaultOption = iota
	DEFAULT_NO
	DEFAULT_NONE
)

type confirmMode struct {
	msg           string
	defaultAction defaultOption
	yesAction     func()
	noAction      func()
	baseMode
}

func (s *confirmMode) handleInput(c *mainController, evt termbox.Event) {
	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyEnter:
			switch s.defaultAction {
			case DEFAULT_YES:
				s.yesAction()
				c.EnterMode(MODE_DEFAULT)
			case DEFAULT_NO:
				s.noAction()
				c.EnterMode(MODE_DEFAULT)
			case DEFAULT_NONE:
				//Don't do anything
			}
		default:
			handled = false
		}
		switch evt.Ch {
		case 'y', 'Y':
			s.yesAction()
			c.EnterMode(MODE_DEFAULT)
		case 'n', 'N':
			s.noAction()
			c.EnterMode(MODE_DEFAULT)
		default:
			if !handled {
				//Neither of us handled it so defer to base.
				s.baseMode.handleInput(c, evt)
			}
		}
	}
}

func (s *confirmMode) statusLine(c *mainController) string {
	confirmMsg := "{y}/{n}"
	if s.defaultAction == DEFAULT_YES {
		confirmMsg = "{Y}/{n}"
	} else if s.defaultAction == DEFAULT_NO {
		confirmMsg = "{y}/{N}"
	}
	return s.msg + "  " + confirmMsg
}

type fileInputMode struct {
	input string
	//which index of the input string the cursor is at
	cursorOffset int
	onCommit     func(string)
	baseMode
}

func (m *fileInputMode) cursorLocation(c *mainController) int {
	return len(STATUS_LOAD) + m.cursorOffset
}

func (m *fileInputMode) statusLine(c *mainController) string {
	return STATUS_LOAD + m.input
}

func (m *fileInputMode) shouldEnter(c *mainController) bool {
	m.input = ""
	m.cursorOffset = 0
	return true
}

func (m *fileInputMode) moveCursorLeft() {
	m.cursorOffset--
	if m.cursorOffset < 0 {
		m.cursorOffset = 0
	}
}

func (m *fileInputMode) moveCursorRight() {
	m.cursorOffset++
	if m.cursorOffset > len(m.input) {
		m.cursorOffset = len(m.input)
	}
}

func (m *fileInputMode) moveCursorToFront() {
	m.cursorOffset = 0
}

func (m *fileInputMode) moveCursorToBack() {
	m.cursorOffset = len(m.input)
}

func (m *fileInputMode) deleteToEndOfLine() {
	m.input = m.input[:m.cursorOffset]
}

func (m *fileInputMode) addCharAtCursor(ch rune) {
	m.input = m.input[0:m.cursorOffset] + string(ch) + m.input[m.cursorOffset:len(m.input)]
	m.moveCursorRight()
}

func (m *fileInputMode) loadPuzzle(c *mainController) {
	c.LoadGridFromFile(m.input)
	c.EnterMode(MODE_DEFAULT)
}

func (m *fileInputMode) removeCharAtCursor() {
	if len(m.input) == 0 {
		return
	}
	if len(m.input) == m.cursorOffset {
		//Treat removing at end specially, since otherwise we'd index off the
		//end of the string.
		m.input = m.input[:m.cursorOffset-1]
	} else if m.cursorOffset == 0 {
		//Avoid indexing off the lower end
		m.input = m.input[m.cursorOffset:len(m.input)]
	} else {
		m.input = m.input[:m.cursorOffset-1] + m.input[m.cursorOffset:len(m.input)]
	}
	m.moveCursorLeft()
}

func (m *fileInputMode) tabComplete(c *mainController) {
	//Only do tab complete if at end of input
	if m.cursorOffset != len(m.input) {
		return
	}
	//Interpret input as a path. Split everthing before the last '/' as the directory to scan
	//and everything after as the prefix to filter those things with.
	//TODO: change this sep on other platforms?
	splitPath := strings.Split(m.input, "/")
	directoryPortion := strings.Join(splitPath[:len(splitPath)-1], "/")

	if directoryPortion != "" {
		directoryPortion += "/"
	}

	rest := splitPath[len(splitPath)-1]
	possibleCompletions, err := ioutil.ReadDir("./" + directoryPortion)

	if err != nil {
		c.SetConsoleMessage("No valid directory"+err.Error(), true)
		return
	}

	//Now, process possibleCompletions, filtering down ones that don't have the prefix.
	var matchedCompletions []string
	for _, completion := range possibleCompletions {
		if strings.HasPrefix(completion.Name(), rest) {
			//Found one!

			completionToAdd := completion.Name()
			if completion.IsDir() {
				completionToAdd += "/"
			}
			matchedCompletions = append(matchedCompletions, completionToAdd)
		}
	}

	if len(matchedCompletions) == 1 {
		//There's only one completion!
		m.input = directoryPortion + matchedCompletions[0]
		m.cursorOffset = len(m.input)
	} else if len(matchedCompletions) > 0 {
		prefix := longestCommonPrefix(matchedCompletions)
		//In some cases the prefix will be the part already typed in; in that
		//case, print completions.
		if prefix != "" && prefix != rest {
			m.input = directoryPortion + prefix
			m.cursorOffset = len(m.input)
		} else {
			//Just list the valid completions

			//We'll print out strings where the part that's not prefixed is bolded.
			var unprefixedMatchedCompletions []string
			for _, match := range matchedCompletions {
				unprefixedMatchedCompletions = append(unprefixedMatchedCompletions, prefix+"{"+strings.TrimPrefix(match, prefix)+"}")
			}

			c.SetConsoleMessage("Possible completions\n"+strings.Join(unprefixedMatchedCompletions, "\n"), true)
		}
	} else if len(matchedCompletions) == 0 {
		c.SetConsoleMessage("{No valid completions}", true)
	}
}

func longestCommonPrefix(files []string) string {
	//This implementation was inspired by some of the examples on
	//http://rosettacode.org/wiki/Longest_common_prefix#
	if len(files) == 0 {
		return ""
	}
	if len(files) == 1 {
		return files[0]
	}
	//find the min and max string,sorted by bytes
	min := files[0]
	max := files[1]
	for _, file := range files {
		if file < min {
			min = file
		}
		if file > max {
			max = file
		}
	}
	//Find the longestCommonPrefix of the min and max
	maxIndex := len(min)
	if len(max) < maxIndex {
		maxIndex = len(max)
	}
	for i := 0; i < maxIndex; i++ {
		if min[i] != max[i] {
			//Found first bit where they differ, return everything up until this point
			return min[:i]
		}
	}
	//If we walked through min and max and found no differences, just return min
	return min
}

func (m *fileInputMode) handleInput(c *mainController, evt termbox.Event) {
	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyEnter:
			if m.onCommit != nil {
				m.onCommit(m.input)
			}
		case termbox.KeyEsc:
			c.EnterMode(MODE_DEFAULT)
		case termbox.KeyArrowLeft:
			m.moveCursorLeft()
		case termbox.KeyArrowRight:
			m.moveCursorRight()
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			m.removeCharAtCursor()
		case termbox.KeyTab:
			m.tabComplete(c)
		case termbox.KeyCtrlA:
			m.moveCursorToFront()
		case termbox.KeyCtrlE:
			m.moveCursorToBack()
		case termbox.KeyCtrlK:
			m.deleteToEndOfLine()
		default:
			handled = false
		}
		switch {
		case evt.Ch != 0:
			m.addCharAtCursor(evt.Ch)
		default:
			if !handled {
				//Neither of us handled it so defer to base.
				m.baseMode.handleInput(c, evt)
			}
		}
	}
}
