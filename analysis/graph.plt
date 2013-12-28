#Expects data to be in data.csv
# Unfortunately, we can't read from the pipe becuase we need to read from the file twice :-(
set datafile separator ","

#0 is a special column that starts from 0 and increments by 1 for each row.
plot 'data.csv' using 0:3 title "Markov Difficulties" with boxes, \
'data.csv' using 0:2 title "QQWing Difficulties" axes x1y2 with lines