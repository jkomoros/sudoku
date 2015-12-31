/*
i-sudoku is an interactive command-line sudoku tool
*/

package main

import (
	"github.com/jkomoros/sudoku"
	"github.com/nsf/termbox-go"
	"log"
	"strings"
)

type mainModel struct {
	grid *sudoku.Grid
}

func main() {
	if err := termbox.Init(); err != nil {
		log.Fatal("Termbox initialization failed:", err)
	}
	defer termbox.Close()

	model := &mainModel{
		sudoku.NewGrid(),
	}

	model.grid.Fill()

	draw(model)

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			}
		}
		draw(model)
	}
}

func draw(model *mainModel) {
	drawGrid(model.grid)
	termbox.Flush()
}

func drawGrid(grid *sudoku.Grid) {
	for y, line := range strings.Split(grid.Diagram(), "\n") {
		for x, ch := range line {
			termbox.SetCell(x, y, ch, termbox.ColorGreen, termbox.ColorDefault)
		}
	}
}
