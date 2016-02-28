
#Instructions:
# * https://weka.wikispaces.com/Primer
# * https://weka.wikispaces.com/How+to+run+WEKA+schemes+from+commandline

#This program assumes that Weka is installed in /Applications

# Convert the provided CSV to arff, capture output, delete the arff.

# java -cp "/Applications/weka-3-6-11-oracle-jvm.app/Contents/Java/weka.jar" weka.core.converters.CSVLoader solves.csv > solves.arff