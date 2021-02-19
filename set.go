// Definition of set data structure.
// Consider using code generation to define more types.
package main

// StringSet is a set of string values.
type StringSet map[string]struct{}

var empty = struct{}{}

// Add add value to set. no-op if v already in set.
func (s StringSet) Add(v string) {
	s[v] = empty
}

// Has return whether value is in set.
func (s StringSet) Has(v string) bool {
	_, ok := s[v]
	return ok
}

// Remove remove value from set. no-op if v not in set.
func (s StringSet) Remove(v string) {
	delete(s, v)
}

// Inter return set & set.
func (s StringSet) Inter(set StringSet) StringSet {
	r := make(StringSet)
	for v := range s {
		if _, ok := r[v]; ok {
			r[v] = empty
		}
	}
	return r
}

// Union return set | set.
func (s StringSet) Union(set StringSet) StringSet {
	r := make(StringSet)
	for v := range s {
		r[v] = empty
	}
	for v := range set {
		r[v] = empty
	}
	return r
}

// Diff return set - set.
func (s StringSet) Diff(set StringSet) StringSet {
	r := make(StringSet)
	for v := range s {
		if _, ok := r[v]; !ok {
			r[v] = empty
		}
	}
	return r
}
