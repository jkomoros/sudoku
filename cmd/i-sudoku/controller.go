package main

import (
	"fmt"
	"github.com/gosuri/uitable"
	"github.com/jkomoros/sudoku"
	"github.com/jkomoros/sudoku/sdkconverter"
	"github.com/mitchellh/go-wordwrap"
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	PUZZLE_SAVED_MESSAGE = "Puzzle saved to "
)

type mainController struct {
	model    *model
	selected *sudoku.Cell
	mode     InputMode
	//The size of the console output. Not used for much.
	outputWidth int
	//What the diagram(true) of the grid looked like at last save.
	snapshot string
	filename string
	//If we load up a file, we aren't sure that overwriting the file is OK.
	//This stores whether we've verified that we can save here.
	fileOKToSave   bool
	lastShownHint  *sudoku.SolveDirections
	consoleMessage string
	//if true, will zero out console message on turn of event loop.
	consoleMessageShort bool
	//If exitNow is flipped to true, we will quit at next turn of event loop.
	exitNow bool
	toggles []toggle
}

const (
	TOGGLE_SOLVED = iota
	TOGGLE_INVALID
	TOGGLE_MARK_MODE
	TOGGLE_FAST_MODE
	TOGGLE_UNSAVED
)

type toggle struct {
	Value     func() bool
	Toggle    func()
	OnText    string
	OffText   string
	Color     termbox.Attribute
	ColorGrid bool
}

func newController() *mainController {
	c := &mainController{
		model: &model{},
		mode:  MODE_DEFAULT,
	}
	c.setUpToggles()
	c.EnsureSelected()
	return c
}

func (c *mainController) setUpToggles() {

	//State variable for the closure
	var fastMode bool
	var markMode bool

	c.toggles = []toggle{
		//Solved
		{
			func() bool {
				return c.Grid().Solved()
			},
			func() {
				//Do nothing; read only
			},
			"  SOLVED  ",
			" UNSOLVED ",
			termbox.ColorYellow,
			true,
		},
		//invalid
		{
			func() bool {
				return c.Grid().Invalid()
			},
			func() {
				//Read only
			},
			" INVALID ",
			"  VALID  ",
			termbox.ColorRed,
			true,
		},
		//Mark mode
		{
			func() bool {
				return markMode
			},
			func() {
				markMode = !markMode
			},
			" MARKING ",
			"         ",
			termbox.ColorCyan,
			false,
		},
		//Fast mode
		{
			func() bool {
				return fastMode
			},
			func() {
				fastMode = !fastMode
			},
			"  FAST MODE  ",
			"             ",
			termbox.ColorBlue,
			false,
		},
		//Unsaved
		{
			func() bool {
				return !c.IsSaved()
			},
			func() {
				//Read only
			},
			" UNSAVED ",
			"  SAVED  ",
			termbox.ColorMagenta,
			false,
		},
	}
}

//EnterState attempts to set the controller to the given state. The state
//object is given a chance to do initalization and potentially cancel the
//transition, leaving the controller in the same state as before.
func (c *mainController) EnterMode(state InputMode) {
	//SetState doesn't do much, it just makes it feel less weird than
	//STATE.enter(m) (which feels backward)

	if state.shouldEnter(c) {
		c.mode = state
	}
}

//enterConfirmState is a special state to set
func (c *mainController) enterConfirmMode(msg string, defaultAction defaultOption, yesAction func(), noAction func()) {
	MODE_CONFIRM.msg = msg
	MODE_CONFIRM.defaultAction = defaultAction
	MODE_CONFIRM.yesAction = yesAction
	MODE_CONFIRM.noAction = noAction
	c.EnterMode(MODE_CONFIRM)
}

func (c *mainController) enterFileInputMode(onCommit func(string)) {
	MODE_FILE_INPUT.onCommit = onCommit
	c.EnterMode(MODE_FILE_INPUT)
}

func (c *mainController) SetConsoleMessage(msg string, shortLived bool) {

	if c.outputWidth != 0 {
		//Wrap to fit in given size
		msg = wordwrap.WrapString(msg, uint(c.outputWidth))
	}

	c.consoleMessage = msg
	c.consoleMessageShort = shortLived
	c.lastShownHint = nil
}

