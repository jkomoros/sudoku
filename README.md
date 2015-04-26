#sudoku
sudoku is a package written in go for generating and solving sudoku puzzles. Notably, it is able to solve a puzzle like a human would, and aspires to provide high-quality difficulty ratings for puzzles based on a model trained on hundreds of thousands of solves by real users.

This is my first non-trivial project written in Go. It is very much a **work in progress**, and may be rough/broken/over-engineered/non-idiomatic/slow in places. Pull requests welcome!

You can find documentation at http://godoc.org/github.com/jkomoros/sudoku

##Included commands

###dokugen

Dokugen is a simple command line utility exposing the main functionality of the sudoku package, including solving, generating, and difficutly-rating puzzles.

Run `dokugen -h` to see information on command line options.

###dokugen-analysis

dokugen-analysis is an analysis utility used to analyze a large collection of real-world user solves in order to generate a model to rate the difficulty of provided puzzles.

komoroske.com/sudoku has been running for over 8 years and during that time has logged information on over 800,000 solves by real users. dokugen-analysis uses that information to determine how difficult puzzles are, as judged by the solve time for real users. The analysis includes using markov chains to do rank aggregation to come up with a cross-user rating of difficulty.

This analysis is then used to train a simple multiple linear regression model that provides difficulty ratings for puzzles.

The tool is not particularly useful unless you have your own large database of solve times.
