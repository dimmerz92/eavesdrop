package notify

// SliceToSet converts a string slice to a map set and returns it.
func SliceToSet(slice []string) map[string]struct{} {
	set := make(map[string]struct{})

	for _, item := range slice {
		set[item] = struct{}{}
	}

	return set
}
