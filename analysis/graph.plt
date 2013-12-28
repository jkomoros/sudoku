#Pipe in the CSV you want to use, e.g. `cat out.csv | gnuplot graph.plt`
set datafile separator ","

#0 is a special column that starts from 0 and increments by 1 for each row.
plot 'out.csv' using 0:3 title "Markov Difficulties" with boxes