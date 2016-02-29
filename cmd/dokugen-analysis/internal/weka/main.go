package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

const outputFile = "solves.arff"

type appOptions struct {
	inFile  string
	help    bool
	flagSet *flag.FlagSet
}

/*


#Instructions:
# * https://weka.wikispaces.com/Primer
# * https://weka.wikispaces.com/How+to+run+WEKA+schemes+from+commandline

#This program assumes that Weka is installed in /Applications

# Convert the provided CSV to arff, capture output, delete the arff.

# java -cp "/Applications/weka-3-6-11-oracle-jvm.app/Contents/Java/weka.jar" weka.core.converters.CSVLoader solves.csv > solves.arff
# java <CLASSPATH> weka.classifiers.functions.SMOreg -C 1.0 -N 2 -I "weka.classifiers.functions.supportVector.RegSMOImproved -L 0.001 -W 1 -P 1.0E-12 -T 0.001 -V" -K "weka.classifiers.functions.supportVector.PolyKernel -C 250007 -E 1.0" -c first -i <ARFF FILE>

#java -cp "/Applications/weka-3-6-11-oracle-jvm.app/Contents/Java/weka.jar" weka.classifiers.functions.SMOreg -C 1.0 -N 2 -I "weka.classifiers.functions.supportVector.RegSMOImproved -L 0.001 -W 1 -P 1.0E-12 -T 0.001 -V" -K "weka.classifiers.functions.supportVector.PolyKernel -C 250007 -E 1.0" -c first -i -t solves.arff

*/

func (a *appOptions) defineFlags() {
	if a.flagSet == nil {
		return
	}
	a.flagSet.StringVar(&a.inFile, "i", "solves.csv", "Which file to read from")
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

	out, err := os.Create(outputFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	cmd.Stdout = out

	err = cmd.Run()

	if err != nil {
		fmt.Println(err)
		return
	}

	//Do the training
	trainCmd := execJavaCommand("weka.classifiers.functions.SMOreg",
		"-C", "1.0", "-N", "2", "-I", `weka.classifiers.functions.supportVector.RegSMOImproved -L 0.001 -W 1 -P 1.0E-12 -T 0.001 -V`,
		"-K", `weka.classifiers.functions.supportVector.PolyKernel -C 250007 -E 1.0`, "-c", "first", "-i", "-t", "solves.arff")

	trainCmd.Stdout = os.Stdout
	trainCmd.Stderr = os.Stderr

	err = trainCmd.Run()

	if err != nil {
		fmt.Println(err)
		return
	}

	//TODO: extract the r2 for comparison.

	//TODO: store the output in a file that we overwrite each time (so the user has it if they want it)

	//Remove the temporary arff file.
	os.Remove(outputFile)

}

func execJavaCommand(input ...string) *exec.Cmd {

	var args []string
	args = append(args, "-cp")
	args = append(args, "/Applications/weka-3-6-11-oracle-jvm.app/Contents/Java/weka.jar")
	args = append(args, input...)

	return exec.Command("java", args...)
}
