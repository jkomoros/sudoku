package sudoku

import (
	"encoding/json"
	"log"
	"sync"
)

//TODO: should this really be in its own file?

//foundStepCache is a cache of SolveSteps that have already been found in a given
//grid. It helps us reuse previously found steps to make repeated searches
//fast when doing a HumanSolve.
type foundStepCache struct {
	//TODO: a field for which grid this is related to as a sanity check?

	//TODO: consider caching the result to GetSteps and expiring if AddStep or
	//RemoveStepsWithCells have been called.
	firstItem *foundStepCacheItem
	length    int
	queue     *foundStepCacheItem
	//Keep track of the hashes of added steps so we won't add the same step
	//multiple times.
	addedSteps map[string]bool
	lock       sync.RWMutex
}

type foundStepCacheItem struct {
	prev *foundStepCacheItem
	next *foundStepCacheItem
	step *SolveStep
}

//TODO: implement foundStepCache.Copy(), which will be necessary to use it
//WITHIN computing a single step. (And will be very challenging to
//implement...)

//remove removes the specified item and heals the list around it.
func (f *foundStepCache) remove(item *foundStepCacheItem) {
	//Check for an item that's already been removed.
	if item.prev == nil && item.next == nil {
		return
	}

	//This is not protected by a lock because our callers must hold the lock
	//already.

	if item.prev == nil {
		//first item
		f.firstItem = item.next
		f.firstItem.prev = nil
	} else {
		item.prev.next = item.next
		if item.next != nil {
			item.next.prev = item.prev
		}
	}
	//Make sure the item is orphaned so if we call remove on it again it won't
	//do anything
	item.prev = nil
	item.next = nil
	f.length--

	delete(f.addedSteps, solveStepHash(item.step))

}

//Len returns the number of items in the cache.
func (f *foundStepCache) Len() int {
	return f.length
}

func (f *foundStepCacheItem) debugPrint() {
	log.Println("***Starting***")
	f.debugPrintImpl(0)
}

func (f *foundStepCacheItem) debugPrintImpl(count int) {
	log.Println("Item", count, &f, "prev:", &f.prev, "next:", &f.next, "step:", f.step)
	if f.next == nil {
		return
	}

	f.next.debugPrintImpl(count + 1)
}

//returns a unique string representing this step.
func solveStepHash(step *SolveStep) string {
	//We tried using fmt.Sprintf("%#v", step) here, but it was slower

	//TODO: normalize step before hashing
	result, err := json.Marshal(step)

	if err != nil {
		panic("Couldn't serialize SolveStep")
	}

	return string(result)
}

//AddStepToQueue adds steps to a queue to be added to the cache when AddQueue is called.
func (f *foundStepCache) AddStepToQueue(step *SolveStep) {
	cacheItem := &foundStepCacheItem{
		prev: nil,
		next: nil,
		step: step,
	}

	f.lock.Lock()

	cacheItem.next = f.queue

	f.queue = cacheItem

	if cacheItem.next != nil {
		cacheItem.next.prev = cacheItem
	}

	f.lock.Unlock()

}

//AddQueue adds all items to the cache that have been queued, in LIFO order.
func (f *foundStepCache) AddQueue() {
	f.lock.Lock()
	item := f.queue
	var next *foundStepCacheItem
	for item != nil {
		//next will be mangled when we're added to the cache, so take note of
		//it now.
		next = item.next
		f.insertCacheItem(item)
		item = next
	}
	f.queue = nil
	f.lock.Unlock()
}

//insertCacheItem adds the given cache item to the cache.
func (f *foundStepCache) insertCacheItem(cacheItem *foundStepCacheItem) {

	//You MUST have the lock held before calling this method.

	hash := solveStepHash(cacheItem.step)

	if f.addedSteps == nil {
		f.addedSteps = make(map[string]bool)
	}

	if f.addedSteps[hash] == true {
		//Skip this one, it was already added.
		return
	}

	cacheItem.prev = nil

	cacheItem.next = f.firstItem

	f.firstItem = cacheItem

	if cacheItem.next != nil {
		cacheItem.next.prev = cacheItem
	}

	f.addedSteps[hash] = true

	f.length++

}

//AddStep adds a SolveStep to the cache.
func (f *foundStepCache) AddStep(step *SolveStep) {

	f.lock.Lock()
	f.insertCacheItem(&foundStepCacheItem{
		prev: nil,
		next: nil,
		step: step,
	})
	f.lock.Unlock()
}

//RemoveStepsWithCells removes all steps whose target or pointer cells overlap
//with cells. That is, steps who rely on something that has changed from when
//they were earlier added.
func (f *foundStepCache) RemoveStepsWithCells(cells []CellRef) {

	//TODO: implement the DIM*DIM map to linked list entries for speed.

	f.lock.Lock()

	set := make(cellSet)

	for _, cell := range cells {
		set[cell] = true
	}

	currentItem := f.firstItem
	var nextItem *foundStepCacheItem

	for currentItem != nil {

		itemRemoved := false
		//We need to save this now, because if we remove ourselves our next
		//will be blown away
		nextItem = currentItem.next

		for _, ref := range currentItem.step.TargetCells {
			if set[ref] {
				f.remove(currentItem)
				itemRemoved = true
				break
			}
		}

		if !itemRemoved {
			for _, ref := range currentItem.step.PointerCells {
				if set[ref] {
					f.remove(currentItem)
					break
				}
			}
		}

		currentItem = nextItem
	}

	f.lock.Unlock()

}

//GetSteps gets all steps currently in the cache.
func (f *foundStepCache) GetSteps() []*SolveStep {

	if f.Len() == 0 {
		return nil
	}

	f.lock.RLock()

	result := make([]*SolveStep, f.Len())

	currentItem := f.firstItem
	i := 0

	for currentItem != nil {
		result[i] = currentItem.step
		i++
		currentItem = currentItem.next
	}

	f.lock.RUnlock()

	return result

}
