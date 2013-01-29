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

	//Test that inserting while getting works.

	//The ones that are already in should be a no-op
	for _, obj := range objects {
		queue.Insert(obj)
	}
	if item := queue.Get(); item != objects[0] {
		t.Log("A subsequent read didn't return the right object")
		t.Fail()
	}
	if item := queue.Get(); item != objects[1] {
		t.Log("A subsequent read didn't return the right object 1")
		t.Fail()
	}
	queue.Insert(objects[0])
	if item := queue.Get(); item != objects[0] {
		t.Log("An insert mid-read stream didn't return the right object")
		t.Fail()
	}
}

func TestFiniteQueueGetter(t *testing.T) {
	queue := NewFiniteQueue(1, DIM)
	//Note that the first item does not fit in the first bucket on purpose.
	objects := [...]*SimpleRankedObject{{3, "a"}, {4, "b"}, {4, "c"}, {5, "d"}}
	for _, object := range objects {
		queue.Insert(object)
	}
	getter := queue.NewGetter()
	if getter == nil {
		t.Log("We didn't get a getter back from NewGetter")
		t.Fail()
	}
	if getter.queue != queue {
		t.Log("We got back a getter that didn't have the right queue")
		t.Fail()
	}
	for _, obj := range objects {
		retrievedObj := getter.Get()
		if retrievedObj == nil {
			t.Log("We got back a nil before we were expecting to")
			t.Fail()
			continue
		}
		if retrievedObj.Rank() != obj.Rank() {
			t.Log("We got back an object with the wrong rank: ", retrievedObj.Rank(), " is not ", obj.Rank())
			t.Fail()
		}
	}
	//Now ensure that the underlying queue was not touched.
	for i := 0; i < len(objects); i++ {
		if queue.Get() == nil {
			t.Log("The underlying queue had fewer items than we expected.")
			t.Fail()
		}
	}
	if queue.Get() != nil {
		t.Log("The underlying queue had MORE objects than we expected.")
		//The next test relies on the queue being back to a clean state.
		t.FailNow()
	}

	//Test that getting is resilient to new inserts, and also that we don't hand out the same objects again.
	for _, object := range objects {
		queue.Insert(object)
	}
	getter = queue.NewGetter()
	seenObjects := make(map[RankedObject]bool)
	item := getter.Get()
	seenObjects[item] = true
	newObject := &SimpleRankedObject{3, "e"}
	queue.Insert(newObject)
	item = getter.Get()
	seenObjects[item] = true
	if item != newObject {
		t.Log("The getter did not return a new item inserted after reads started.")
		t.Fail()
	}
	//Consume the rest to check for dupes.
	for i := 0; i < len(objects)-1; i++ {
		item = getter.Get()
		if _, exists := seenObjects[item]; exists {
			t.Log("Getter returned a dupe.")
			t.Fail()
		}
		seenObjects[item] = true
	}
	item = getter.Get()
	if item != nil {
		t.Log("Getter returned more items than it should have after insert while read.")
		t.Fail()
	}

	//TODO: test that it's resilient to GETs while reading, too.
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
