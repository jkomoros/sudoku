##Generating weights

cd to cmd/dokugen-analysis

(If you don't have a good internet connection and have saved db info in mock_data.secret.csv, then add -m to all of these)

go run main.go mock_data.go -a -v -p > relativedifficulties.csv
go run main.go mock_data.go -a -v -w -t -h relativedifficulties.csv > solves.csv

Load up Weka (the JVM version), load solves.csv, switch to Classify tab, select SMOReg (set filterType to No normalization/standardization). Make sure Difficulty is showing in the drop down. (LeastMedSq used to work slightly better)

Copy/paste the output into util/input.txt . It doesn't matter exactly how much you copy paste as long as all of the numbers are there.

Don't commit input.txt or any of the intermediate csvs.

In root of project, run `go generate`. Test the change and then commit.


## Scikit version

Install pip: `sudo easy_install pip`
Install scikit: `sudo pip install -U numpy scipy scikit-learn`

## Quick A/B pattern

To quickkly tell if a change to the library helped r2, from dokugen-analysis folder, run:

go run main.go mock_data.go -a -v -w -t -h -no-cache relativedifficulties_SAMPLED.csv > internal/weka-trainer/solves_BEFORE.csv

Then make the change, then run:

go run main.go mock_data.go -a -v -w -t -h -no-cache relativedifficulties_SAMPLED.csv > internal/weka-trainer/solves_AFTER.csv

Then, from internal/weka-trainer, run: 

go build && ./weka-trainer -i solves_BEFORE.csv -o analysis_BEFORE.txt
go build && ./weka-trainer -i solves_AFTER.csv -o analysis_AFTER.txt

And see which one prints a higher r2 to the screen.

*TODO* Implement a script that does this automatically