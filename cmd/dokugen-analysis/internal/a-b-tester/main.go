package main

import (
	"flag"
	"log"
	"os"
)

type appOptions struct {
	message string
	help    bool
	flagSet *flag.FlagSet
}

func (a *appOptions) defineFlags() {
	if a.flagSet == nil {
		return
	}
	a.flagSet.StringVar(&a.message, "m", "Hello, world!", "The message to print to the screen")
	a.flagSet.BoolVar(&a.help, "h", false, "If provided, will print help and exit.")
}

func (a *appOptions) parse(args []string) {
	a.flagSet.Parse(args)
}

func newAppOptions(flagSet *flag.FlagSet) *appOptions {
	a := &appOptions{
		flagSet: flagSet,
	}
	a.defineFlags()
	return a
}

func main() {
	a := newAppOptions(flag.CommandLine)
	a.parse(os.Args[1:])

	log.Println(a.message)
}
