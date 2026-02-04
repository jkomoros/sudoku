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

// GenericIntersection is the internal generic implementation of Intersection().
// Returns a new slice containing items that appear in both slices.
func GenericIntersection[T comparable](slice1, slice2 []T) []T {
	set := GenericToCellSet(slice2)
	var result []T
	for _, item := range slice1 {
		if set[item] {
			result = append(result, item)
		}
	}
	return result
}

// GenericDifference is the internal generic implementation of Difference().
// Returns a new slice containing items in slice1 that are not in slice2.
func GenericDifference[T comparable](slice1, slice2 []T) []T {
	set := GenericToCellSet(slice2)
	var result []T
	for _, item := range slice1 {
		if !set[item] {
			result = append(result, item)
		}
	}
	return result
}

// GenericUnion is the internal generic implementation of Union().
// Returns a new slice containing all items from both slices (no duplicates).
func GenericUnion[T comparable](slice1, slice2 []T) []T {
	set := GenericToCellSet(slice1)
	result := make([]T, len(slice1))
	copy(result, slice1)

	for _, item := range slice2 {
		if !set[item] {
			result = append(result, item)
			set[item] = true
		}
	}
	return result
}

// GenericIntersectionSet returns the intersection of two sets (map[T]bool).
// Items that appear in both sets are included in the result.
func GenericIntersectionSet[T comparable](set1, set2 map[T]bool) map[T]bool {
	result := make(map[T]bool)
	for item, value := range set1 {
		if value {
			if val, ok := set2[item]; ok && val {
				result[item] = true
			}
		}
	}
	return result
}

// GenericDifferenceSet returns the difference of two sets (items in set1 but not in set2).
func GenericDifferenceSet[T comparable](set1, set2 map[T]bool) map[T]bool {
	result := make(map[T]bool)
	for item, value := range set1 {
		if value {
			if val, ok := set2[item]; !ok || !val {
				result[item] = true
			}
		}
	}
	return result
}

// GenericUnionSet returns the union of two sets (all items from both sets).
func GenericUnionSet[T comparable](set1, set2 map[T]bool) map[T]bool {
	result := make(map[T]bool)
	for item, value := range set1 {
		result[item] = value
	}
	for item, value := range set2 {
		result[item] = value
	}
	return result
}

// GenericOverlaps checks if two sets have any items in common.
// This is optimized compared to checking len(intersection) > 0.
func GenericOverlaps[T comparable](set1, set2 map[T]bool) bool {
	for item, value := range set1 {
		if value {
			if val, ok := set2[item]; ok && val {
				return true
			}
		}
	}
	return false
}

// GenericSetToSlice converts a set (map[T]bool) to a slice of items where value is true.
func GenericSetToSlice[T comparable](set map[T]bool) []T {
	var result []T
	for item, val := range set {
		if val {
			result = append(result, item)
		}
	}
	return result
}

// GenericSlicesEqual checks if two slices contain the same elements (order-independent).
// Uses set-based comparison for efficiency.
func GenericSlicesEqual[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	set1 := GenericToCellSet(slice1)
	set2 := GenericToCellSet(slice2)

	if len(set1) != len(set2) {
		return false
	}

	for item := range set1 {
		if !set2[item] {
			return false
		}
	}

	return true
}

// Predicate factory functions for common cell filtering operations.
// These create reusable predicates that can be shared across CellSlice and MutableCellSlice.

// IsUnfilled returns a predicate that checks if a cell is not filled.
func IsUnfilled() func(Cell) bool {
	return func(cell Cell) bool {
		return cell.Number() == 0
	}
}

// IsFilled returns a predicate that checks if a cell is filled.
func IsFilled() func(Cell) bool {
	return func(cell Cell) bool {
		return cell.Number() != 0
	}
}

// HasPossible returns a predicate that checks if a cell has a specific number as a possibility.
func HasPossible(possible int) func(Cell) bool {
	return func(cell Cell) bool {
		return cell.Possible(possible)
	}
}

// HasNumPossibilities returns a predicate that checks if a cell has exactly target possibilities.
func HasNumPossibilities(target int) func(Cell) bool {
	return func(cell Cell) bool {
		return len(cell.Possibilities()) == target
	}
}

// HasPossibilities returns a predicate that checks if a cell has any possibilities.
func HasPossibilities() func(Cell) bool {
	return func(cell Cell) bool {
		return len(cell.Possibilities()) > 0
	}
}
