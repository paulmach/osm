package osm

import (
	"reflect"
	"testing"
)

func TestWayPolygon(t *testing.T) {
	w := &Way{}
	w.Nodes = []WayNode{
		{ID: 1}, {ID: 2},
		{ID: 3}, {ID: 1},
	}

	w2 := &Way{Nodes: w.Nodes[:3]}
	if w2.Polygon() {
		t.Errorf("should be over 3 nodes to be polygon")
	}

	w.Nodes[3].ID = 10
	if w2.Polygon() {
		t.Errorf("first and last node must have same id")
	}

	c := polyConditions[1].Values
	if !reflect.DeepEqual(c, []string{"elevator", "escape", "rest_area", "services"}) {
		t.Errorf("values not sorted")
	}

	cases := []struct {
		name  string
		tags  []Tag
		value bool
	}{
		{
			name: "area no overrides",
			tags: []Tag{
				{Key: "area", Value: "no"},
				{Key: "building", Value: "yes"},
			},
			value: false,
		},
		{
			name: "non-empty area and not no",
			tags: []Tag{
				{Key: "area", Value: "maybe"},
				{Key: "building", Value: "no"},
			},
			value: true,
		},
		{
			name: "at least one condition met",
			tags: []Tag{
				{Key: "building", Value: "no"},
				{Key: "boundary", Value: "yes"},
			},
			value: true,
		},
		{
			name: "match within whitelist",
			tags: []Tag{
				{Key: "railway", Value: "station"},
			},
			value: true,
		},
		{
			name: "not match if not within whitelist",
			tags: []Tag{
				{Key: "railway", Value: "line"},
			},
			value: false,
		},
		{
			name: "not match within blacklist",
			tags: []Tag{
				{Key: "man_made", Value: "cutline"},
			},
			value: false,
		},
		{
			name: "match if not within blacklist",
			tags: []Tag{
				{Key: "man_made", Value: "thing"},
			},
			value: true,
		},
		{
			name: "indoor anything is a polygon",
			tags: []Tag{
				{Key: "indoor", Value: "anything"},
			},
			value: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := &Way{
				Nodes: []WayNode{
					{ID: 1}, {ID: 2},
					{ID: 3}, {ID: 1},
				},
				Tags: Tags(tc.tags),
			}

			if v := w.Polygon(); v != tc.value {
				t.Errorf("not correctly detected, %v != %v", v, tc.value)
			}
		})
	}
}
