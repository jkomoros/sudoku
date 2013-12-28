#Pipe in the CSV you want to use, e.g. `cat out.csv | gnuplot graph.plt`
set datafile separator ","
plot 'out.csv' using 0:3