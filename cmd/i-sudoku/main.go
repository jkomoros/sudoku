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

const STATUS_DEFAULT = "(→,←,↓,↑) to move cells, (0-9) to enter number, (m)ark mode, other (c)ommand"
const STATUS_MARKING = "MARKING:"
const STATUS_MARKING_POSTFIX = "  (1-9) to toggle marks, (ENTER) to commit, (ESC) to cancel"
const STATUS_COMMAND = "COMMAND: (q)uit, (n)ew puzzle, (ESC) cancel"

const GRID_INVALID = " INVALID "
const GRID_VALID = "  VALID  "
const GRID_SOLVED = "  SOLVED  "
const GRID_NOT_SOLVED = " UNSOLVED "

func main() {

	//TODO: should be possible to run it and pass in a puzzle to use.

	if err := termbox.Init(); err != nil {
		log.Fatal("Termbox initialization failed:", err)
	}
	defer termbox.Close()

	termbox.SetOutputMode(termbox.Output256)

	model := newModel()

	width, _ := termbox.Size()
	model.outputWidth = width

	draw(model)

mainloop:
	for {
		evt := termbox.PollEvent()
		switch evt.Type {
		case termbox.EventKey:
			model.state.handleInput(model, evt)
		}
		draw(model)
		if model.exitNow {
			break mainloop
		}
		model.EndOfEventLoop()
	}
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
			} else if runeIsNum(ch) {
				defaultColor = termbox.ColorGreen | termbox.AttrBold
			}

			backgroundColor := termbox.ColorDefault

			if x >= selectedTop && x < (selectedTop+selectedHeight) && y >= selectedLeft && y < (selectedLeft+selectedWidth) {
				//We're on the selected cell
				backgroundColor = 0xf0
			}

			termbox.SetCell(x, y, ch, defaultColor, backgroundColor)
			x++
		}
		y++
	}

	x = 0
	solvedMsg := GRID_NOT_SOLVED
	fg := termbox.ColorBlue
	bg := termbox.ColorBlack
	if grid.Solved() {
		solvedMsg = GRID_SOLVED
		fg, bg = bg, fg
	}

	for _, ch := range solvedMsg {
		termbox.SetCell(x, y, ch, fg, bg)
		x++
	}

	//don't reset x; this next message should go to the right.
	validMsg := GRID_VALID
	fg = termbox.ColorBlue
	bg = termbox.ColorBlack
	if grid.Invalid() {
		validMsg = GRID_INVALID
		fg, bg = bg, fg
	}
	for _, ch := range validMsg {
		termbox.SetCell(x, y, ch, fg, bg)
		x++
	}

	y++

	x = 0
	underlined := false
	for _, ch := range model.StatusLine() {
		//The ( and ) are non-printing control characters
		if ch == '(' {
			underlined = true
			continue
		} else if ch == ')' {
			underlined = false
			continue
		}
		fg := termbox.ColorWhite
		if underlined {
			fg = fg | termbox.AttrUnderline | termbox.AttrBold
		}

		termbox.SetCell(x, y, ch, fg, termbox.ColorDefault)
		x++
	}

	underlined = false

	splitMessage := strings.Split(model.consoleMessage, "\n")

	for _, line := range splitMessage {
		y++
		x = 0
		for _, ch := range line {
			if ch == '(' {
				underlined = true
				continue
			} else if ch == ')' {
				underlined = false
				continue
			}
			fg := termbox.ColorWhite
			if underlined {
				fg = fg | termbox.AttrBold
			}
			termbox.SetCell(x, y, ch, fg, termbox.ColorBlack)
			x++
		}
	}

	termbox.Flush()
}
