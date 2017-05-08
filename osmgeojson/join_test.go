package osmgeojson

import (
	"testing"

	"github.com/paulmach/orb/geo"
)

func TestJoinLineString(t *testing.T) {
	cases := []struct {
		name   string
		input  []geo.LineString
		output []geo.LineString
	}{
		{
			name: "single line",
			input: []geo.LineString{
				{
					geo.NewPoint(0, 0),
					geo.NewPoint(1, 1),
				},
			},
			output: []geo.LineString{
				{
					geo.NewPoint(0, 0),
					geo.NewPoint(1, 1),
				},
			},
		},
		{
			name: "two loops",
			input: []geo.LineString{
				{
					geo.NewPoint(0, 0),
					geo.NewPoint(1, 1),
					geo.NewPoint(1, 2),
					geo.NewPoint(0, 0),
				},
				{
					geo.NewPoint(1, 0),
					geo.NewPoint(2, 1),
					geo.NewPoint(2, 2),
					geo.NewPoint(1, 0),
				},
			},
			output: []geo.LineString{
				{
					geo.NewPoint(1, 0),
					geo.NewPoint(2, 1),
					geo.NewPoint(2, 2),
					geo.NewPoint(1, 0),
				},
				{
					geo.NewPoint(0, 0),
					geo.NewPoint(1, 1),
					geo.NewPoint(1, 2),
					geo.NewPoint(0, 0),
				},
			},
		},
		{
			name: "joins two lines",
			input: []geo.LineString{
				{
					geo.NewPoint(0, 0),
					geo.NewPoint(1, 1),
				},
				{
					geo.NewPoint(2, 2),
					geo.NewPoint(1, 1),
				},
			},
			output: []geo.LineString{
				{
					geo.NewPoint(2, 2),
					geo.NewPoint(1, 1),
					geo.NewPoint(0, 0),
				},
			},
		},
		{
			name: "dangling line",
			input: []geo.LineString{
				{
					geo.NewPoint(0, 0),
					geo.NewPoint(1, 1),
				},
				{
					geo.NewPoint(2, 2),
					geo.NewPoint(1, 1),
				},
				{
					geo.NewPoint(3, 3),
					geo.NewPoint(4, 4),
				},
			},
			output: []geo.LineString{
				{
					geo.NewPoint(3, 3),
					geo.NewPoint(4, 4),
				},
				{
					geo.NewPoint(2, 2),
					geo.NewPoint(1, 1),
					geo.NewPoint(0, 0),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := joinLineStrings(tc.input)
			if len(result) != len(tc.output) {
				t.Fatalf("not matching lines: %v != %v", len(result), len(tc.output))
			}

			for i, l := range result {
				if !l.Equal(tc.output[i]) {
					t.Errorf("line %d did not match", i)
					t.Logf("%v", l)
					t.Logf("%v", tc.output[i])
				}
			}
		})
	}

}

func TestMerge(t *testing.T) {
	cases := []struct {
		name   string
		input  geo.MultiLineString
		output geo.LineString
	}{
		{
			name:   "empty",
			input:  geo.MultiLineString{},
			output: geo.LineString{},
		},
		{
			name: "single line",
			input: geo.MultiLineString{
				{
					geo.NewPoint(0, 0),
					geo.NewPoint(1, 1),
				},
			},
			output: geo.LineString{
				geo.NewPoint(0, 0),
				geo.NewPoint(1, 1),
			},
		},
		{
			name: "multiple lines",
			input: geo.MultiLineString{
				{
					geo.NewPoint(0, 0),
					geo.NewPoint(1, 1),
				},
				{
					geo.NewPoint(2, 2),
					geo.NewPoint(3, 3),
				},
			},
			output: geo.LineString{
				geo.NewPoint(0, 0),
				geo.NewPoint(1, 1),
				geo.NewPoint(2, 2),
				geo.NewPoint(3, 3),
			},
		},
		{
			name: "empty line",
			input: geo.MultiLineString{
				{
					geo.NewPoint(0, 0),
					geo.NewPoint(1, 1),
				},
				{},
				{
					geo.NewPoint(2, 2),
					geo.NewPoint(3, 3),
				},
			},
			output: geo.LineString{
				geo.NewPoint(0, 0),
				geo.NewPoint(1, 1),
				geo.NewPoint(2, 2),
				geo.NewPoint(3, 3),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := merge(tc.input)
			if !result.Equal(tc.output) {
				t.Errorf("results did not match")
				t.Logf("%v", result)
				t.Logf("%v", tc.output)
			}
		})
	}

}
