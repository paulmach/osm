package mputil

import (
	"reflect"
	"testing"

	"github.com/paulmach/orb"
)

func TestJoin(t *testing.T) {
	cases := []struct {
		name   string
		input  []Segment
		output []MultiSegment
	}{
		{
			name: "single line",
			input: []Segment{
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{0, 0}, {1, 1}},
					},
				},
			},
		},
		{
			name: "two loops",
			input: []Segment{
				{
					Line: orb.LineString{{0, 0}, {1, 1}, {1, 2}, {0, 0}},
				},
				{
					Line: orb.LineString{{1, 0}, {2, 1}, {2, 2}, {1, 0}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{1, 0}, {2, 1}, {2, 2}, {1, 0}},
					},
				},
				{
					{
						Line: orb.LineString{{0, 0}, {1, 1}, {1, 2}, {0, 0}},
					},
				},
			},
		},
		{
			name: "joins two lines",
			input: []Segment{
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
				{
					Line: orb.LineString{{1, 1}, {2, 2}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{0, 0}},
					},
					{
						Line: orb.LineString{{1, 1}, {2, 2}},
					},
				},
			},
		},
		{
			name: "joins two lines reverse order",
			input: []Segment{
				{
					Line: orb.LineString{{1, 1}, {2, 2}},
				},
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{0, 0}, {1, 1}},
					},
					{
						Line: orb.LineString{{2, 2}},
					},
				},
			},
		},
		{
			name: "joins two lines with reverse",
			input: []Segment{
				{
					Line: orb.LineString{{2, 2}, {1, 1}},
				},
				{
					Line: orb.LineString{{2, 2}, {3, 3}},
				},
			},
			output: []MultiSegment{
				{
					{
						Reversed: true,
						Line:     orb.LineString{{1, 1}},
					},
					{
						Line: orb.LineString{{2, 2}, {3, 3}},
					},
				},
			},
		},
		{
			name: "joins two lines with reverse second",
			input: []Segment{
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
				{
					Line: orb.LineString{{2, 2}, {1, 1}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{2, 2}, {1, 1}},
					},
					{
						Reversed: true,
						Line:     orb.LineString{{0, 0}},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Join(tc.input)
			compareMultiSegment(t, result, tc.output)
		})
	}
}

func TestJoin_lineString_singlePointLine(t *testing.T) {
	cases := []struct {
		name   string
		input  []Segment
		output []MultiSegment
	}{
		{
			name: "single point line, first",
			input: []Segment{
				{
					Line: orb.LineString{{1, 1}},
				},
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{0, 0}, {1, 1}},
					},
				},
			},
		},
		{
			name: "single point line, last",
			input: []Segment{
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
				{
					Line: orb.LineString{{1, 1}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{0, 0}, {1, 1}},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Join(tc.input)
			compareMultiSegment(t, result, tc.output)
		})
	}
}

func TestJoin_lineString_danglingLine(t *testing.T) {
	cases := []struct {
		name   string
		input  []Segment
		output []MultiSegment
	}{
		{
			name: "dangling line, last",
			input: []Segment{
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
				{
					Line: orb.LineString{{2, 2}, {1, 1}},
				},
				{
					Line: orb.LineString{{3, 3}, {4, 4}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{3, 3}, {4, 4}},
					},
				},
				{
					{
						Line: orb.LineString{{2, 2}, {1, 1}},
					},
					{
						Reversed: true,
						Line:     orb.LineString{{0, 0}},
					},
				},
			},
		},
		{
			name: "dangling line, middle",
			input: []Segment{
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
				{
					Line: orb.LineString{{3, 3}, {4, 4}},
				},
				{
					Line: orb.LineString{{2, 2}, {1, 1}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{2, 2}, {1, 1}},
					},
					{
						Reversed: true,
						Line:     orb.LineString{{0, 0}},
					},
				},
				{
					{
						Line: orb.LineString{{3, 3}, {4, 4}},
					},
				},
			},
		},
		{
			name: "dangling line, first",
			input: []Segment{
				{
					Line: orb.LineString{{3, 3}, {4, 4}},
				},
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
				{
					Line: orb.LineString{{2, 2}, {1, 1}},
				},
			},
			output: []MultiSegment{
				{
					{
						Line: orb.LineString{{2, 2}, {1, 1}},
					},
					{
						Reversed: true,
						Line:     orb.LineString{{0, 0}},
					},
				},
				{
					{
						Line: orb.LineString{{3, 3}, {4, 4}},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Join(tc.input)
			compareMultiSegment(t, result, tc.output)
		})
	}
}

func compareMultiSegment(t testing.TB, result, expected []MultiSegment) {
	t.Helper()

	if len(result) != len(expected) {
		t.Fatalf("length mismatch: %v != %v", len(result), len(expected))
	}

	for i, sm := range result {
		if !reflect.DeepEqual(sm, expected[i]) {
			t.Errorf("segment %d did not match", i)
			t.Logf("%v", sm)
			t.Logf("%v", expected[i])
		}
	}
}
