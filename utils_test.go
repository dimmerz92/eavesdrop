package eavesdrop_test

import (
	"reflect"
	"testing"

	"github.com/dimmerz92/eavesdrop"
)

func TestToSet(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		tests := []struct {
			name     string
			values   []string
			expected eavesdrop.Set[string]
		}{
			{name: "nil slice", expected: make(eavesdrop.Set[string])},
			{name: "empty slice", values: []string{}, expected: make(eavesdrop.Set[string])},
			{name: "three unique", values: []string{"1", "2", "3"}, expected: map[string]struct{}{"1": {}, "2": {}, "3": {}}},
			{name: "duplicates", values: []string{"1", "1", "2", "2", "3", "3"}, expected: map[string]struct{}{"1": {}, "2": {}, "3": {}}},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				got := eavesdrop.ToSet(test.values...)
				if !reflect.DeepEqual(got, test.expected) {
					t.Errorf("expected %#v, got %#v", test.expected, got)
				}
			})
		}
	})

	t.Run("ints", func(t *testing.T) {
		tests := []struct {
			name     string
			values   []int
			expected eavesdrop.Set[int]
		}{
			{name: "nil slice", expected: make(eavesdrop.Set[int])},
			{name: "empty slice", values: []int{}, expected: make(eavesdrop.Set[int])},
			{name: "three unique", values: []int{1, 2, 3}, expected: map[int]struct{}{1: {}, 2: {}, 3: {}}},
			{name: "duplicates", values: []int{1, 1, 2, 2, 3, 3}, expected: map[int]struct{}{1: {}, 2: {}, 3: {}}},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				got := eavesdrop.ToSet(test.values...)
				if !reflect.DeepEqual(got, test.expected) {
					t.Errorf("expected %#v, got %#v", test.expected, got)
				}
			})
		}
	})

	t.Run("float32s", func(t *testing.T) {
		tests := []struct {
			name     string
			values   []float32
			expected eavesdrop.Set[float32]
		}{
			{name: "nil slice", expected: make(eavesdrop.Set[float32])},
			{name: "empty slice", values: []float32{}, expected: make(eavesdrop.Set[float32])},
			{name: "three unique", values: []float32{1.1, 2.2, 3.3}, expected: map[float32]struct{}{1.1: {}, 2.2: {}, 3.3: {}}},
			{name: "duplicates", values: []float32{1.1, 1.1, 2.2, 2.2, 3.3, 3.3}, expected: map[float32]struct{}{1.1: {}, 2.2: {}, 3.3: {}}},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				got := eavesdrop.ToSet(test.values...)
				if !reflect.DeepEqual(got, test.expected) {
					t.Errorf("expected %#v, got %#v", test.expected, got)
				}
			})
		}
	})
}

func TestIsRelative(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		path     string
		expected bool
	}{
		{name: "both empty"},
		{name: "current dir dots", dir: ".", path: "."},
		{name: "relative dots", dir: "..", path: ".."},
		{name: "relative to current dots", dir: "..", path: "."},
		{name: "current to relative dots", dir: "..", path: "."},
		{name: "valid", dir: "/path/to/dir", path: "path/to"},
		{name: "same dir and path", dir: "/path/to/dir", path: "path/to/dir"},
		{name: "invalid", dir: "/path/to/dir", path: "to/get/there"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := eavesdrop.IsRelative(test.dir, test.path); got != test.expected {
				t.Errorf("expected %t, got %t", test.expected, got)
			}
		})
	}
	// t.Run("test valid child", func(t *testing.T) {
	//
	// 	if !eavesdrop.Is("a/b", "a/b/c/file") {
	// 		t.Fatalf("expected true, got false")
	// 	}
	// })
	//
	// t.Run("test invalid", func(t *testing.T) {
	// 	if eavesdrop.IsChild("a/b", "x/y/z/file") {
	// 		t.Fatalf("expected false, got true")
	// 	}
	// })
}
