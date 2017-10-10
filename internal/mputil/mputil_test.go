package mputil

import (
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

func TestMultiSegmentToRing_NoAnnotation(t *testing.T) {
	cases := []struct {
		name        string
		orientation orb.Orientation
		input       MultiSegment
		output      geo.Ring
	}{
		{
			name:        "ring is direction requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Line: geo.LineString{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "ring opposite direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Line: geo.LineString{{0, 0}, {1, 0}, {1, 1}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "multi segments in direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Line: geo.LineString{{0, 0}, {1, 1}},
				},
				{
					Line: geo.LineString{{1, 0}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "multi segments in opposite direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Line: geo.LineString{{0, 0}, {1, 0}},
				},
				{
					Line: geo.LineString{{1, 1}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testToRing(t, tc.input, tc.output, tc.orientation)
		})
	}
}

func TestMultiSegmentToRing_Annotation(t *testing.T) {
	cases := []struct {
		name        string
		orientation orb.Orientation
		input       MultiSegment
		output      geo.Ring
	}{
		{
			name:        "ring is direction requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CW,
					Line:        geo.LineString{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "ring opposite direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CCW,
					Line:        geo.LineString{{0, 0}, {1, 0}, {1, 1}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "multi segments in direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CW,
					Line:        geo.LineString{{0, 0}, {1, 1}},
				},
				{
					Orientation: orb.CW,
					Line:        geo.LineString{{1, 0}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "multi segments in opposite direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CCW,
					Line:        geo.LineString{{0, 0}, {1, 0}},
				},
				{
					Orientation: orb.CCW,
					Line:        geo.LineString{{1, 1}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "reversed to correct",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CW,
					Line:        geo.LineString{{0, 0}, {1, 1}},
				},
				{
					Orientation: orb.CCW,
					Reversed:    true,
					Line:        geo.LineString{{1, 0}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "reversed to wrong direction",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CCW,
					Line:        geo.LineString{{0, 0}, {1, 0}},
				},
				{
					Orientation: orb.CW,
					Reversed:    true,
					Line:        geo.LineString{{1, 1}, {0, 0}},
				},
			},
			output: geo.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testToRing(t, tc.input, tc.output, tc.orientation)
		})
	}
}

func testToRing(t testing.TB, input MultiSegment, expected geo.Ring, orient orb.Orientation) {
	t.Helper()

	ring := input.ToRing(orient)
	if o := ring.Orientation(); o != orient {
		t.Errorf("different orientation: %v != %v", o, orient)
	}

	if !ring.Equal(expected) {
		t.Errorf("wrong ring")
		t.Logf("%v", ring)
		t.Logf("%v", expected)
	}

	// with reverse orientation
	orient *= -1
	expected.Reverse()
	ring = input.ToRing(orient)
	if o := ring.Orientation(); o != orient {
		t.Errorf("reversed, different orientation: %v != %v", o, orient)
	}

	if !ring.Equal(expected) {
		t.Errorf("reversed, wrong ring")
		t.Logf("%v", ring)
		t.Logf("%v", expected)
	}
}
