package sudoku

import (
	"testing"
)

type SimpleRankedObject struct {
	rank int
	id   string
}

func (self *SimpleRankedObject) Rank() int {
	return self.rank
}

func TestFiniteQueue(t *testing.T) {
	queue := NewFiniteQueue(1, DIM)
	if queue == nil {
		t.Log("We didn't get a queue back from the constructor")
		t.Fail()
	}
	//Note that the first item does not fit in the first bucket on purpose.
	objects := [...]*SimpleRankedObject{{3, "a"}, {4, "b"}, {4, "c"}, {5, "d"}}
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
			t.Log("We got back an object with the wrong rank: ", retrievedObj.Rank(), " is not ", obj.Rank())
			t.Fail()
		}
		convertedObj, _ := retrievedObj.(*SimpleRankedObject)
		//We tried comparing addresses here, but they weren't the same. Why? Are we copying something somewhere?
		if convertedObj.id != obj.id {
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

	//Now test changing rank.
	for _, obj := range objects {
		queue.Insert(obj)
	}

	objects[1].rank = 6
	queue.Insert(objects[1])
	//We'll sneak in a test for double-inserting here.
	queue.Insert(objects[1])
	//Keep track of our golden set, too.
	temp := objects[1]
	objects[1] = objects[2]
	objects[2] = objects[3]
	objects[3] = temp

	for _, obj := range objects {
		retrievedObj := queue.Get()
		if retrievedObj == nil {
			t.Log("We got back a nil before we were expecting to")
			t.Fail()
			continue
		}
		if retrievedObj.Rank() != obj.Rank() {
			t.Log("We got back an object with the wrong rank: ", retrievedObj.Rank(), " is not ", obj.Rank())
			t.Fail()
		}
		convertedObj, _ := retrievedObj.(*SimpleRankedObject)
		//We tried comparing addresses here, but they weren't the same. Why? Are we copying something somewhere?
		if convertedObj.id != obj.id {
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

	for _, obj := range objects {
		queue.Insert(obj)
	}

	if queue.GetSmallerThan(5) == nil {
		t.Log("We expected two items smaller than 5 and got back 0")
		t.Fail()
	}
	if queue.GetSmallerThan(5) == nil {
		t.Log("We expected two items smaller than 5 and got only one back.")
		t.Fail()
	}
	if queue.GetSmallerThan(5) != nil {
		t.Log("We expcted only two items smaller than 5 and got more back.")
		t.Fail()
	}

}

func TestSyncedFiniteQueue(t *testing.T) {
	queue := NewSyncedFiniteQueue(1, DIM)

	select {
	case <-queue.Out:
		t.Log("We got something out of the queue before we got anything back.")
		t.Fail()
	default:
		//Pass
	}

	select {
	case queue.Exit <- true:
		//pass
	default:
		t.Log("We couldn't tell the finite queue to exit.")
		t.Fail()
	}

	secondQueue := NewSyncedFiniteQueue(1, DIM)
	//Note that the first item does not fit in the first bucket on purpose.
	objects := [...]*SimpleRankedObject{{3, "a"}, {4, "b"}, {4, "c"}, {5, "d"}}
	for _, object := range objects {
		secondQueue.In <- object
	}
	for _, obj := range objects {
		retrievedObj := <-secondQueue.Out
		if retrievedObj == nil {
			t.Log("We got back a nil before we were expecting to")
			t.Fail()
			continue
		}
		if retrievedObj.Rank() != obj.Rank() {
			t.Log("We got back an object with the wrong rank: ", retrievedObj.Rank(), " is not ", obj.Rank())
			t.Fail()
		}
		convertedObj, _ := retrievedObj.(*SimpleRankedObject)
		//We tried comparing addresses here, but they weren't the same. Why? Are we copying something somewhere?
		if convertedObj.id != obj.id {
			//Note that technically the API doesn't require that items with the same rank come back out in the same order.
			//So this test will fail even in some valid cases.
			t.Log("We didn't get back the objects we put in.")
			t.Fail()
		}
	}

	select {
	case <-secondQueue.Out:
		t.Log("We were able to get something else out of the queue more than we put in.")
		t.Fail()
	default:
		//pass
	}

	secondQueue.Exit <- true

}
