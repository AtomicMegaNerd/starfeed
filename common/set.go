package common

import (
	"fmt"
	"iter"
	"strings"
)

// This is a basic wrapper around map[T]struct{} to provide a basic non thread safe set type.
// Please note that it is best to construct this set using the NewSet function to ensure that the
// underlying map is initialized.
type Set[T comparable] struct {
	elems map[T]struct{}
}

// We can construct a new Set with any number of elements
func NewSet[T comparable](items ...T) *Set[T] {
	set := Set[T]{elems: make(map[T]struct{})}
	for _, item := range items {
		set.Add(item)
	}
	return &set
}

// Add an element to the Set
func (s *Set[T]) Add(item T) {
	// Just in-case someone doesn't listen and use NewSet()
	if s.elems == nil {
		s.elems = make(map[T]struct{})
	}
	s.elems[item] = struct{}{}
}

// Remove an item from the Set
func (s *Set[T]) Remove(item T) {
	delete(s.elems, item)
}

// Returns true if the Set contain a specified element otherwise false
func (s *Set[T]) Contains(item T) bool {
	_, ok := s.elems[item]
	return ok
}

// How many elements in the Set?
func (s *Set[T]) Len() int {
	return len(s.elems)
}

// All() allows us to do for ... range over the Set. Obviously mutating the set
// while iterating over it is not ok as we iterate over the live elements here
func (s *Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for elem := range s.elems {
			if !yield(elem) {
				return
			}
		}
	}
}

func (s *Set[T]) String() string {
	var b strings.Builder
	b.WriteString("{")
	first := true
	for elem := range s.elems {
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", elem)
		first = false
	}
	b.WriteString("}")
	return b.String()
}
