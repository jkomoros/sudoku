##Generating weights

cd to cmd/dokugen-analysis

(If you don't have a good internet connection and have saved db info in mock_data.secret.csv, then add -m to all of these)

go run main.go mock_data.go -a -v -p > relativedifficulties.csv
go run main.go mock_data.go -a -v -w -t -h relativedifficulties.csv > solves.csv

Load up Weka (the JVM version), load solves.csv, switch to Classify tab, select SMOReg (set filterType to No normalization/standardization). Make sure Difficulty is showing in the drop down. (LeastMedSq used to work slightly better)

Copy/paste the output into util/input.txt . It doesn't matter exactly how much you copy paste as long as all of the numbers are there.

Don't commit input.txt or any of the intermediate csvs.

In root of project, run `go generate`. Test the change and then commit.


