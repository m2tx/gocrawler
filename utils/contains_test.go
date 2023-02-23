package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceContainsElement(t *testing.T) {
	testcase := []struct {
		List     []string
		Value    string
		Expected bool
	}{
		{List: []string{"A", "B", "C"}, Value: "A", Expected: true},
		{List: []string{"A", "B", "C"}, Value: "B", Expected: true},
		{List: []string{"A", "B", "C"}, Value: "C", Expected: true},
		{List: []string{"A", "B", "C"}, Value: "D", Expected: false},
		{List: []string{"A", "B", "C"}, Value: "", Expected: false},
	}
	for _, test := range testcase {
		t.Run(test.Value, func(t *testing.T) {
			contains := SliceContainsElement(test.List, test.Value)
			assert.Equal(t, test.Expected, contains)
		})
	}
}

func TestSliceContainsSlice(t *testing.T) {
	testcase := []struct {
		List     []string
		Values   []string
		Expected bool
	}{
		{List: []string{"A", "B", "C"}, Values: []string{"A", "C"}, Expected: true},
		{List: []string{"A", "B", "C"}, Values: []string{"A", "B"}, Expected: true},
		{List: []string{"A", "B", "C"}, Values: []string{"B", "C"}, Expected: true},
		{List: []string{"A", "B", "C"}, Values: []string{"A", "D"}, Expected: false},
		{List: []string{"A", "B", "C"}, Values: []string{"B", "D"}, Expected: false},
		{List: []string{"A", "B", "C"}, Values: []string{"C", "D"}, Expected: false},
		{List: []string{"A", "B", "C"}, Values: []string{"D", "E"}, Expected: false},
	}
	for _, test := range testcase {
		t.Run(fmt.Sprint(test.Values), func(t *testing.T) {
			contains := SliceContainsSlice(test.List, test.Values)
			assert.Equal(t, test.Expected, contains)
		})
	}
}
