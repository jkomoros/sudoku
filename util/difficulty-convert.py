import logging
import string

input_file_name = "input.txt"


#input.txt should be a copy/pasted output from the Weka training.
#yes, copy/pasted; Weka makes it really hard to export the weights in any other way.

def convertDifficulties():
	f = open(input_file_name)

	result = {}

	if not f:
		logging.error("Couldn't find input.txt")
		return

	for line in f:
		line = string.lstrip(line)
		if len(line) == 0:
			continue
		if line[0] != "+" and line[0] != "-":
			#All lines we're looking for start with either + or -
			continue
		print line



convertDifficulties()