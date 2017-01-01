package sudoku

//foundStepCache is a cache of SolveSteps that have already been found in a given
//grid. It helps us reuse previously found steps to make repeated searches
//fast when doing a HumanSolve.
type foundStepCache struct {
	//TODO: a field for which grid this is related to as a sanity check?

	//TODO: consider caching the result to GetSteps and expiring if AddStep or
	//RemoveStepsWithCells have been called.
	firstItem *foundStepCacheItem
	length    int
}

type foundStepCacheItem struct {
	prev *foundStepCacheItem
	next *foundStepCacheItem
	step *SolveStep
}

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

func (f *foundStepCache) lastItem() *foundStepCacheItem {
	//TODO: consider having this just be a field in the struct that's kept up
	//to date.
	item := f.firstItem
	var lastItem *foundStepCacheItem

	for item != nil {
		lastItem = item
		item = item.next
	}

	return lastItem
}

//AddStep adds a SolveStep to the cache.
func (f *foundStepCache) AddStep(step *SolveStep) {

	cacheItem := &foundStepCacheItem{
		prev: f.lastItem(),
		next: nil,
		step: step,
	}

	if cacheItem.prev == nil {
		//First item in the cache
		f.firstItem = cacheItem
	} else {
		//Not the first item.
		cacheItem.prev.next = cacheItem
	}

	f.length++

	//TODO: how to handle adding steps that are effectively duplicates?
}

//RemoveStepsWithCells removes all steps whose target or pointer cells overlap
//with cells. That is, steps who rely on something that has changed from when
//they were earlier added.
func (f *foundStepCache) RemoveStepsWithCells(cells []CellRef) {

	//TODO: implement

}

//GetSteps gets all steps currently in the cache.
func (f *foundStepCache) GetSteps() []*SolveStep {

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
