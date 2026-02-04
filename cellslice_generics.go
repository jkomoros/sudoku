package sudoku

import "sort"

// Package sudoku provides generic helper functions for slice operations.
// These functions are used internally by CellSlice, MutableCellSlice,
// CellRefSlice, and IntSlice to reduce code duplication while maintaining
// backward compatibility with the existing API.
//
// The generic helpers enable compile-time type safety and eliminate
// significant code duplication across the four slice types.

// GenericFilter is the internal generic implementation of Filter().
// Each slice type wraps this function to maintain its specific return type
// and preserve method chaining capabilities.
func GenericFilter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// GenericSubset is the internal generic implementation of Subset().
// Returns a new slice containing items at the specified indexes.
func GenericSubset[T any](slice []T, indexes IntSlice) []T {
	result := make([]T, len(indexes))
	max := len(slice)
	for i, index := range indexes {
		if index >= max {
			// This probably is indicative of a larger problem.
			continue
		}
		result[i] = slice[index]
	}
	return result
}

// GenericInverseSubset is the internal generic implementation of InverseSubset().
// Returns a new slice containing all items NOT at the specified indexes.
func GenericInverseSubset[T any](slice []T, indexes IntSlice) []T {
	var result []T

	// Ensure indexes are in sorted order.
	sort.Ints(indexes)

	// Index into indexes we're considering
	currentIndex := 0

	for i := 0; i < len(slice); i++ {
		if currentIndex < len(indexes) && i == indexes[currentIndex] {
			// Skip it!
			currentIndex++
		} else {
			// Output it!
			result = append(result, slice[i])
		}
	}

	return result
}

// GenericRemoveCells is the internal generic implementation of RemoveCells().
// Returns a new slice with the items in toRemove excluded.
func GenericRemoveCells[T comparable](slice []T, toRemove map[T]bool) []T {
	var result []T
	for _, item := range slice {
		if !toRemove[item] {
			result = append(result, item)
		}
	}
	return result
}

// GenericCollectNums is the internal generic implementation of CollectNums().
// Applies the fetcher function to each element and returns the results as an IntSlice.
func GenericCollectNums[T any](slice []T, fetcher func(T) int) IntSlice {
	result := make(IntSlice, len(slice))
	for i, item := range slice {
		result[i] = fetcher(item)
	}
	return result
}

// GenericToCellSet is the internal generic implementation of toCellSet helpers.
// Converts a slice to a map for set operations.
func GenericToCellSet[T comparable](slice []T) map[T]bool {
	result := make(map[T]bool)
	for _, item := range slice {
		result[item] = true
	}
	return result
}

// genericSorter is a generic implementation of sort.Interface.
// It eliminates the need for separate sorter types for each slice type.
type genericSorter[T any] struct {
	slice []T
	less  func(i, j int) bool
}

func (s *genericSorter[T]) Len() int {
	return len(s.slice)
}

func (s *genericSorter[T]) Swap(i, j int) {
	s.slice[i], s.slice[j] = s.slice[j], s.slice[i]
}

func (s *genericSorter[T]) Less(i, j int) bool {
	return s.less(i, j)
}

// GenericSort is the internal generic implementation of Sort().
// It sorts the slice in place using the provided less function.
func GenericSort[T any](slice []T, less func(i, j int) bool) {
	sort.Sort(&genericSorter[T]{slice, less})
}
