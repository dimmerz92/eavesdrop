package eavesdrop

import (
	"path/filepath"
	"strings"
)

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

// IsAncestor checks if the path is a child of the directory.
func IsChild(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}

	return rel != "." && !strings.HasPrefix(rel, "..")
}
