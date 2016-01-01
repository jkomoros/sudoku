/*
i-sudoku is an interactive command-line sudoku tool
*/

package main

import (
	"github.com/jkomoros/sudoku"
	"github.com/nsf/termbox-go"
	"log"
	"strconv"
	"strings"
	"unicode/utf8"
)

type mainModel struct {
	grid     *sudoku.Grid
	Selected *sudoku.Cell
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
			case termbox.KeyEsc, termbox.KeyCtrlC:
				break mainloop
			case termbox.KeyArrowDown:
				model.MoveSelectionDown()
			case termbox.KeyArrowLeft:
				model.MoveSelectionLeft()
			case termbox.KeyArrowRight:
				model.MoveSelectionRight()
			case termbox.KeyArrowUp:
				model.MoveSelectionUp()
			}
			switch ev.Ch {
			case 'q':
				break mainloop
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
				model.SetSelectedNumber(num)
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

func (m *mainModel) EnsureSelected() {
	m.EnsureGrid()
	//Ensures that at least one cell is selected.
	if m.Selected == nil {
		m.Selected = m.grid.Cell(0, 0)
	}
}

func (m *mainModel) MoveSelectionLeft() {
	m.EnsureSelected()
	r := m.Selected.Row()
	c := m.Selected.Col()
	c--
	if c < 0 {
		c = 0
	}
	m.Selected = m.grid.Cell(r, c)
}

func (m *mainModel) MoveSelectionRight() {
	m.EnsureSelected()
	r := m.Selected.Row()
	c := m.Selected.Col()
	c++
	if c >= sudoku.DIM {
		c = sudoku.DIM - 1
	}
	m.Selected = m.grid.Cell(r, c)
}

func (m *mainModel) MoveSelectionUp() {
	m.EnsureSelected()
	r := m.Selected.Row()
	c := m.Selected.Col()
	r--
	if r < 0 {
		r = 0
	}
	m.Selected = m.grid.Cell(r, c)
}

func (m *mainModel) MoveSelectionDown() {
	m.EnsureSelected()
	r := m.Selected.Row()
	c := m.Selected.Col()
	r++
	if r >= sudoku.DIM {
		r = sudoku.DIM - 1
	}
	m.Selected = m.grid.Cell(r, c)
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
	m.EnsureSelected()
	m.Selected.SetNumber(num)
}

func draw(model *mainModel) {
	drawGrid(model)
	termbox.Flush()
}

func drawGrid(model *mainModel) {

	grid := model.grid

	selectedTop, selectedLeft, selectedHeight, selectedWidth := model.Selected.DiagramExtents()

	for y, line := range strings.Split(grid.Diagram(true), "\n") {
		x := 0
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
	}
}
