/* analysis-pipeline is a utility program that makes it easy to test and see if a
/* change to the core library is helping us build a better model. Normal usage
/* is to provide it a relativedifficulties.csv file and then it will output
/* r2, but you can also compare multiple configs and have it report the best
/* one. To do that, create different branches with each configuration set.
/* Then run analysis-pipeline with -b and a space delimited string of branch names to
/* try. analysis-pipeline will run each in turn, save out analysis and solves files
/* for each, and then report which one has the best r2.*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gosuri/uitable"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const pathToDokugenAnalysis = "../../"
const pathFromDokugenAnalysis = "internal/analysis-pipeline/"

const pathToWekaTrainer = "../weka-trainer/"
const pathFromWekaTrainer = "../analysis-pipeline/"

const rowSeparator = "****************"

const uncommittedChangesBranchName = "STASHED"
const committedChangesBranchName = "COMMITTED"

//Temp files that should be deleted when the program exits.
var filesToDelete []string

var initialPath string

//TODO: amek this resilient to not being run in the package's directory

type appOptions struct {
	//The actual relative difficulties file to use
	relativeDifficultiesFile string
	//TODO: currently we quit if this is provided with g, but in some cases you want it to export AND keep going.
	outputRelativeDifficultiesFile string
	//TODO: deleteRelativeDifficultiesFile is named incorrectly because we use it more specifically than that in main.
	deleteRelativeDifficultiesFile bool
	solvesFile                     string
	analysisFile                   string
	sampleRate                     int
	numRuns                        int
	stashMode                      bool
	startingWithUncommittedChanges bool
	branches                       string
	branchesList                   []string
	help                           bool
	generateRelativeDifficulties   bool
	//TODO: this is probably named wrong, since currently it's only used to exit if -g passed.
	exitEarly bool
	flagSet   *flag.FlagSet
}

func (a *appOptions) defineFlags() {
	if a.flagSet == nil {
		return
	}
	a.flagSet.IntVar(&a.sampleRate, "sample-rate", 0, "An optional sample rate of relative difficulties. Will use 1/n lines in calculation. 0 to use all.")
	a.flagSet.BoolVar(&a.stashMode, "s", false, "If in stash mode, will do the a-b test between uncommitted and committed changes, automatically figuring out which state we're currently in. Cannot be combined with -b")
	a.flagSet.StringVar(&a.branches, "b", "", "Git branch to checkout. Can also be a space delimited list of multiple branches to checkout.")
	a.flagSet.StringVar(&a.relativeDifficultiesFile, "r", "relativedifficulties.csv", "The file to use as relative difficulties input.")
	//TODO: this is a terrible name for this flag. Can we reuse -o? ... no, because then it's not a clear signal to exit if provided.
	a.flagSet.StringVar(&a.outputRelativeDifficultiesFile, "rd-out", "", "If -g is also provided and this path does not point to an existing file, will save out the generated relative difficulties to that location.")
	a.flagSet.StringVar(&a.solvesFile, "o", "solves.csv", "The file to output solves to")
	a.flagSet.StringVar(&a.analysisFile, "a", "analysis.txt", "The file to output analysis to")
	a.flagSet.IntVar(&a.numRuns, "n", 1, "The number of runs of each config to do and then average together")
	a.flagSet.BoolVar(&a.generateRelativeDifficulties, "g", false, "If true, then will generate relative difficulties file.")
	a.flagSet.BoolVar(&a.help, "h", false, "If provided, will print help and exit.")
	a.flagSet.BoolVar(&a.exitEarly, "exit", false, "If provided with -g and rd-out, will generate relative difficulty file to rd-out and exit.")
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func randomFileName(prefix, suffix string) string {
	//Look for a file name that doesn't already exist. At each step make the random part bigger,
	//but at some point we have to give up.
	for i := 0; i < 1000; i++ {
		size := i
		if size > 5 {
			size = 5
		}

		str := randomString(size)

		if len(str) > 0 {
			str = "_" + str
		}

		candidate := prefix + str + suffix
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			//found one that doesn't exist!
			return candidate
		}
	}
	panic("Couldn't find a non used filename")
	return ""
}

func randomString(length int) string {
	var letters = []rune("0123456789")
	b := make([]rune, length)
	for i := 0; i < length; i++ {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (a *appOptions) fixUp() error {
	if a.branches != "" && a.stashMode {
		return errors.New("-b and -s cannot both be passed")
	}
	if a.stashMode {
		a.startingWithUncommittedChanges = gitUncommittedChanges()
		if a.startingWithUncommittedChanges {
			a.branchesList = []string{
				uncommittedChangesBranchName,
				committedChangesBranchName,
			}
		} else {
			a.branchesList = []string{
				committedChangesBranchName,
				uncommittedChangesBranchName,
			}
		}
	} else {
		a.branchesList = strings.Split(a.branches, " ")
	}

	if a.numRuns < 1 {
		a.numRuns = 1
	}

	if a.generateRelativeDifficulties {
		if a.outputRelativeDifficultiesFile == "" {
			//They didn't provide a file, so we'll store the relative difficulties in a temporary file.
			a.outputRelativeDifficultiesFile = randomFileName("relative_difficulties_TEMP", ".csv")

			//We want to delete this one when we're done
			a.deleteRelativeDifficultiesFile = true

		} else {
			//We'll be outputting the generated relative difficulties to this location. Make sure it's empty

			if _, err := os.Stat(a.outputRelativeDifficultiesFile); !os.IsNotExist(err) {
				return errors.New("Passed -g and -r pointing to a non-empty file.")
			}
		}
		if a.exitEarly && a.outputRelativeDifficultiesFile == "" {
			return errors.New("Exit passed without both g and rd-out")
		}
	} else {
		if a.outputRelativeDifficultiesFile != "" {
			return errors.New("rd-out passed without g")
		}
		if a.exitEarly {
			return errors.New("-exit passed without g")
		}
	}

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

func cleanUpTempFiles() {
	for _, filename := range filesToDelete {
		err := os.Remove(path.Join(initialPath, filename))
		if err != nil {
			log.Println("Couldn't delete", filename, err)
		}
	}
}

func buildExecutables() bool {
	if !buildWeka() {
		return false
	}

	if !buildDokugenAnalysis() {
		return false
	}

	return true
}

func main() {

	defer cleanUpTempFiles()

	//Keep track of the working directory so cleanupTempFiles always has it.
	initialPath, _ = os.Getwd()

	//Make sure that even if we get exited early we still clean up.
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		cleanUpTempFiles()
		os.Exit(1)
	}()

	//build the executables

	//TODO: if -no-build is passed, skip this
	if !buildExecutables() {
		return
	}

	//parse the flags

	a := newAppOptions(flag.CommandLine)
	if err := a.parse(os.Args[1:]); err != nil {
		log.Println("Invalid options provided:", err.Error())
		return
	}

	if a.help {
		a.flagSet.PrintDefaults()
		return
	}

	//TODO: most of this method should be factored into a separate func, so
	//main is just configuring hte options and passing them in.

	if a.generateRelativeDifficulties {
		log.Println("Generating relative difficulties.")

		//If we're just using a temp file we should be sure to delete when done.
		//We add this now in case the user exits the program while we're generating the difficulties.
		if a.deleteRelativeDifficultiesFile {
			filesToDelete = append(filesToDelete, a.outputRelativeDifficultiesFile)
		}

		//a.fixUp put a valid filename in a.outputRelativeDifficultiesFile
		generateRelativeDifficulties(a.outputRelativeDifficultiesFile)

		if !a.deleteRelativeDifficultiesFile && a.exitEarly {
			//We're done, all we wanted to do was generate the file and quit.
			return
		}

		//Make sure we're wired up to use the file we're outputting it to.
		a.relativeDifficultiesFile = a.outputRelativeDifficultiesFile
	}

	if _, err := os.Stat(a.relativeDifficultiesFile); os.IsNotExist(err) {
		log.Println("The specified relative difficulties file does not exist:", a.relativeDifficultiesFile)
		return
	}

	results := make(map[string]float64)

	startingBranch := gitCurrentBranch()

	branchSwitchMessage := "Switching to branch"

	relativeDifficultiesFile := a.relativeDifficultiesFile

	if a.sampleRate > 0 {
		relativeDifficultiesFile = strings.Replace(a.relativeDifficultiesFile, ".csv", "", -1)
		relativeDifficultiesFile += "_SAMPLED_" + strconv.Itoa(a.sampleRate) + ".csv"
		if !sampledRelativeDifficulties(a.relativeDifficultiesFile, relativeDifficultiesFile, a.sampleRate) {
			log.Println("Couldn't create sampled relative difficulties file")
			return
		}
		filesToDelete = append(filesToDelete, relativeDifficultiesFile)
	}

	//TODO: this is off by one
	log.Println(strconv.Itoa(numLinesInFile(relativeDifficultiesFile)), "lines in", relativeDifficultiesFile)

	if a.stashMode {
		branchSwitchMessage = "Calculating on"
	}

	for i, branch := range a.branchesList {

		if branch == "" {
			log.Println("Staying on the current branch.")
		} else {
			log.Println(branchSwitchMessage, branch)
		}

		//Get the repo in the right state for this run.
		if a.stashMode {
			// if i == 0
			switch i {
			case 0:
				//do nothing, we already ahve the right changes to start with
			case 1:
				//If we have uncommitted changes right now, stash them. Otherwise, stash pop.
				if !gitStash(a.startingWithUncommittedChanges) {
					log.Println("We couldn't stash/stash-pop.")
					return
				}
			default:
				//This should never happen
				//Note: panicing here will mean we don't do any clean up.
				panic("Got more than 2 'branches' in stash mode")
			}
		} else {
			if !checkoutGitBranch(branch) {
				log.Println("Couldn't switch to branch", branch, " (perhaps you have uncommitted changes?). Quitting.")
				return
			}
		}

		for i := 0; i < a.numRuns; i++ {

			//The run number reported to humans will be one indexed
			oneIndexedRun := strconv.Itoa(i + 1)

			if a.numRuns > 1 {
				log.Println("Starting run", oneIndexedRun, "of", strconv.Itoa(a.numRuns))
			}

			//a.analysisFile and a.solvesFile have had their extension removed, if they had one.
			effectiveSolvesFile := a.solvesFile
			effectiveAnalysisFile := a.analysisFile

			if branch != "" {
				effectiveSolvesFile += "_" + strings.ToUpper(branch)
				effectiveAnalysisFile += "_" + strings.ToUpper(branch)
			}

			if a.numRuns > 1 {
				effectiveSolvesFile += "_" + oneIndexedRun
				effectiveAnalysisFile += "_" + oneIndexedRun
			}

			effectiveSolvesFile += ".csv"
			effectiveAnalysisFile += ".txt"

			runSolves(relativeDifficultiesFile, effectiveSolvesFile)

			branchKey := branch

			if branchKey == "" {
				branchKey = "<default>"
			}

			log.Println("Running Weka on solves...")

			///Accumulate the R2 for each run; we'll divide by numRuns after the loop.
			results[branchKey] += runWeka(effectiveSolvesFile, effectiveAnalysisFile)
		}
	}

	//Take the average of each r2
	for key, val := range results {
		results[key] = val / float64(a.numRuns)
	}

	if len(results) > 1 || a.numRuns > 1 {
		//We only need to go to the trouble of painting the table if more than
		//one branch was run
		printR2Table(results)
	}

	//Put the repo back in the state it was when we found it.
	if a.stashMode {
		//Reverse the gitStash operation to put it back

		if a.startingWithUncommittedChanges {
			log.Println("Unstashing changes to put repo back in starting state")
		} else {
			log.Println("Stashing changes to put repo back in starting state")
		}

		if !gitStash(!a.startingWithUncommittedChanges) {
			log.Println("We couldn't unstash/unpop to put the repo back in the same state.")
		}
	} else {
		//If we aren't in the branch we started in, switch back to that branch
		if gitCurrentBranch() != startingBranch {
			log.Println("Checking out", startingBranch, "to put repo back in the starting state.")
			checkoutGitBranch(startingBranch)
		}
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

func numLinesInFile(filename string) int {
	contents, err := ioutil.ReadFile(filename)

	if err != nil {
		return 0
	}

	return len(strings.Split(string(contents), "\n"))
}

func buildDokugenAnalysis() bool {
	os.Chdir(pathToDokugenAnalysis)

	defer func() {
		os.Chdir(pathFromDokugenAnalysis)
	}()

	//Build the dokugen-analysis executable to make sure we get the freshest version of the sudoku pacakge.

	cmd := exec.Command("go", "build")
	err := cmd.Run()

	if err != nil {
		log.Println("Couldn't build dokugen-analysis", err)
		return false
	}
	return true
}

func runSolves(difficultiesFile, solvesOutputFile string) {

	os.Chdir(pathToDokugenAnalysis)

	defer func() {
		os.Chdir(pathFromDokugenAnalysis)
	}()

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

func generateRelativeDifficulties(outputFile string) {
	os.Chdir(pathToDokugenAnalysis)

	defer func() {
		os.Chdir(pathFromDokugenAnalysis)
	}()

	outFile, err := os.Create(path.Join(pathFromDokugenAnalysis, outputFile))

	if err != nil {
		log.Println(err)
		return
	}

	analysisCmd := exec.Command("./dokugen-analysis", "-a", "-v", "-p")
	analysisCmd.Stdout = outFile
	analysisCmd.Stderr = os.Stderr
	err = analysisCmd.Run()

	if err != nil {
		log.Println(err)
	}
}

func buildWeka() bool {
	os.Chdir(pathToWekaTrainer)

	defer func() {
		os.Chdir(pathFromWekaTrainer)
	}()

	//Build the weka-trainer executable to make sure we get the freshest version of the sudoku pacakge.

	//TODO: we should only have to do this once, not every time this method is called
	cmd := exec.Command("go", "build")
	err := cmd.Run()

	if err != nil {
		log.Println("Couldn't build weka:", err)
		return false
	}

	return true
}

func runWeka(solvesFile string, analysisFile string) float64 {

	os.Chdir(pathToWekaTrainer)

	defer func() {
		os.Chdir(pathFromWekaTrainer)
	}()

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

	//Note: we don't use wekaparser.ParseR2 here because we're getting a much
	//simpler output from weka-trainer.

	input = strings.TrimPrefix(input, "R2 = ")
	input = strings.TrimSpace(input)

	result, _ := strconv.ParseFloat(input, 64)

	return result

}

func sampledRelativeDifficulties(inputFile, sampledFile string, sampleRate int) bool {

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		log.Println(inputFile, "does not exist")
		return false
	}

	if sampleRate < 1 {
		sampleRate = 1
	}

	awkPattern := `NR % ` + strconv.Itoa(sampleRate) + ` == 0`

	awkCmd := exec.Command("awk", awkPattern, inputFile)

	out, err := os.Create(sampledFile)

	if err != nil {
		log.Println(err)
		return false
	}

	awkCmd.Stdout = out
	awkCmd.Stderr = os.Stderr

	err = awkCmd.Run()

	if err != nil {
		log.Println("Awk error", err)
		return false
	}

	return true
}

//gitStash will use git stash if true, git stash pop if false.
func gitStash(stashChanges bool) bool {
	var stashCmd *exec.Cmd

	if stashChanges {
		stashCmd = exec.Command("git", "stash")
		if !gitUncommittedChanges() {
			//That's weird, there aren't any changes to stash
			log.Println("Can't stash: no uncommitted changes!")
			return false
		}
	} else {
		stashCmd = exec.Command("git", "stash", "pop")
		if gitUncommittedChanges() {
			//That's weird, there are uncommitted changes that this would overwrite.
			log.Println("Can't stash pop: uncommitted changes that would be overwritten")
			return false
		}
	}

	err := stashCmd.Run()

	if err != nil {
		log.Println(err)
		return false
	}

	//Verify it worked
	if stashChanges {
		//Stashing apaprently didn't work
		if gitUncommittedChanges() {
			log.Println("Stashing didn't work; there are still uncommitted changes")
			return false
		}
	} else {
		//Weird, stash popping didn't do anything.
		if !gitUncommittedChanges() {
			log.Println("Stash popping didn't work; there are no uncommitted changes that resulted.	")
			return false
		}
	}

	return true
}

//Returns true if there are currently uncommitted changes
func gitUncommittedChanges() bool {

	statusCmd := exec.Command("git", "status", "-s")

	output, err := statusCmd.Output()

	if err != nil {
		log.Println(err)
		return false
	}

	//In git status -s(hort), each line starts with two characters. ?? is hte
	//only prefix that we should ignore, since it means untracked files.

	for _, line := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, "??") {
			//Found a non-committed change
			return true
		}
	}

	return false

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