package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

const temporaryArff = "solves.arff"

const r2RegularExpression = `=== Cross-validation ===\n\nCorrelation coefficient\s*(\d\.\d{1,10})`

var wekaJar string

type appOptions struct {
	inFile  string
	outFile string
	help    bool
	flagSet *flag.FlagSet
}

func init() {

	//Check for various installed versions of Weka

	//TODO: make this WAY more resilient to different versions
	possibleJarLocations := []string{
		"/Applications/weka-3-6-11-oracle-jvm.app/Contents/Java/weka.jar",
		"/Applications/weka-3-6-12-oracle-jvm.app/Contents/Java/weka.jar",
	}

	for _, path := range possibleJarLocations {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			//Found it!
			wekaJar = path
			continue
		}
	}

	if wekaJar == "" {
		log.Fatalln("Could not find Weka")
	}

}

func (a *appOptions) defineFlags() {
	if a.flagSet == nil {
		return
	}
	a.flagSet.StringVar(&a.inFile, "i", "solves.csv", "Which file to read from")
	a.flagSet.StringVar(&a.outFile, "o", "analysis.txt", "Which file to output analysis to")
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

	options := newAppOptions(flag.CommandLine)
	options.parse(os.Args[1:])

	if options.help {
		options.flagSet.PrintDefaults()
		return
	}

	//TODO: allow configuring just a relativedifficulties file and run the whole pipeline

	//First, convert the file to arff.

	cmd := execJavaCommand("weka.core.converters.CSVLoader", options.inFile)

	out, err := os.Create(temporaryArff)

	if err != nil {
		log.Println(err)
		return
	}

	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		log.Println(err)
		return
	}

	//Do the training
	trainCmd := execJavaCommand("weka.classifiers.functions.SMOreg",
		"-C", "1.0", "-N", "2", "-I", `weka.classifiers.functions.supportVector.RegSMOImproved -L 0.001 -W 1 -P 1.0E-12 -T 0.001 -V`,
		"-K", `weka.classifiers.functions.supportVector.PolyKernel -C 250007 -E 1.0`, "-c", "first", "-i", "-t", "solves.arff")

	trainCmd.Stderr = os.Stderr

	output, err := trainCmd.Output()

	if err != nil {
		log.Println(err)
		return
	}

	r2 := extractR2(string(output))

	fmt.Println("R2 =", r2)

	ioutil.WriteFile(options.outFile, output, 0644)

	//Remove the temporary arff file.
	os.Remove(temporaryArff)

}

func execJavaCommand(input ...string) *exec.Cmd {

	var args []string
	args = append(args, "-cp")
	args = append(args, wekaJar)
	args = append(args, input...)

	return exec.Command("java", args...)
}

func extractR2(input string) float64 {
	re := regexp.MustCompile(r2RegularExpression)
	result := re.FindStringSubmatch(input)

	if len(result) != 2 {
		return 0.0
	}

	//Match 0 is the entire expression, so the float is in match 1

	float, _ := strconv.ParseFloat(result[1], 64)
	return float
}
