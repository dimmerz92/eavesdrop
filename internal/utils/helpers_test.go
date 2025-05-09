package utils_test

import (
	"reflect"
	"testing"

	"github.com/dimmerz92/eavesdrop/internal/utils"
)

func TestSliceToSet(t *testing.T) {
	tests := [][]any{
		{"hello", "world"},
		{1, 2, 3},
		{'a', '5'},
	}

	expected := []map[any]struct{}{
		{"hello": {}, "world": {}},
		{1: {}, 2: {}, 3: {}},
		{'a': {}, '5': {}},
	}

	for i, test := range tests {
		set := utils.SliceToSet(test)
		if !reflect.DeepEqual(expected[i], set) {
			t.Errorf("expected %+v\ngot %+v", expected[i], set)
		}
	}
}
