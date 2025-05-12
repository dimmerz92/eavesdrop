package utils_test

import (
	"reflect"
	"testing"

	"github.com/dimmerz92/eavesdrop/internal/utils"
)

func TestSliceToSet(t *testing.T) {
	test := []string{"hello", "world"}

	expected := map[string]struct{}{"hello": {}, "world": {}}

	set := utils.SliceToSet(test)
	if !reflect.DeepEqual(expected, set) {
		t.Errorf("expected %+v\ngot %+v", expected, set)
	}
}