//WillProcessEvent is cleared right before we call handleInput on the current
//state--that is, right before we process an event. That is a convenient time
//to clear state and prepare for the next state. This is *not* called before a
//timer/display tick.
func (c *mainController) WillProcessEvent() {
	if c.consoleMessageShort {
		c.ClearConsole()
	}
}

func (c *mainController) ClearConsole() {
	c.consoleMessage = ""
	c.consoleMessageShort = false
	c.lastShownHint = nil
}

func (c *mainController) StatusLine() string {
	return c.mode.statusLine(c)
}

//TODO: should this vend a copy of the grid? I want to make it so the only
//easy way to mutate the grid is via model mutators.
func (c *mainController) Grid() *sudoku.Grid {
	return c.model.grid
}

func (c *mainController) SetGrid(grid *sudoku.Grid) {
	oldCell := c.Selected()
	c.model.SetGrid(grid)
	//The currently selected cell is tied to the grid, so we need to fix it up.
	if oldCell != nil {
		c.SetSelected(oldCell.InGrid(c.model.grid))
	}
	if c.model.grid != nil {
		//IF there are already some locked cells, we assume that only those
		//cells should be locked. If there aren't any locked cells at all, we
		//assume that all filled cells should be locked.

		//TODO: this seems like magic behavior that's hard to reason about.
		foundLockedCell := false
		for _, cell := range c.model.grid.Cells() {
			if cell.Locked() {
				foundLockedCell = true
				break
			}
		}
		if !foundLockedCell {
			c.model.grid.LockFilledCells()
		}
	}
	c.snapshot = ""
}

func (c *mainController) LoadGridFromFile(file string) {

	if file == "" {
		return
	}

	puzzleBytes, err := ioutil.ReadFile(file)

	if err != nil {
		c.SetConsoleMessage("Invalid file: "+err.Error(), true)
		return
	}
	puzzle := string(puzzleBytes)

	if sdkconverter.Format(puzzle) == "" {
		c.SetConsoleMessage("Provided puzzle is in unknown format.", true)
		return
	}

	c.SetGrid(sdkconverter.Load(puzzle))
	c.SetConsoleMessage(GRID_LOADED_MESSAGE, true)
	c.filename = file
	c.fileOKToSave = false
	c.saveSnapshot()
}

//Actually save
func (c *mainController) SaveGrid() {

	if c.filename == "" {
		return
	}

	if !c.fileOKToSave {
		return
	}

	converter := sdkconverter.Converters["doku"]

	//TODO: if the filename doesn't have an extension, add doku.

	if converter == nil {
		return
	}

	ioutil.WriteFile(c.filename, []byte(converter.DataString(c.Grid())), 0644)

	c.SetConsoleMessage(PUZZLE_SAVED_MESSAGE+c.filename, true)

	c.saveSnapshot()
}

func (c *mainController) SaveCommandIssued() {
	c.saveCommandImpl(false)
}

func (c *mainController) SaveAsCommandIssued() {
	c.saveCommandImpl(true)
}

//The user told us to save. what we actually do depends on current state.
func (c *mainController) saveCommandImpl(forceFilenamePrompt bool) {
	if c.filename == "" || forceFilenamePrompt {
		c.enterFileInputMode(func(input string) {
			if _, err := os.Stat(input); err == nil {
				//The file exists. Confirm.
				_ = "breakpoint"
				c.enterConfirmMode(input+" already exists. Overwrite?",
					DEFAULT_YES,
					func() {
						c.SetFilename(input)
						c.SaveCommandIssued()
					},
					func() {
						//Don't write and try again.
						c.SaveCommandIssued()
					},
				)
				return
			}
			c.SetFilename(input)
			c.SaveGrid()
			c.EnterMode(MODE_DEFAULT)
		})
		return
	}
	if !c.fileOKToSave {
		c.enterConfirmMode("OK to save to "+c.filename+"?",
			DEFAULT_YES,
			func() {
				c.fileOKToSave = true
				c.SaveCommandIssued()
			},
			func() {
				c.SetFilename("")
				c.SaveCommandIssued()
			},
		)
		return
	}
	c.SaveGrid()
	c.EnterMode(MODE_DEFAULT)
}

