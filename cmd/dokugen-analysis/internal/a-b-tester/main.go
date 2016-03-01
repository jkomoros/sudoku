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

const pathToWekaTrainer = "../weka-trainer/"
const pathFromWekaTrainer = "../a-b-tester/"

//TODO: amek this resilient to not being run in the package's directory

//TODO: allow the user to specify git branches to switch between for the
//before and after to do automated comparisons. (Verify that the checkout
//works--that there's no unstashed changes.)

//TODO: allow the user to specify multiple branches/configs to test, and it reports the best config.

type appOptions struct {

	//TODO: allow configuring a suffix, e.g. "BEFORE", "AFTER" that is appended to all output files
	//TODO: allow -a and -b to automatically set suffix to BEFORE/AFTER
	relativeDifficultiesFile string
	solvesFile               string
	analysisFile             string
	branch                   string
	help                     bool
	flagSet                  *flag.FlagSet
}

func (a *appOptions) defineFlags() {
	if a.flagSet == nil {
		return
	}
	a.flagSet.StringVar(&a.branch, "b", "", "Git branch to checkout")
	a.flagSet.StringVar(&a.relativeDifficultiesFile, "r", "relativedifficulties_SAMPLED.csv", "The file to use as relative difficulties input")
	a.flagSet.StringVar(&a.solvesFile, "s", "solves.csv", "The file to output solves to")
	a.flagSet.StringVar(&a.analysisFile, "a", "analysis.txt", "The file to output analysis to")
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

	checkoutGitBranch(a.branch)

	//TODO: support sampling from relative_difficulties via command line option here.

	runSolves(a.relativeDifficultiesFile, a.solvesFile)

	runWeka(a.solvesFile, a.analysisFile)

	//TODO: should we be cleaning up the files we output (perhaps only if option provided?0)
}

func runSolves(difficultiesFile, solvesOutputFile string) {

	os.Chdir(pathToDokugenAnalysis)

	defer func() {
		os.Chdir(pathFromDokugenAnalysis)
	}()

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

func runWeka(solvesFile string, analysisFile string) {

	os.Chdir(pathToWekaTrainer)

	defer func() {
		os.Chdir(pathFromWekaTrainer)
	}()

	//Build the weka-trainer executable to make sure we get the freshest version of the sudoku pacakge.
	cmd := exec.Command("go", "build")
	err := cmd.Run()

	if err != nil {
		log.Println(err)
		return
	}

	trainCmd := exec.Command("./weka-trainer", "-i", path.Join(pathFromWekaTrainer, solvesFile), "-o", path.Join(pathFromWekaTrainer, analysisFile))
	trainCmd.Stdout = os.Stdout
	trainCmd.Stderr = os.Stderr
	err = trainCmd.Run()

	//TODO: parse the r2 here

	if err != nil {
		log.Println(err)
	}

}

func checkoutGitBranch(branch string) bool {

	if branch == "" {
		return true
	}

	checkoutCmd := exec.Command("git", "checkout", branch)

	combinedOutput, err := checkoutCmd.CombinedOutput()

	if err != nil {
		log.Println(err)
		return false
	}

	if len(combinedOutput) > 0 {
		log.Println(string(combinedOutput))
		return false
	}

	return true

}
