package sudoku

import (
	"testing"
	"time"
)

type SimpleRankedObject struct {
	_rank int
	id    string
}

func (self *SimpleRankedObject) rank() int {
	return self._rank
}

func TestReadOnlyCellQueue(t *testing.T) {
	grid := NewGrid()
	//Load up a realistic grid with realistic ranks for cells.
	grid.Load(ADVANCED_TEST_GRID)

	queue := readOnlyCellQueue{
		grid: grid,
	}

	queue.defaultRefs()
	queue.fix()

	getter := queue.NewGetter()

	lastRank := 0
	counter := 0
	item := getter.Get()
	for item != nil {

		if item.rank() < lastRank {
			t.Error("Item", counter, "was smaller than a rank already seen:", item.rank())
		}

		lastRank = item.rank()

		item = getter.Get()

		counter++
	}

	if counter != DIM*DIM {
		t.Error("Default getter didn't give us all items")
	}

	//Make sure a new getter starts at beginning
	newGetter := queue.NewGetter()

	item = newGetter.Get()

	if item == nil {
		t.Error("Getting from a new getter gave us nil")
	}

	item = newGetter.GetSmallerThan(4)

	for item != nil {
		if item.rank() >= 4 {
			t.Error("GetSmallerThan returned too high a rank", item.rank())
		}
		item = newGetter.GetSmallerThan(4)
	}

	//Test copying in state from a previous one.

	modification := newCellModification(grid.Cell(0, 0))
	modification.Number = 5
	modifiedGrid := grid.CopyWithModifications(GridModifcation{modification})

	newQueue := readOnlyCellQueue{
		grid:     modifiedGrid,
		cellRefs: queue.cellRefs,
	}

	newQueue.fix()

	getter = newQueue.NewGetter()

	item = getter.Get()
	counter = 0
	lastRank = 0
	for item != nil {

		if item.rank() < lastRank {
			t.Error("In new getter got an out-of-rank item")
		}
		lastRank = item.rank()
		counter++
		item = getter.Get()
	}

	if counter != DIM*DIM {
		t.Error("New getter returned too early:", counter)
	}

}

func TestFiniteQueue(t *testing.T) {
	queue := newFiniteQueue(1, DIM)
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
		if retrievedObj.rank() != obj.rank() {
			t.Log("We got back an object with the wrong rank: ", retrievedObj.rank(), " is not ", obj.rank())
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

	objects[1]._rank = 6
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
		if retrievedObj.rank() != obj.rank() {
			t.Log("We got back an object with the wrong rank: ", retrievedObj.rank(), " is not ", obj.rank())
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
	queue := newFiniteQueue(1, DIM)
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
	for _, obj := range objects {
		retrievedObj := getter.Get()
		if retrievedObj == nil {
			t.Log("We got back a nil before we were expecting to")
			t.Fail()
			continue
		}
		if retrievedObj.rank() != obj.rank() {
			t.Log("We got back an object with the wrong rank: ", retrievedObj.rank(), " is not ", obj.rank())
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
	seenObjects := make(map[rankedObject]bool)
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
	//Test that it's resilient to removes.
	for _, object := range objects[1:] {
		queue.Insert(object)
	}
	getter = queue.NewGetter()
	_ = getter.Get()
	//Get into the second bucket.
	_ = getter.Get()
	queue.Get()
	removedItem := queue.Get()
	item = getter.Get()
	for item != nil {
		if item == removedItem {
			t.Log("We got an item we shouldn't have gotten because the underlying queue changed.")
			t.Fail()
		}
		item = getter.Get()
	}

	//Test that we will get an item back out if its rank changes.
	for _, object := range objects[1:] {
		queue.Insert(object)
	}
	getter = queue.NewGetter()
	_ = getter.Get()
	//Now change its rank and make sure we get again.
	rawObject := objects[0]
	rawObject._rank = 1
	queue.Insert(rawObject)
	item = getter.Get()
	if item != rawObject {
		t.Log("We expected to see the same item again since its rank changed but we did not.")
		t.Fail()
	}
	//Note: queue still has many items in it.
}

func TestSyncedFiniteQueue(t *testing.T) {

	done := make(chan bool, 1)

	queue := newSyncedFiniteQueue(1, DIM, done)

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

	secondDone := make(chan bool, 1)

	secondQueue := newSyncedFiniteQueue(1, DIM, secondDone)
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
		if retrievedObj.rank() != obj.rank() {
			t.Log("We got back an object with the wrong rank: ", retrievedObj.rank(), " is not ", obj.rank())
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

	if secondQueue.IsDone() {
		t.Log("The second queue thought it was done even though the items haven't been marked as done.")
		t.Fail()
	}

	secondQueue.ItemDone <- true
	secondQueue.ItemDone <- true
	secondQueue.ItemDone <- true
	secondQueue.ItemDone <- true

	//Having this here helps ensure that the NEXT test's condition is true if ti will be.
	select {
	case <-secondQueue.done:
		//As expected.

	//This seems like a crazy amount of time to wait. It used to be '10', but in go 1.5 that
	//was reliably not enough time, and deeper investigation revealed that everything was
	//operating fine. Presumably something about when GC is scheduled or something?
	case <-time.After(1 * time.Second):
		t.Log("We didn't get the done signal after some time of waiting.")
		t.Fail()
	}

	if !secondQueue.IsDone() {
		t.Log("The second queue didn't realize it was done even though all items are marked as done.")
		t.Fail()
	}

	secondQueue.Exit <- true

}
