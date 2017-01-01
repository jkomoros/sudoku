package sudoku

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
}

type foundStepCacheItem struct {
	prev *foundStepCacheItem
	next *foundStepCacheItem
	step *SolveStep
}

//TODO: implement foundStepCache.Copy(), which will be necessary to use it
//WITHIN computing a single step.

//remove removes the specified item and heals the list around it.
func (f *foundStepCache) remove(item *foundStepCacheItem) {
	//Check for an item that's already been removed.
	if item.prev == nil && item.next == nil {
		return
	}
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
}

//Len returns the number of items in the cache.
func (f *foundStepCache) Len() int {
	return f.length
}

//Follows the chain and returns the last cache item
func (f *foundStepCacheItem) lastItem() *foundStepCacheItem {

	//TODO: consider having this just be a field in the struct that's kept up
	//to date (for firstItem at least)

	item := f

	var lastItem *foundStepCacheItem

	for item != nil {
		lastItem = item
		item = item.next
	}

	return lastItem
}

//AddStepToQueue adds steps to a queue to be added to the cache when AddQueue is called.
func (f *foundStepCache) AddStepToQueue(step *SolveStep) {
	cacheItem := &foundStepCacheItem{
		prev: nil,
		next: nil,
		step: step,
	}

	cacheItem.next = f.queue

	f.queue = cacheItem

	if cacheItem.next != nil {
		cacheItem.next.prev = cacheItem
	}

}

//AddQueue adds all items to the cache that have been queued, in FIFO order.
func (f *foundStepCache) AddQueue() {
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
}

//insertCacheItem adds the given cache item to the cache.
func (f *foundStepCache) insertCacheItem(cacheItem *foundStepCacheItem) {

	cacheItem.next = f.firstItem

	f.firstItem = cacheItem

	if cacheItem.next != nil {
		cacheItem.next.prev = cacheItem
	}

	f.length++

	//TODO: how to handle adding steps that are effectively duplicates?

}

//AddStep adds a SolveStep to the cache.
func (f *foundStepCache) AddStep(step *SolveStep) {

	f.insertCacheItem(&foundStepCacheItem{
		prev: nil,
		next: nil,
		step: step,
	})
}

//RemoveStepsWithCells removes all steps whose target or pointer cells overlap
//with cells. That is, steps who rely on something that has changed from when
//they were earlier added.
func (f *foundStepCache) RemoveStepsWithCells(cells []CellRef) {

	//TODO: implement the DIM*DIM map to linked list entries for speed.

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

}

//GetSteps gets all steps currently in the cache.
func (f *foundStepCache) GetSteps() []*SolveStep {

	if f.Len() == 0 {
		return nil
	}

	result := make([]*SolveStep, f.Len())

	currentItem := f.firstItem
	i := 0

	for currentItem != nil {
		result[i] = currentItem.step
		i++
		currentItem = currentItem.next
	}

	return result

}
