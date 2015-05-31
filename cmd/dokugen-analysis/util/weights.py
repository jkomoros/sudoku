#originally inspired by http://scikit-learn.org/stable/auto_examples/linear_model/plot_ols.html ,
#but substantially modified from that

import numpy as np
from sklearn import datasets, linear_model

import csv

#Load the data

#TODO: allow loading from an arbitrary input
f = open('../solves.csv', 'rb')
reader = csv.reader(f)
firstRow = True

targets_basic = []
data_basic = []

for row in reader:
	if firstRow:
		firstRow = False
		continue
	convertedRow = [float(a) for a in row]
	targets_basic.append(convertedRow[:1][0])
	data_basic.append(convertedRow[1:])

#TODO: figure out if I can just create a numpy array from the beginning
targets = np.array(targets_basic)
data = np.array(data_basic)

#TODO: do folds of test/training sets.

# Split the data into training/testing sets
data_train = data[:-20]
data_test = data[-20:]

#Split the targets into training/testing sets
targets_train = targets[:-20]
targets_test = targets[-20:]


# Create linear regression object
regr = linear_model.Ridge(alpha=1.0)

# Train the model using the training sets
regr.fit(data_train, targets_train)

# The coefficients
print('Coefficients: \n', regr.coef_)
# The mean square error
print("Residual sum of squares: %.2f"
      % np.mean((regr.predict(data_test) - targets_test) ** 2))
# Explained variance score: 1 is perfect prediction
print('Variance score: %.2f' % regr.score(data_test, targets_test))