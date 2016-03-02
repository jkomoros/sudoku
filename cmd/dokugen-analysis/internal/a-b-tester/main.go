/* a-b-tester is a utility program that makes it easy to test and see if a
/* change to the core library is helping us build a better model. Normal usage
/* is to provide it a relativedifficulties.csv file and then it will output
/* r2, but you can also compare multiple configs and have it report the best
/* one. To do that, create different branches with each configuration set.
/* Then run a-b-tester with -b and a space delimited string of branch names to
/* try. a-b-tester will run each in turn, save out analysis and solves files
/* for each, and then report which one has the best r2.*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gosuri/uitable"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

const pathToDokugenAnalysis = "../../"
const pathFromDokugenAnalysis = "internal/a-b-tester/"

const pathToWekaTrainer = "../weka-trainer/"
const pathFromWekaTrainer = "../a-b-tester/"

const rowSeparator = "****************"

//TODO: amek this resilient to not being run in the package's directory

type appOptions struct {
	relativeDifficultiesFile string
	solvesFile               string
	analysisFile             string
	stashMode                bool
	branches                 string
	branchesList             []string
	help                     bool
	flagSet                  *flag.FlagSet
}

func (a *appOptions) defineFlags() {
	if a.flagSet == nil {
		return
	}
	a.flagSet.BoolVar(&a.stashMode, "s", false, "If in stash mode, will do the a-b test between uncommitted and committed changes, automatically figuring out which state we're currently in. Cannot be combined with -b")
	a.flagSet.StringVar(&a.branches, "b", "", "Git branch to checkout. Can also be a space delimited list of multiple branches to checkout.")
	a.flagSet.StringVar(&a.relativeDifficultiesFile, "r", "relativedifficulties_SAMPLED.csv", "The file to use as relative difficulties input")
	a.flagSet.StringVar(&a.solvesFile, "o", "solves.csv", "The file to output solves to")
	a.flagSet.StringVar(&a.analysisFile, "a", "analysis.txt", "The file to output analysis to")
	a.flagSet.BoolVar(&a.help, "h", false, "If provided, will print help and exit.")
}

func (a *appOptions) fixUp() error {
	if a.branches != "" && a.stashMode {
		return errors.New("-b and -s cannot both be passed")
	}
	a.branchesList = strings.Split(a.branches, " ")
	a.solvesFile = strings.Replace(a.solvesFile, ".csv", "", -1)
	a.analysisFile = strings.Replace(a.analysisFile, ".txt", "", -1)
	return nil
}

func (a *appOptions) parse(args []string) error {
	a.flagSet.Parse(args)
	return a.fixUp()
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
	if err := a.parse(os.Args[1:]); err != nil {
		log.Println("Invalid options provided:", err.Error())
		return
	}

	results := make(map[string]float64)

	startingBranch := gitCurrentBranch()

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

		runSolves(a.relativeDifficultiesFile, effectiveSolvesFile)

		branchKey := branch

		if branchKey == "" {
			branchKey = "<default>"
		}

		results[branchKey] = runWeka(effectiveSolvesFile, effectiveAnalysisFile)
	}

	if len(results) > 1 {
		//We only need to go to the trouble of painting the table if more than
		//one branch was run
		printR2Table(results)
	}

	if gitCurrentBranch() != startingBranch {
		checkoutGitBranch(startingBranch)
	}
}

func printR2Table(results map[string]float64) {
	bestR2 := 0.0
	bestR2Branch := ""

	for key, val := range results {
		if val > bestR2 {
			bestR2 = val
			bestR2Branch = key
		}
	}

	fmt.Println(rowSeparator)
	fmt.Println("Results:")
	fmt.Println(rowSeparator)

	table := uitable.New()

	table.AddRow("Best?", "Branch", "R2")

	for key, val := range results {
		isBest := " "
		if key == bestR2Branch {
			isBest = "*"
		}
		table.AddRow(isBest, key, val)
	}

	fmt.Println(table.String())
	fmt.Println(rowSeparator)
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

func runWeka(solvesFile string, analysisFile string) float64 {

	os.Chdir(pathToWekaTrainer)

	defer func() {
		os.Chdir(pathFromWekaTrainer)
	}()

	//Build the weka-trainer executable to make sure we get the freshest version of the sudoku pacakge.
	cmd := exec.Command("go", "build")
	err := cmd.Run()

	if err != nil {
		log.Println(err)
		return 0.0
	}

	trainCmd := exec.Command("./weka-trainer", "-i", path.Join(pathFromWekaTrainer, solvesFile), "-o", path.Join(pathFromWekaTrainer, analysisFile))
	trainCmd.Stderr = os.Stderr
	output, err := trainCmd.Output()

	if err != nil {
		log.Println(err)
		return 0.0
	}

	fmt.Printf("%s", string(output))

	return extractR2(string(output))

}

//extractR2 extracts R2 out of the string formatted like "R2 = <float>"
func extractR2(input string) float64 {

	input = strings.TrimPrefix(input, "R2 = ")
	input = strings.TrimSpace(input)

	result, _ := strconv.ParseFloat(input, 64)

	return result

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