func (c *mainController) Filename() string {
	return c.filename
}

func (c *mainController) saveSnapshot() {
	if c.Grid() == nil {
		c.snapshot = ""
	} else {
		c.snapshot = c.Grid().Diagram(true)
	}
}

func (c *mainController) IsSaved() bool {
	if c.Grid() == nil {
		return true
	}
	return c.Grid().Diagram(true) == c.snapshot
}

func (c *mainController) SetFilename(filename string) {
	c.filename = filename
	c.fileOKToSave = true
}

//Just used for sorting the steps by probabilities
type nextSteps struct {
	steps         []*sudoku.CompoundSolveStep
	probabilities []float64
}

func (n *nextSteps) Len() int {
	return len(n.steps)
}

func (n nextSteps) Less(i, j int) bool {
	return n.probabilities[i] > n.probabilities[j]
}

func (n nextSteps) Swap(i, j int) {
	n.steps[i], n.steps[j] = n.steps[j], n.steps[i]
	n.probabilities[i], n.probabilities[j] = n.probabilities[j], n.probabilities[i]
}

//countedNums is used in ShowCount
type countedNums struct {
	num   int
	count int
}

type countedNumsSlice []countedNums

func (s countedNumsSlice) Len() int {
	return len(s)
}

func (s countedNumsSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s countedNumsSlice) Less(i, j int) bool {
	//This number is the reverse of the order we'll print it out (DIM - num)
	return s[i].count > s[j].count
}

func (c *mainController) ShowCount() {
	//TODO: test this

	nums := make(countedNumsSlice, sudoku.DIM)
	for i := 0; i < sudoku.DIM; i++ {
		nums[i] = countedNums{i + 1, 0}
	}

	for _, cell := range c.Grid().Cells() {
		num := cell.Number()
		if num <= 0 || num > sudoku.DIM {
			continue
		}
		nums[num-1].count++
	}

	sort.Sort(nums)

	lowestCount := -1

	msg := "{Counts remaining}:\n"
	for i := 0; i < sudoku.DIM; i++ {
		currentNum := nums[i].num
		count := sudoku.DIM - nums[i].count
		countString := strconv.Itoa(count)
		numString := strconv.Itoa(currentNum)
		if count == 0 {
			numString = "{" + numString + "}"
			countString = "{0}"
		} else if lowestCount == -1 || lowestCount == count {
			//If this is the lowest count that's not zero, highlight it, since
			//it's most important.
			numString = "{" + numString + "}"
			countString = "{" + countString + "}"
			if lowestCount == -1 {
				lowestCount = count
			}
		}
		msg += "Number " + numString + " : " + countString + " remaining\n"
	}
	c.SetConsoleMessage(msg, true)
}

func (c *mainController) ShowDebugHint() {
	options := sudoku.DefaultHumanSolveOptions()
	options.NumOptionsToCalculate = 100

	//TODO: this feels like a HORRENDOUS hack, and very brittle. :-(
	fakeLastFillSteps := []*sudoku.CompoundSolveStep{
		{
			FillStep: &sudoku.SolveStep{
				Technique:   sudoku.Techniques[0],
				TargetCells: c.model.LastModifiedCells(),
			},
		},
	}

	steps, probabilities := c.Grid().HumanSolvePossibleSteps(options, fakeLastFillSteps)

	sort.Sort(&nextSteps{steps, probabilities})

	msg := "{Possible steps} (" + strconv.Itoa(len(steps)) + " possibilities)\n"

	table := uitable.New()

	cumulative := 0.0

	for i, step := range steps {
		cumulative += probabilities[i] * 100
		fillStep := step.FillStep

		var precusorStepsDescription []string

		for _, step := range step.PrecursorSteps {
			precusorStepsDescription = append(precusorStepsDescription, step.TechniqueVariant())
		}

		table.AddRow(strconv.Itoa(i), fmt.Sprintf("%4.2f", probabilities[i]*100)+"%", fmt.Sprintf("%4.2f", cumulative)+"%", fillStep.TechniqueVariant(), fillStep.TargetCells.Description(), fillStep.TargetNums.Description(), len(step.PrecursorSteps), strings.Join(precusorStepsDescription, ", "))
	}

	c.SetConsoleMessage(msg+table.String(), true)
}

