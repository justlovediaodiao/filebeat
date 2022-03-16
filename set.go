// Definition of set data structure.

package main

// Set is a set of values.
type Set[T comparable] map[T]struct{}

var empty = struct{}{}

// Add add value to set. no-op if v already in set.
func (s Set[T]) Add(v T) {
	s[v] = empty
}

// Has return whether value is in set.
func (s Set[T]) Has(v T) bool {
	_, ok := s[v]
	return ok
}

// Remove remove value from set. no-op if v not in set.
func (s Set[T]) Remove(v T) {
	delete(s, v)
}

// Inter return set & set.
func (s Set[T]) Inter(set Set[T]) Set[T] {
	r := make(Set[T])
	for v := range s {
		if _, ok := r[v]; ok {
			r[v] = empty
		}
	}
	return r
}

// Union return set | set.
func (s Set[T]) Union(set Set[T]) Set[T] {
	r := make(Set[T])
	for v := range s {
		r[v] = empty
	}
	for v := range set {
		r[v] = empty
	}
	return r
}

// Diff return set - set.
func (s Set[T]) Diff(set Set[T]) Set[T] {
	r := make(Set[T])
	for v := range s {
		if _, ok := r[v]; !ok {
			r[v] = empty
		}
	}
	return r
}
