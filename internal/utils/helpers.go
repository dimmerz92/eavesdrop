package utils

// SliceToSet converts a slice to a set and returns it.
func SliceToSet[T comparable](slice []T) map[T]struct{} {
	set := make(map[T]struct{})
	for _, item := range slice {
		set[item] = struct{}{}
	}
	return set
}
