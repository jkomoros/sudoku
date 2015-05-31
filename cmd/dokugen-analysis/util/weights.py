#originally inspired by http://scikit-learn.org/stable/auto_examples/linear_model/plot_ols.html ,
#but substantially modified from that

import numpy as np
from sklearn import datasets, linear_model
from sklearn.cross_validation import cross_val_score, KFold

import csv

#Load the data

#TODO: allow loading from an arbitrary input
f = open('../solves.csv', 'rb')
reader = csv.reader(f)
firstRow = True

targets_basic = []
data_basic = []

labels = []

for row in reader:
	if firstRow:
		firstRow = False
		labels = row
		continue
	convertedRow = [float(a) for a in row]
	targets_basic.append(convertedRow[:1][0])
	data_basic.append(convertedRow[1:])

#TODO: figure out if I can just create a numpy array from the beginning
targets = np.array(targets_basic)
data = np.array(data_basic)

# Create linear regression object
regr = linear_model.Ridge(alpha=1.0)

# Train the model using the training sets
regr.fit(data, targets)

print("Coefficients")
for i in xrange(0, len(regr.coef_)):
	print(labels[i+1] + " = " + str(regr.coef_[i]))

# # The mean square error
# print("Residual sum of squares: %.2f"
#       % np.mean((regr.predict(data_test) - targets_test) ** 2))
# # Explained variance score: 1 is perfect prediction
# print('Variance score: %.2f' % regr.score(data_test, targets_test))

K = 20 #folds
R2 = cross_val_score(regr, data, y=targets, cv=KFold(targets.size, K)).mean()
print("The %d-Folds estimate of the coefficient of determination is R2 = %s"
      % (K, R2))

#TODO: export the model in a way that Go can run it