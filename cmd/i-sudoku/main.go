/*
i-sudoku is an interactive command-line sudoku tool. Although its main
puropse is debugging the sudoku package, it's a reasonably full-featured
command-line sudoku game in its own right.
*/

package main

import (
	"flag"
	"fmt"
	"github.com/jkomoros/sudoku"
	"github.com/jkomoros/sudoku/sdkconverter"
	"github.com/nsf/termbox-go"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

//A debug override; if true will print a color palette to the screen, wait for
//a keypress, and then quit. Useful for seeing what different colors are
//available to use.
const DRAW_PALETTE = false

var tickCount int

type appOptions struct {
	START_PUZZLE_FILENAME string
	//accounting
	flagSet *flag.FlagSet
}

func defineFlags(options *appOptions) {
	options.flagSet.StringVar(&options.START_PUZZLE_FILENAME, "p", "", "filename of a puzzle (any format) to start from")
}

//If it returns true, the program should quit.
func (o *appOptions) fixUp(errOutput io.ReadWriter) bool {
	//No processing necessary yet
	return false
}

func getOptions(flagSet *flag.FlagSet, flagArguments []string, errOutput io.ReadWriter) *appOptions {
	options := &appOptions{flagSet: flagSet}
	defineFlags(options)
	flagSet.Parse(flagArguments)
	if options.fixUp(errOutput) {
		os.Exit(1)
	}
	return options
}

func main() {

	flagSet := flag.CommandLine
	options := getOptions(flagSet, os.Args[1:], nil)
	model := makeMainModel(options, os.Stderr)

	if err := termbox.Init(); err != nil {
		log.Fatal("Termbox initialization failed:", err)
	}
	defer termbox.Close()

	termbox.SetOutputMode(termbox.Output256)

	if DRAW_PALETTE {
		drawColorPalette()
		//Wait until something happens, generally a key is pressed.
		termbox.PollEvent()
		return
	}

	width, _ := termbox.Size()
	model.outputWidth = width

	mainLoop(model)

}

func makeMainModel(options *appOptions, errOutput io.ReadWriter) *mainController {

	logger := log.New(errOutput, "", log.LstdFlags)

	model := newModel()

	if options.START_PUZZLE_FILENAME != "" {
		puzzleBytes, err := ioutil.ReadFile(options.START_PUZZLE_FILENAME)

		if err != nil {
			logger.Fatalln("Invalid file:", err)
		}
		puzzle := string(puzzleBytes)

		if sdkconverter.Format(puzzle) == "" {
			logger.Fatalln("Provided puzzle is in unknown format.")
		}

		//TODO: this logic should all be wrapped in SetGrid logic, since it's
		//so finicky.

		model.grid = sdkconverter.Load(puzzle)
		model.grid.LockFilledCells()
		model.SetSelected(nil)
		model.EnsureSelected()

	}

	return model
}

func mainLoop(model *mainController) {
	draw(model)

	eventChan := make(chan termbox.Event, 1)
	go func() {
		for {
			eventChan <- termbox.PollEvent()
		}
	}()

	timeTick := time.Tick(time.Millisecond * 750)

mainloop:
	for {
		select {
		case evt := <-eventChan:
			switch evt.Type {
			case termbox.EventKey:
				model.WillProcessEvent()
				model.state.handleInput(model, evt)
			}
		case <-timeTick:
			tickCount++
		}

		draw(model)
		if model.exitNow {
			break mainloop
		}
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

func drawColorPalette() {
	clearScreen()
	x := 0
	y := 0

	for i := 0x00; i <= 0xFF; i++ {
		numToPrint := "  " + fmt.Sprintf("%02X", i) + "  "
		for _, ch := range numToPrint {
			termbox.SetCell(x, y, ch, termbox.ColorBlack, termbox.Attribute(i))
			x++
		}
		//Fit 8 print outs on a line before creating a new one
		if i%8 == 0 {
			x = 0
			y++
		}
	}

	termbox.Flush()
}

func drawGrid(y int, model *mainController) (endY int) {
	var x int
	grid := model.grid

	fg := termbox.ColorBlack
	bg := termbox.ColorGreen

	var toggleBackgrounds []termbox.Attribute

	//Figure out which possible colors we should paint the grid in.
	//Iterate through toggles backwards, since earlier ones have higher preference
	for i := len(model.toggles) - 1; i >= 0; i-- {
		toggle := model.toggles[i]
		if toggle.Value() {
			toggleBackgrounds = append(toggleBackgrounds, toggle.GridColor)
		}
	}
	//Now, as long as there's one possibiltiy, select a grid color to paint.
	//Otherwise, just leave it as default.
	if len(toggleBackgrounds) > 0 {
		bg = toggleBackgrounds[tickCount%len(toggleBackgrounds)]
	}

	//The column where the grid starts
	gridLeft := 1
	gridTop := 1

	termbox.SetCell(0, y, ' ', fg, bg)
	x++

	for i := 0; i < sudoku.DIM; i++ {

		cellLeft, _, _, _ := grid.Cell(0, i).DiagramExtents()
		cellLeft += gridLeft
		//Pad until we get to the start of this cell area
		for x < cellLeft {
			termbox.SetCell(x, y, '|', fg, bg)
			x++
		}
		for _, ch := range " " + strconv.Itoa(i) + " " {
			termbox.SetCell(x, y, ch, fg, bg)
			x++
		}
	}

	y++

	//Draw diagram down left rail
	x = 0
	tempY := y
	for i := 0; i < sudoku.DIM; i++ {

		_, cellTop, _, _ := grid.Cell(i, 0).DiagramExtents()
		cellTop += gridTop
		//Pad until we get to the start of this cell area
		for tempY < cellTop {
			termbox.SetCell(x, tempY, '-', fg, bg)
			tempY++
		}
		for _, ch := range " " + strconv.Itoa(i) + " " {
			termbox.SetCell(x, tempY, ch, fg, bg)
			tempY++
		}
	}

	fg, bg = bg, fg

	//TODO: I'm pretty sure top/left are reversed
	selectedTop, selectedLeft, selectedHeight, selectedWidth := model.Selected().DiagramExtents()
	//Correct the selected coordinate for the offset of the grid from the top.
	selectedLeft += gridTop
	selectedTop += gridLeft
	for _, line := range strings.Split(grid.Diagram(true), "\n") {
		//Grid starts at 1 cell over from left edge
		x = gridLeft
		//The first number in range will be byte offset, but for some items like the bullet, it's two bytes.
		//But what we care about is that each item is a character.
		for _, ch := range line {

			defaultColor := fg

			numberRune, _ := utf8.DecodeRuneInString(sudoku.DIAGRAM_NUMBER)
			lockedRune, _ := utf8.DecodeRuneInString(sudoku.DIAGRAM_LOCKED)

			if ch == numberRune {
				defaultColor = 0x12
			} else if ch == lockedRune {
				defaultColor = 0x35
			} else if runeIsNum(ch) {
				defaultColor = termbox.ColorWhite | termbox.AttrBold
			}

			backgroundColor := bg
			if x >= selectedTop && x < (selectedTop+selectedHeight) && y >= selectedLeft && y < (selectedLeft+selectedWidth) {
				//We're on the selected cell
				backgroundColor = 0xf0
			}

			termbox.SetCell(x, y, ch, defaultColor, backgroundColor)
			x++
		}
		y++
	}
	//The last loop added one extra to y.
	y--
	return y
}

func drawToggleLine(y int, model *mainController) (newY int) {
	x := 0

	for _, toggle := range model.toggles {
		msg := toggle.OffText
		fg := toggle.GridColor
		bg := termbox.ColorBlack
		if toggle.Value() {
			msg = toggle.OnText
			fg, bg = bg, fg
		}
		for _, ch := range msg {
			termbox.SetCell(x, y, ch, fg, bg)
			x++
		}
	}
	return y
}

func drawStatusLine(y int, model *mainController) (newY int) {
	x := 0
	var fg termbox.Attribute
	underlined := false
	for _, ch := range ">>> " + model.StatusLine() {
		//The ( and ) are non-printing control characters
		if ch == '{' {
			underlined = true
			continue
		} else if ch == '}' {
			underlined = false
			continue
		}
		fg = termbox.ColorBlack
		if underlined {
			fg = fg | termbox.AttrUnderline | termbox.AttrBold
		}

		termbox.SetCell(x, y, ch, fg, termbox.ColorWhite)
		x++
	}

	width, _ := termbox.Size()

	for x < width {
		termbox.SetCell(x, y, ' ', fg, termbox.ColorWhite)
		x++
	}
	return y
}

func drawConsole(y int, model *mainController) (newY int) {

	x := 0

	underlined := false

	splitMessage := strings.Split(model.consoleMessage, "\n")

	for _, line := range splitMessage {

		x = 0
		for _, ch := range line {
			if ch == '{' {
				underlined = true
				continue
			} else if ch == '}' {
				underlined = false
				continue
			}
			fg := termbox.Attribute(0xf6)
			if underlined {
				fg = termbox.Attribute(0xFC) | termbox.AttrBold
			}
			termbox.SetCell(x, y, ch, fg, termbox.ColorBlack)
			x++
		}
		y++
	}
	//y is one too many
	y--
	return y
}

func draw(model *mainController) {

	clearScreen()

	y := 0

	y = drawGrid(y, model)
	y++
	y = drawToggleLine(y, model)
	y++
	y = drawStatusLine(y, model)
	y++
	y = drawConsole(y, model)

	termbox.Flush()
}
