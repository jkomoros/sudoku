/*
i-sudoku is an interactive command-line sudoku tool
*/

package main

import (
	"github.com/nsf/termbox-go"
	"log"
)

func main() {
	if err := termbox.Init(); err != nil {
		log.Fatal("Termbox initialization failed:", err)
	}
	defer termbox.Close()

	i := 0

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			case termbox.KeySpace:
				termbox.SetCell(i, 0, '*', termbox.ColorBlue, termbox.ColorGreen)
				i++
			}
		}
		termbox.Flush()
	}
}
