package dokugen

import (
	"testing"
)

type SimpleRankedObject struct {
	rank int
}

func (self SimpleRankedObject) Rank() int {
	return self.rank
}

func TestFiniteQueue(t *testing.T) {
	queue := NewFiniteQueue(1, DIM)
	if queue == nil {
		t.Log("We didn't get a queue back from the constructor")
		t.Fail()
	}
	objects := [...]SimpleRankedObject{{1}, {2}, {2}, {3}}
	for _, object := range objects {
		queue.Insert(object)
	}
	for _, obj := range objects {
		retrievedObj := queue.Get()
		if retrievedObj == nil {
			t.Log("We got back a nil before we were expecting to")
			t.Fail()
			continue
		}
		if retrievedObj.Rank() != obj.Rank() {
			t.Log("We got back an object with the wrong rank")
			t.Fail()
		}
		if retrievedObj != obj {
			//Note that technically the API doesn't require that items with the same rank come back out in the same order.
			//So this test will fail even in some valid cases.
			t.Log("We didn't get back the objects we put in.")
			t.Fail()
		}
	}
	if queue.Get() != nil {
		t.Log("We were able to get back more objects than what we put in.")
		t.Fail()
	}
	//TODO: test changing the rank.
	//TODO: test inputting the same object twice.
	//TODO: test inputting one object, changing its rank, and then getting (to make sure we don't get again).
}