func (c *mainController) ShowHint() {
	options := sudoku.DefaultHumanSolveOptions()
	options.NumOptionsToCalculate = 100
	hint := c.Grid().Hint(options)

	if hint == nil || len(hint.CompoundSteps) == 0 {
		c.SetConsoleMessage("No hint to give.", true)
		return
	}
	c.SetConsoleMessage("{Hint}\n"+strings.Join(hint.Description(), "\n")+"\n\n"+"{ENTER} to accept, {ESC} to ignore", false)
	//This hast to be after setting console message, since SetConsoleMessage clears the last hint.
	c.lastShownHint = hint
	lastStep := hint.CompoundSteps[0].FillStep
	c.SetSelected(lastStep.TargetCells[0].InGrid(c.Grid()))
}

func (c *mainController) EnterHint() {
	if c.lastShownHint == nil {
		return
	}
	lastStep := c.lastShownHint.CompoundSteps[0].FillStep
	cell := lastStep.TargetCells[0]
	num := lastStep.TargetNums[0]

	c.SetSelected(cell.InGrid(c.Grid()))
	c.SetSelectedNumber(num)

	c.ClearConsole()
}

func (c *mainController) Selected() *sudoku.Cell {
	return c.selected
}

func (c *mainController) SetSelected(cell *sudoku.Cell) {
	if cell == c.selected {
		//Already done
		return
	}
	c.selected = cell
	c.mode.newCellSelected(c)
}

func (c *mainController) EnsureSelected() {
	c.EnsureGrid()
	//Ensures that at least one cell is selected.
	if c.Selected() == nil {
		c.SetSelected(c.Grid().Cell(0, 0))
	}
}

