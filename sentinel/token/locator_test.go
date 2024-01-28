package token

import (
	"fmt"
	"testing"
)

type locatorTestCase struct {
	Line int
	Col  int
}

func TestParserLocator(t *testing.T) {
	testcases := []struct {
		input        string
		expectedLocs map[int]locatorTestCase
	}{
		{
			input: "abc123",
			expectedLocs: map[int]locatorTestCase{
				0: {Line: 0, Col: 0},
				5: {Line: 0, Col: 5},
				6: {Line: 0, Col: 6}, // EOF is special
				7: {Line: -1, Col: -1},
			},
		},
		{
			input: "abc123\nabc123\nabc123\n",
			expectedLocs: map[int]locatorTestCase{
				0:    {Line: 0, Col: 0},
				5:    {Line: 0, Col: 5},
				6:    {Line: 0, Col: 6},
				7:    {Line: 1, Col: 0},
				2000: {Line: -1, Col: -1},
			},
		},
		{
			input: "\n\n\n\n\n\n\n",
			expectedLocs: map[int]locatorTestCase{
				0:    {Line: 0, Col: 0},
				5:    {Line: 5, Col: 0},
				2000: {Line: -1, Col: -1},
			},
		},
	}

	for _, testcase := range testcases {
		l := NewContentLocator("input.sentinel", []byte(testcase.input))

		for pos, locTest := range testcase.expectedLocs {
			r := l.PositionOf(pos)

			actual := fmt.Sprintf("(Line %d: Column %d)", r.Line, r.Column)
			expected := fmt.Sprintf("(Line %d: Column %d)", locTest.Line, locTest.Col)

			if actual != expected {
				t.Fatalf("Expected input %q at position %d would be %s, but got %s",
					testcase.input,
					pos,
					expected,
					actual,
				)
			}
		}
	}
}
