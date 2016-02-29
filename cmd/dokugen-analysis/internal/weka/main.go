package main

import (
	"fmt"
	"os"
	"os/exec"
)

const outputFile = "solves.arff"

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

func main() {

	//TODO: allow configuring a different in file.

	//First, convert the file to arff.

	cmd := execJavaCommand("weka.core.converters.CSVLoader", "solves.csv")

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
