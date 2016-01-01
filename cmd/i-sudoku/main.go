/*
i-sudoku is an interactive command-line sudoku tool
*/

package main

import (
	"github.com/jkomoros/sudoku"
	"github.com/nsf/termbox-go"
	"log"
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
			}
		}
		draw(model)
	}
}

func newModel() *mainModel {
	model := &mainModel{
		sudoku.GenerateGrid(nil),
		nil,
	}
	model.EnsureSelected()
	return model
}

func (m *mainModel) EnsureSelected() {
	//Ensures that at least one cell is selected.
	if m.Selected == nil {
		m.Selected = m.grid.Cell(0, 0)
	}
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

			if ch == numberRune {
				defaultColor = termbox.ColorBlue
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
