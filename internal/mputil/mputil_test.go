package mputil

import (
	"reflect"
	"testing"
	"time"

	"github.com/onMaps/osm"
	"github.com/paulmach/orb"
)

func TestMultiSegment_LineString(t *testing.T) {
	ms := MultiSegment{
		{
			Line: orb.LineString{{1, 1}, {2, 2}},
		},
		{
			Line: orb.LineString{{3, 3}, {4, 4}},
		},
	}

	ls := ms.LineString()
	expected := orb.LineString{{1, 1}, {2, 2}, {3, 3}, {4, 4}}

	if !ls.Equal(expected) {
		t.Errorf("incorrect line string: %v", ls)
	}
}

func TestMultiSegment_Ring_noAnnotation(t *testing.T) {
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
			testRing(t, tc.input, tc.output, tc.orientation)
		})
	}
}

func TestMultiSegment_Ring_annotation(t *testing.T) {
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
			testRing(t, tc.input, tc.output, tc.orientation)
		})
	}
}

func TestMultiSegment_Orientation(t *testing.T) {
	ms := MultiSegment{
		{
			Line: orb.LineString{{0, 0}, {1, 0}},
		},
		{
			Line: orb.LineString{{1, 1}, {0, 1}},
		},
	}

	if o := ms.Orientation(); o != orb.CCW {
		t.Errorf("incorrect orientation: %v != %v", o, orb.CCW)
	}
}

func TestGroup(t *testing.T) {
	members := osm.Members{
		{Type: osm.TypeNode, Ref: 1},
		{Type: osm.TypeWay, Ref: 1, Role: "outer", Orientation: orb.CW},
		{Type: osm.TypeWay, Ref: 2, Role: "inner", Orientation: orb.CCW},
		{Type: osm.TypeWay, Ref: 3, Role: "inner", Orientation: orb.CCW},
		{Type: osm.TypeRelation, Ref: 3},
	}

	ways := map[osm.WayID]*osm.Way{
		1: {ID: 1, Nodes: osm.WayNodes{
			{Lat: 1.0, Lon: 2.0},
			{Lat: 2.0, Lon: 3.0},
		}},
		2: {ID: 1, Nodes: osm.WayNodes{
			{Lat: 3.0, Lon: 4.0},
			{Lat: 4.0, Lon: 5.0},
		}},
	}

	outer, inner, tainted := Group(members, ways, time.Time{})
	if !tainted {
		t.Errorf("should be tainted")
	}

	// outer
	expected := []Segment{
		{
			Index: 1, Orientation: orb.CW, Reversed: true,
			Line: orb.LineString{{3, 2}, {2, 1}},
		},
	}
	if !reflect.DeepEqual(outer, expected) {
		t.Errorf("incorrect outer: %+v", inner)
	}

	// inner
	expected = []Segment{
		{
			Index: 2, Orientation: orb.CCW, Reversed: true,
			Line: orb.LineString{{5, 4}, {4, 3}},
		},
	}
	if !reflect.DeepEqual(inner, expected) {
		t.Errorf("incorrect inner: %+v", inner)
	}
}

func TestGroup_zeroLengthWays(t *testing.T) {
	// should not panic
	Group(
		osm.Members{
			{Type: osm.TypeWay, Ref: 1, Role: "outer", Orientation: orb.CW},
			{Type: osm.TypeWay, Ref: 1, Role: "inner", Orientation: orb.CCW},
		},
		map[osm.WayID]*osm.Way{
			1: {ID: 1},
		},
		time.Time{},
	)
}

func testRing(t testing.TB, input MultiSegment, expected orb.Ring, orient orb.Orientation) {
	t.Helper()

	ring := input.Ring(orient)
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
	ring = input.Ring(orient)
	if o := ring.Orientation(); o != orient {
		t.Errorf("reversed, different orientation: %v != %v", o, orient)
	}

	if !ring.Equal(expected) {
		t.Errorf("reversed, wrong ring")
		t.Logf("%v", ring)
		t.Logf("%v", expected)
	}
}
