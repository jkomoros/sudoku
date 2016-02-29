package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path"
)

const pathToDokugenAnalysis = "../../"
const pathFromDokugenAnalysis = "internal/a-b-tester/"

//TODO: amek this resilient to not being run in the package's directory

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

	runSolves("relativedifficulties_SAMPLED.csv", "solves_SAMPLED.csv")

	//TODO: should we be cleaning up the files we output (perhaps only if option provided?0)
}

func runSolves(difficultiesFile, solvesOutputFile string) {

	os.Chdir(pathToDokugenAnalysis)

	//Build the dokugen-analysis executable to make sure we get the freshest version of the sudoku pacakge.
	cmd := exec.Command("go", "build")
	err := cmd.Run()

	if err != nil {
		log.Println(err)
		return
	}

	outFile, err := os.Create(path.Join(pathFromDokugenAnalysis, solvesOutputFile))

	if err != nil {
		log.Println(err)
		return
	}

	analysisCmd := exec.Command("./dokugen-analysis", "-a", "-v", "-w", "-t", "-h", "-no-cache", path.Join(pathFromDokugenAnalysis, difficultiesFile))
	analysisCmd.Stdout = outFile
	analysisCmd.Stderr = os.Stderr
	err = analysisCmd.Run()

	if err != nil {
		log.Println(err)
	}

}
