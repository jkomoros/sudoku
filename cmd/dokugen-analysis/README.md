##Generating weights

cd to cmd/dokugen-analysis

(If you don't have a good internet connection and have saved db info in mock_data.secret.csv, then add -m to all of these)

go run main.go mock_data.go -a -v -p > relativedifficulties.csv
go run main.go mock_data.go -a -v -w -t -h relativedifficulties.csv > solves.csv

Load up Weka (the JVM version), load solves.csv, switch to Classify tab, select LeastMedSq (leave defaults -S 4), make sure (num) Difficulty is showing in the drop down.

Recently SMOReg with the filterType set to No normalization/standardization has been giving higher correlations.


