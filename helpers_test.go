package eavesdrop_test

import (
	"reflect"
	"testing"

	"github.com/dimmerz92/eavesdrop"
)

func TestToSet_Ints(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  map[int]struct{}
	}{
		{
			name:  "Empty slice",
			input: []int{},
			want:  map[int]struct{}{},
		},
		{
			name:  "Single element",
			input: []int{1},
			want:  map[int]struct{}{1: {}},
		},
		{
			name:  "Multiple unique elements",
			input: []int{1, 2, 3},
			want:  map[int]struct{}{1: {}, 2: {}, 3: {}},
		},
		{
			name:  "With duplicates",
			input: []int{1, 2, 2, 3, 1},
			want:  map[int]struct{}{1: {}, 2: {}, 3: {}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := eavesdrop.ToSet(test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestToSet_Strings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  map[string]struct{}
	}{
		{
			name:  "Empty slice",
			input: []string{},
			want:  map[string]struct{}{},
		},
		{
			name:  "Single element",
			input: []string{"a"},
			want:  map[string]struct{}{"a": {}},
		},
		{
			name:  "Multiple unique elements",
			input: []string{"a", "b", "c"},
			want:  map[string]struct{}{"a": {}, "b": {}, "c": {}},
		},
		{
			name:  "With duplicates",
			input: []string{"a", "b", "a", "c"},
			want:  map[string]struct{}{"a": {}, "b": {}, "c": {}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := eavesdrop.ToSet(test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
