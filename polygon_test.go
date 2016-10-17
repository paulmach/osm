package osm

import (
	"reflect"
	"testing"
)

func TestWayPolygon(t *testing.T) {
	w := &Way{}
	w.Nodes = []WayNode{
		WayNode{ID: 1}, WayNode{ID: 2},
		WayNode{ID: 3}, WayNode{ID: 1},
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
		Name  string
		Tags  []Tag
		Value bool
	}{
		{
			Name: "area no overrides",
			Tags: []Tag{
				{Key: "area", Value: "no"},
				{Key: "building", Value: "yes"},
			},
			Value: false,
		},
		{
			Name: "at least one condition met",
			Tags: []Tag{
				{Key: "building", Value: "no"},
				{Key: "boundary", Value: "yes"},
			},
			Value: true,
		},
		{
			Name: "match within whitelist",
			Tags: []Tag{
				{Key: "railway", Value: "station"},
			},
			Value: true,
		},
		{
			Name: "not match if not within whitelist",
			Tags: []Tag{
				{Key: "railway", Value: "line"},
			},
			Value: false,
		},
		{
			Name: "not match within blacklist",
			Tags: []Tag{
				{Key: "man_made", Value: "cutline"},
			},
			Value: false,
		},
		{
			Name: "match if not within blacklist",
			Tags: []Tag{
				{Key: "man_made", Value: "thing"},
			},
			Value: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			w := &Way{
				Nodes: []WayNode{
					WayNode{ID: 1}, WayNode{ID: 2},
					WayNode{ID: 3}, WayNode{ID: 1},
				},
				Tags: Tags(tc.Tags),
			}

			if v := w.Polygon(); v != tc.Value {
				t.Errorf("not correctly detected, %v != %v", v, tc.Value)
			}
		})
	}

}
