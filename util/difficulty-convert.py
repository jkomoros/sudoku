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
		negative = line[0] == "-"
		line = string.lstrip(line[1:])
		parts = line.split(" * ")
		if len(parts) > 2:
			logging.error("Skipped line " + line + " because it was not shaped the way we expected.")
		
		if len(parts) == 1:
			name = "Constant"
		else:
			name = string.strip(parts[1])

		if negative:
			parts[0] = "-" + parts[0]

		result[name] = float(parts[0])

	print result



convertDifficulties()