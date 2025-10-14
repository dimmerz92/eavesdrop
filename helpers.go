package eavesdrop

// ToSet returns a set from a slice of comparables.
// args:
// slice is a slice or array of a comparable type.
func ToSet[T comparable](slice []T) map[T]struct{} {
	set := make(map[T]struct{})
	for _, v := range slice {
		set[v] = struct{}{}
	}
	return set
}
