##Generating weights

cd to cmd/dokugen-analysis

(If you don't have a good internet connection and have saved db info in mock_data.secret.csv, then add -m to all of these)

Note: now the preferred way to do this (once you have relativedifficulties generated) is just to use a-b-tester (see below).

go run main.go mock_data.go -a -v -p > relativedifficulties.csv
go run main.go mock_data.go -a -v -w -t -h relativedifficulties.csv > solves.csv

Load up Weka (the JVM version), load solves.csv, switch to Classify tab, select SMOReg (set filterType to No normalization/standardization). Make sure Difficulty is showing in the drop down. (LeastMedSq used to work slightly better)

Copy/paste the output into util/input.txt . It doesn't matter exactly how much you copy paste as long as all of the numbers are there.

Don't commit input.txt or any of the intermediate csvs.

In root of project, run `go generate`. Test the change and then commit.


## Scikit version

Install pip: `sudo easy_install pip`
Install scikit: `sudo pip install -U numpy scipy scikit-learn`

## Creating a sampled relative_difficulties

To create a sampled relativedifficulties, build a full relativedifficulties.csv. Then run:

awk 'NR == 1 || NR % 10 == 0' relativedifficulties.csv > relativedifficulties_SAMPLED.csv

## Quick A/B pattern

To quickly tell if a change to the library helped r2, create a relativedifficulties.csv in internal/a-b-tester with the command above.

Then configure the changes and either leave them uncommitted or stash them. Then run:

go build && ./a-b-tester -r relativedifficulties.csv -s

This will run the tester, then either stash or stash pop, then run again, then leave the repo in the state in started in.

If there are multiple configs you want to test, store each on in a named branched, then run:

go build && ./a-b-tester -r relativedifficulties.csv -b "branch-1 branch-2 branch-3"

and it will try all branches and output which one is better.

