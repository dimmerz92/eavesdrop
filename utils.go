package eavesdrop

import (
	"path/filepath"
	"strings"
)

type Set[T comparable] map[T]struct{}

// ToSet adds all passed values to a Set.
func ToSet[T comparable](values ...T) Set[T] {
	set := make(Set[T])
	for _, item := range values {
		set[item] = struct{}{}
	}
	return set
}

// IsRelative checks if the directory is relative to the given path.
func IsRelative(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}

	return rel != "." && !strings.HasPrefix(rel, "..")
}
