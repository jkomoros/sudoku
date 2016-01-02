/*
i-sudoku is an interactive command-line sudoku tool
*/

package main

import (
	"fmt"
	"github.com/jkomoros/sudoku"
	"github.com/nsf/termbox-go"
	"log"
	"strconv"
	"strings"
	"unicode/utf8"
)

const STATUS_DEFAULT = "Type arrows to move, a number to input a number, 'm' to enter mark mode, or ESC to quit"
const STATUS_MARKING = "MARKING:"
const STATUS_MARKING_POSTFIX = "   ENTER to commit, ESC to cancel"

type mainModel struct {
	grid         *sudoku.Grid
	selected     *sudoku.Cell
	marksToInput []int
}

func main() {
	if err := termbox.Init(); err != nil {
		log.Fatal("Termbox initialization failed:", err)
	}
	defer termbox.Close()

	model := newModel()

	draw(model)

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				if model.ModeInputEsc() {
					break mainloop
				}
			case termbox.KeyCtrlC:
				break mainloop
			case termbox.KeyArrowDown:
				model.MoveSelectionDown()
			case termbox.KeyArrowLeft:
				model.MoveSelectionLeft()
			case termbox.KeyArrowRight:
				model.MoveSelectionRight()
			case termbox.KeyArrowUp:
				model.MoveSelectionUp()
			case termbox.KeyEnter:
				model.ModeCommitMarkMode()
			}
			switch ev.Ch {
			case 'q':
				break mainloop
			case 'm':
				//TODO: ideally Ctrl+Num would work to put in one mark. But termbox doesn't appear to let that work.
				model.ModeEnterMarkMode()
			case 'n':
				//TODO: since this is a destructive action, require a confirmation
				model.NewGrid()
			//TODO: do this in a more general way related to DIM
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				//TODO: this is a seriously gross way of converting a rune to a string.
				num, err := strconv.Atoi(strings.Replace(strconv.QuoteRuneToASCII(ev.Ch), "'", "", -1))
				if err != nil {
					panic(err)
				}
				model.ModeInputNumber(num)
			}
		}
		draw(model)
	}
}

func newModel() *mainModel {
	model := &mainModel{}
	model.EnsureSelected()
	return model
}

func (m *mainModel) StatusLine() string {
	//TODO: return something dynamic depending on mode.

	//TODO: in StatusLine, the keyboard shortcuts should be in bold.
	//Perhaps make it so at open parens set to bold, at close parens set
	//to normal.

	if m.marksToInput == nil {
		return STATUS_DEFAULT
	} else {
		//Marks mode
		return STATUS_MARKING + fmt.Sprint(m.marksToInput) + STATUS_MARKING_POSTFIX
	}
}

func (m *mainModel) Selected() *sudoku.Cell {
	return m.selected
}

func (m *mainModel) SetSelected(cell *sudoku.Cell) {
	if cell == m.selected {
		//Already done
		return
	}
	m.selected = cell
	m.ModeCancelMarkMode()
}

func (m *mainModel) ModeInputEsc() (quit bool) {
	if m.marksToInput != nil {
		m.ModeCancelMarkMode()
		return false
	}
	return true
}

func (m *mainModel) ModeEnterMarkMode() {
	//TODO: if already in Mark mode, ignore.
	selected := m.Selected()
	if selected != nil {
		if selected.Number() != 0 || selected.Locked() {
			//Dion't enter mark mode.
			return
		}
	}
	m.marksToInput = make([]int, 0)
}

func (m *mainModel) ModeCommitMarkMode() {
	for _, num := range m.marksToInput {
		m.ToggleSelectedMark(num)
	}
	m.marksToInput = nil
}

func (m *mainModel) ModeCancelMarkMode() {
	m.marksToInput = nil
}

func (m *mainModel) ModeInputNumber(num int) {
	if m.marksToInput == nil {
		m.SetSelectedNumber(num)
	} else {
		m.marksToInput = append(m.marksToInput, num)
	}
}

func (m *mainModel) EnsureSelected() {
	m.EnsureGrid()
	//Ensures that at least one cell is selected.
	if m.Selected() == nil {
		m.SetSelected(m.grid.Cell(0, 0))
	}
}

func (m *mainModel) MoveSelectionLeft() {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	c--
	if c < 0 {
		c = 0
	}
	m.SetSelected(m.grid.Cell(r, c))
}

func (m *mainModel) MoveSelectionRight() {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	c++
	if c >= sudoku.DIM {
		c = sudoku.DIM - 1
	}
	m.SetSelected(m.grid.Cell(r, c))
}

func (m *mainModel) MoveSelectionUp() {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	r--
	if r < 0 {
		r = 0
	}
	m.SetSelected(m.grid.Cell(r, c))
}

func (m *mainModel) MoveSelectionDown() {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	r++
	if r >= sudoku.DIM {
		r = sudoku.DIM - 1
	}
	m.SetSelected(m.grid.Cell(r, c))
}

func (m *mainModel) EnsureGrid() {
	if m.grid == nil {
		m.NewGrid()
	}
}

func (m *mainModel) NewGrid() {
	m.grid = sudoku.GenerateGrid(nil)
	m.grid.LockFilledCells()
}

func (m *mainModel) SetSelectedNumber(num int) {
	//TODO: if the number already has that number set, set 0.
	m.EnsureSelected()
	if m.Selected().Locked() {
		return
	}
	m.Selected().SetNumber(num)
}

func (m *mainModel) ToggleSelectedMark(num int) {
	m.EnsureSelected()
	if m.Selected().Locked() {
		return
	}
	m.Selected().SetMark(num, !m.Selected().Mark(num))
}

func clearScreen() {
	width, height := termbox.Size()
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func draw(model *mainModel) {

	//TODO: have a mode line after the grid for if the grid is invalid, if it's solved.

	clearScreen()

	grid := model.grid

	selectedTop, selectedLeft, selectedHeight, selectedWidth := model.Selected().DiagramExtents()

	x := 0
	y := 0

	for _, line := range strings.Split(grid.Diagram(true), "\n") {
		x = 0
		//The first number in range will be byte offset, but for some items like the bullet, it's two bytes.
		//But what we care about is that each item is a character.
		for _, ch := range line {

			defaultColor := termbox.ColorGreen

			numberRune, _ := utf8.DecodeRuneInString(sudoku.DIAGRAM_NUMBER)
			lockedRune, _ := utf8.DecodeRuneInString(sudoku.DIAGRAM_LOCKED)

			if ch == numberRune {
				defaultColor = termbox.ColorBlue
			} else if ch == lockedRune {
				defaultColor = termbox.ColorRed
			}

			backgroundColor := termbox.ColorDefault

			if x >= selectedTop && x < (selectedTop+selectedHeight) && y >= selectedLeft && y < (selectedLeft+selectedWidth) {
				//We're on the selected cell
				backgroundColor = termbox.ColorWhite
			}

			termbox.SetCell(x, y, ch, defaultColor, backgroundColor)
			x++
		}
		y++
	}

	x = 0
	for _, ch := range model.StatusLine() {
		termbox.SetCell(x, y, ch, termbox.ColorWhite, termbox.ColorDefault)
		x++
	}

	termbox.Flush()
}
