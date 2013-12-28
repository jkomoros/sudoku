#Expects data to be in data.csv
# Unfortunately, we can't read from the pipe becuase we need to read from the file twice :-(
set datafile separator ","

set logscale y

set title "Puzzle provided difficulty vs markov difficulty"
set xlabel "Puzzles sorted by Markov Difficulty"
set ylabel "Markov Difficulty (log scale)"
set y2label "Provided Difficulty"

set ytics scale 0.01
set y2tics scale 1.0

#0 is a special column that starts from 0 and increments by 1 for each row.
plot 'data.csv' using 0:3title "Markov Difficulties" with lines, \
'data.csv' using 0:2 title "Official Difficulties" axes x1y2 with lines