package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

const pathToDokugenAnalysis = "../../"
const pathFromDokugenAnalysis = "internal/a-b-tester/"

const pathToWekaTrainer = "../weka-trainer/"
const pathFromWekaTrainer = "../a-b-tester/"

//TODO: amek this resilient to not being run in the package's directory

//TODO: allow the user to specify multiple branches/configs to test, and it reports the best config.

//TODO: allow the user to generate relativedifficulties from scratch, too, which automates the entire pipeline.

type appOptions struct {
	relativeDifficultiesFile string
	solvesFile               string
	analysisFile             string
	branches                 string
	branchesList             []string
	help                     bool
	flagSet                  *flag.FlagSet
}

func (a *appOptions) defineFlags() {
	if a.flagSet == nil {
		return
	}
	a.flagSet.StringVar(&a.branches, "b", "", "Git branch to checkout. Can also be a space delimited list of multiple branches to checkout.")
	a.flagSet.StringVar(&a.relativeDifficultiesFile, "r", "relativedifficulties_SAMPLED.csv", "The file to use as relative difficulties input")
	a.flagSet.StringVar(&a.solvesFile, "s", "solves.csv", "The file to output solves to")
	a.flagSet.StringVar(&a.analysisFile, "a", "analysis.txt", "The file to output analysis to")
	a.flagSet.BoolVar(&a.help, "h", false, "If provided, will print help and exit.")
}

func (a *appOptions) fixUp() {
	a.branchesList = strings.Split(a.branches, " ")
	a.solvesFile = strings.Replace(a.solvesFile, ".csv", "", -1)
	a.analysisFile = strings.Replace(a.analysisFile, ".txt", "", -1)
}

func (a *appOptions) parse(args []string) {
	a.flagSet.Parse(args)
	a.fixUp()
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

	for _, branch := range a.branchesList {

		if branch == "" {
			log.Println("Staying on the current branch.")
		} else {
			log.Println("Switching to branch", branch)
		}

		//a.analysisFile and a.solvesFile have had their extension removed, if they had one.
		effectiveSolvesFile := a.solvesFile + ".csv"
		effectiveAnalysisFile := a.analysisFile + ".txt"

		if branch != "" {

			effectiveSolvesFile = a.solvesFile + "_" + strings.ToUpper(branch) + ".csv"
			effectiveAnalysisFile = a.analysisFile + "_" + strings.ToUpper(branch) + ".txt"
		}

		if !checkoutGitBranch(branch) {
			log.Println("Couldn't switch to branch", branch, " (perhaps you have uncommitted changes?). Quitting.")
			return
		}

		//TODO: support sampling from relative_difficulties via command line option here.

		runSolves(a.relativeDifficultiesFile, effectiveSolvesFile)

		runWeka(effectiveSolvesFile, effectiveAnalysisFile)

		//TODO: understand r2 so we can compare and find the best branch.
	}

	//TODO: print out remembered r2 here, bolding the one that is best.

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
	trainCmd.Stderr = os.Stderr
	output, err := trainCmd.Output()

	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%s", string(output))

}

//gitCurrentBranch returns the current branch that the current repo is in.
func gitCurrentBranch() string {
	branchCmd := exec.Command("git", "branch")

	output, err := branchCmd.Output()

	if err != nil {
		log.Println(err)
		return ""
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "*") {
			//Found it!
			line = strings.Replace(line, "*", "", -1)
			line = strings.TrimSpace(line)
			return line
		}
	}

	return ""
}

func checkoutGitBranch(branch string) bool {

	if branch == "" {
		return true
	}

	checkoutCmd := exec.Command("git", "checkout", branch)
	checkoutCmd.Run()

	if gitCurrentBranch() != branch {
		return false
	}

	return true

}