func (c *mainController) MoveSelectionLeft(fast bool) {
	c.EnsureSelected()
	row := c.Selected().Row()
	col := c.Selected().Col()
	for {
		col--
		if col < 0 {
			col = 0
		}
		if fast && c.Grid().Cell(row, col).Number() != 0 {
			if col == 0 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				c.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		c.SetSelected(c.Grid().Cell(row, col))
		break
	}
}

func (c *mainController) MoveSelectionRight(fast bool) {
	c.EnsureSelected()
	row := c.Selected().Row()
	col := c.Selected().Col()
	for {
		col++
		if col >= sudoku.DIM {
			col = sudoku.DIM - 1
		}
		if fast && c.Grid().Cell(row, col).Number() != 0 {
			if col == sudoku.DIM-1 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				c.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		c.SetSelected(c.Grid().Cell(row, col))
		break
	}
}

func (m *mainController) MoveSelectionUp(fast bool) {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	for {
		r--
		if r < 0 {
			r = 0
		}
		if fast && m.Grid().Cell(r, c).Number() != 0 {
			if r == 0 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				m.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		m.SetSelected(m.Grid().Cell(r, c))
		break
	}
}

func (m *mainController) MoveSelectionDown(fast bool) {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	for {
		r++
		if r >= sudoku.DIM {
			r = sudoku.DIM - 1
		}
		if fast && m.Grid().Cell(r, c).Number() != 0 {
			if r == sudoku.DIM-1 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				m.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		m.SetSelected(m.Grid().Cell(r, c))
		break
	}
}

func (m *mainController) FastMode() bool {
	return m.toggles[TOGGLE_FAST_MODE].Value()
}

func (m *mainController) ToggleFastMode() {
	m.toggles[TOGGLE_FAST_MODE].Toggle()
}

func (m *mainController) MarkMode() bool {
	return m.toggles[TOGGLE_MARK_MODE].Value()
}

func (m *mainController) ToggleMarkMode() {
	m.toggles[TOGGLE_MARK_MODE].Toggle()
}

func (m *mainController) EnsureGrid() {
	if m.Grid() == nil {
		m.NewGrid()
	}
}

func (c *mainController) NewGrid() {
	c.SetGrid(sudoku.GenerateGrid(nil))
	c.filename = ""
	c.fileOKToSave = false
}

func (c *mainController) ResetGrid() {
	c.Grid().ResetUnlockedCells()
}

func (c *mainController) Undo() {
	//TODO: test that this prints a message if there's nothing to undo.
	if c.model.Undo() {
		c.SetConsoleMessage("Undid one move.", true)
	} else {
		c.SetConsoleMessage("No moves to undo.", true)
	}
}

func (c *mainController) Redo() {
	//TODO: test that this prints a message if there's nothing to redo.
	if c.model.Redo() {
		c.SetConsoleMessage("Redid one move.", true)
	} else {
		c.SetConsoleMessage("No moves to redo.", true)
	}
}

//If the selected cell has only one mark, fill it.
func (c *mainController) SetSelectedToOnlyMark() {
	c.EnsureSelected()
	marks := c.Selected().Marks()
	if len(marks) != 1 {
		c.SetConsoleMessage(SINGLE_FILL_MORE_THAN_ONE_MARK, true)
		return
	}
	//Rely on SetSelectedNumber to barf if it's not allowed for some other reason.
	c.SetSelectedNumber(marks[0])
}

func (c *mainController) SetSelectedNumber(num int) {
	c.EnsureSelected()
	if c.Selected().Locked() {
		c.SetConsoleMessage(DEFAULT_MODE_FAIL_LOCKED, true)
		return
	}

	if c.Selected().Number() != num {
		c.model.SetNumber(c.Selected().Row(), c.Selected().Col(), num)
	} else {
		//If the number to set is already set, then empty the cell instead.
		c.model.SetNumber(c.Selected().Row(), c.Selected().Col(), 0)
	}

	c.checkHintDone()
}

func (c *mainController) checkHintDone() {
	if c.lastShownHint == nil {
		return
	}
	lastStep := c.lastShownHint.CompoundSteps[0].FillStep
	num := lastStep.TargetNums[0]
	cell := lastStep.TargetCells[0]
	if cell.InGrid(c.Grid()).Number() == num {
		c.ClearConsole()
	}
}

func (c *mainController) ToggleSelectedMark(num int) {
	c.EnsureSelected()
	if c.Selected().Locked() {
		c.SetConsoleMessage(MARKS_MODE_FAIL_LOCKED, true)
		return
	}
	if c.Selected().Number() != 0 {
		c.SetConsoleMessage(MARKS_MODE_FAIL_NUMBER, true)
		return
	}
	c.model.SetMarks(c.Selected().Row(), c.Selected().Col(), map[int]bool{num: !c.Selected().Mark(num)})
}

func (c *mainController) FillAllLegalMarks() {
	c.model.StartGroup()

	for _, cell := range c.Grid().Cells() {
		if cell.Number() != 0 {
			continue
		}
		markMap := make(map[int]bool)
		for _, num := range cell.Possibilities() {
			markMap[num] = true
		}
		c.model.SetMarks(cell.Row(), cell.Col(), markMap)
	}

	c.model.FinishGroupAndExecute()
}

func (c *mainController) RemovedInvalidMarksFromAll() {
	c.model.StartGroup()

	for _, cell := range c.Grid().Cells() {
		if cell.Number() != 0 {
			continue
		}
		markMap := make(map[int]bool)
		for _, num := range cell.Marks() {
			if !cell.Possible(num) {
				markMap[num] = false
			}
		}
		c.model.SetMarks(cell.Row(), cell.Col(), markMap)
	}

	c.model.FinishGroupAndExecute()
}

func (c *mainController) FillSelectedWithLegalMarks() {
	c.EnsureSelected()
	c.Selected().ResetMarks()
	markMap := make(map[int]bool)
	for _, num := range c.Selected().Possibilities() {
		markMap[num] = true
	}
	c.model.SetMarks(c.Selected().Row(), c.Selected().Col(), markMap)
}

func (c *mainController) RemoveInvalidMarksFromSelected() {
	c.EnsureSelected()
	markMap := make(map[int]bool)
	for _, num := range c.Selected().Marks() {
		if !c.Selected().Possible(num) {
			markMap[num] = false
		}
	}
	c.model.SetMarks(c.Selected().Row(), c.Selected().Col(), markMap)
}
