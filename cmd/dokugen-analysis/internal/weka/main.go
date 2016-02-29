package main

import (
	"fmt"
	"os"
	"os/exec"
)

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

	//TODO: factor out the class path

	//First, convert the file to arff.

	cmd := exec.Command("java",
		"-cp", "/Applications/weka-3-6-11-oracle-jvm.app/Contents/Java/weka.jar", "weka.core.converters.CSVLoader", "solves.csv")

	out, err := os.Create("solves.arff")

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

	//TODO: delete the arff files.

}
