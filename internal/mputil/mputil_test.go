package mputil

import (
	"testing"

	"github.com/paulmach/orb"
)

func TestMultiSegment_ToRing_noAnnotation(t *testing.T) {
	cases := []struct {
		name        string
		orientation orb.Orientation
		input       MultiSegment
		output      orb.Ring
	}{
		{
			name:        "ring is direction requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Line: orb.LineString{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "ring opposite direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Line: orb.LineString{{0, 0}, {1, 0}, {1, 1}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "multi segments in direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Line: orb.LineString{{0, 0}, {1, 1}},
				},
				{
					Line: orb.LineString{{1, 0}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "multi segments in opposite direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Line: orb.LineString{{0, 0}, {1, 0}},
				},
				{
					Line: orb.LineString{{1, 1}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testToRing(t, tc.input, tc.output, tc.orientation)
		})
	}
}

func TestMultiSegment_ToRing_annotation(t *testing.T) {
	cases := []struct {
		name        string
		orientation orb.Orientation
		input       MultiSegment
		output      orb.Ring
	}{
		{
			name:        "ring is direction requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CW,
					Line:        orb.LineString{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "ring opposite direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CCW,
					Line:        orb.LineString{{0, 0}, {1, 0}, {1, 1}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "multi segments in direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CW,
					Line:        orb.LineString{{0, 0}, {1, 1}},
				},
				{
					Orientation: orb.CW,
					Line:        orb.LineString{{1, 0}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "multi segments in opposite direction of requested",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CCW,
					Line:        orb.LineString{{0, 0}, {1, 0}},
				},
				{
					Orientation: orb.CCW,
					Line:        orb.LineString{{1, 1}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "reversed to correct",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CW,
					Line:        orb.LineString{{0, 0}, {1, 1}},
				},
				{
					Orientation: orb.CCW,
					Reversed:    true,
					Line:        orb.LineString{{1, 0}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
		{
			name:        "reversed to wrong direction",
			orientation: orb.CW,
			input: MultiSegment{
				{
					Orientation: orb.CCW,
					Line:        orb.LineString{{0, 0}, {1, 0}},
				},
				{
					Orientation: orb.CW,
					Reversed:    true,
					Line:        orb.LineString{{1, 1}, {0, 0}},
				},
			},
			output: orb.Ring{{0, 0}, {1, 1}, {1, 0}, {0, 0}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testToRing(t, tc.input, tc.output, tc.orientation)
		})
	}
}

func testToRing(t testing.TB, input MultiSegment, expected orb.Ring, orient orb.Orientation) {
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
