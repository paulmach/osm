package annotate

import (
	"context"
	"testing"

	"github.com/onXmaps/osm"
)

func TestChildFirstOrdering(t *testing.T) {
	relations := osm.Relations{
		{ID: 8, Members: osm.Members{{Type: osm.TypeNode, Ref: 12}}},
		{ID: 10, Members: osm.Members{
			{Type: osm.TypeWay, Ref: 8},
			{Type: osm.TypeRelation, Ref: 8},
		}},
		{ID: 12, Members: osm.Members{{Type: osm.TypeRelation, Ref: 10}}},
		{ID: 14, Members: osm.Members{{Type: osm.TypeNode, Ref: 12}}},
	}

	ordering := NewChildFirstOrdering(
		context.Background(),
		relations.IDs(),
		(&osm.OSM{Relations: relations}).HistoryDatasource())

	ids := make([]osm.RelationID, 0, len(relations))
	for ordering.Next() {
		ids = append(ids, ordering.RelationID())
	}

	if len(ids) != len(relations) {
		t.Errorf("wrong number of ids, %v != %v", len(ids), len(relations))
	}

	aheadOf := [][2]osm.RelationID{
		{8, 10},  // 8 ahead of, or before 10
		{10, 12}, // 10 before 12
	}

	for i, p := range aheadOf {
		if indexOf(ids, p[0]) > indexOf(ids, p[1]) {
			t.Errorf("incorrect ordering, test %v", i)
			t.Logf("ids: %v", ids)
		}
	}

	if err := ordering.Err(); err != nil {
		t.Errorf("unexpected error, got %v", err)
	}
}

func TestChildFirstOrdering_cycle(t *testing.T) {
	relations := osm.Relations{
		{ID: 1, Version: 1, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 2},
			{Type: osm.TypeRelation, Ref: 3},
		}},
		{ID: 1, Version: 2, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 2},
			{Type: osm.TypeRelation, Ref: 3},
			{Type: osm.TypeRelation, Ref: 5},
		}},
		{ID: 2, Version: 1, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 4},
			{Type: osm.TypeRelation, Ref: 1},
		}},
		{ID: 2, Version: 2, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 6},
		}},
		{ID: 3, Members: osm.Members{{Type: osm.TypeWay, Ref: 8}}},
		{ID: 4, Members: osm.Members{{Type: osm.TypeWay, Ref: 8}}},
		{ID: 5, Members: osm.Members{{Type: osm.TypeWay, Ref: 8}}},
		{ID: 6, Members: osm.Members{{Type: osm.TypeWay, Ref: 8}}},

		// self cycle
		{ID: 9, Members: osm.Members{{Type: osm.TypeRelation, Ref: 9}}},
	}

	ds := (&osm.OSM{Relations: relations}).HistoryDatasource()
	ordering := NewChildFirstOrdering(context.Background(), relations.IDs(), ds)

	ids := make([]osm.RelationID, 0, len(relations))
	for ordering.Next() {
		ids = append(ids, ordering.RelationID())
	}

	if len(ids) != len(ds.Relations) {
		t.Errorf("wrong number of ids, %v != %v", len(ids), len(relations))
	}

	aheadOf := [][2]osm.RelationID{
		{3, 1}, // 3 ahead of, or before 1
		{4, 2}, // 4 before 2
		{6, 2},
		{5, 1},
	}

	for i, p := range aheadOf {
		if indexOf(ids, p[0]) > indexOf(ids, p[1]) {
			t.Errorf("incorrect ordering, test %v", i)
			t.Logf("ids: %v", ids)
		}
	}

	if err := ordering.Err(); err != nil {
		t.Errorf("unexpected error, got %v", err)
	}
}

func TestChildFirstOrdering_Cancel(t *testing.T) {
	relations := osm.Relations{
		{ID: 8, Members: osm.Members{{Type: osm.TypeNode, Ref: 12}}},
		{ID: 10, Members: osm.Members{{Type: osm.TypeRelation, Ref: 8}}},
		{ID: 12, Members: osm.Members{{Type: osm.TypeRelation, Ref: 10}}},
	}

	ctx, done := context.WithCancel(context.Background())
	ordering := NewChildFirstOrdering(
		ctx,
		relations.IDs(),
		(&osm.OSM{Relations: relations}).HistoryDatasource())

	ordering.Next()
	ordering.Next()
	done()

	if ordering.Next() {
		t.Errorf("expect scan after cancel to be false")
	}

	if err := ordering.Err(); err != context.Canceled {
		t.Errorf("incorrect error, got %v", err)
	}
}

func TestChildFirstOrdering_Close(t *testing.T) {
	relations := osm.Relations{
		{ID: 8, Members: osm.Members{{Type: osm.TypeNode, Ref: 12}}},
		{ID: 10, Members: osm.Members{{Type: osm.TypeRelation, Ref: 8}}},
		{ID: 12, Members: osm.Members{{Type: osm.TypeRelation, Ref: 10}}},
	}

	ordering := NewChildFirstOrdering(
		context.Background(),
		relations.IDs(),
		(&osm.OSM{Relations: relations}).HistoryDatasource())

	ordering.Next()
	ordering.Next()
	ordering.Close()

	if ordering.Next() {
		t.Errorf("expect scan after cancel to be false")
	}

	if err := ordering.Err(); err != context.Canceled {
		t.Errorf("incorrect error, got %v", err)
	}
}

func TestChildFirstOrdering_Walk(t *testing.T) {
	relations := osm.Relations{
		{ID: 2, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 4},
		}},
		{ID: 4, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 6},
		}},
		{ID: 6, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 10},
		}},
		{ID: 8, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 6},
		}},
		{ID: 10, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 8},
		}},

		// circular relation of self.
		{ID: 16, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 16},
		}},
	}

	ordering := &ChildFirstOrdering{
		ctx:     context.Background(),
		ds:      (&osm.OSM{Relations: relations}).HistoryDatasource(),
		visited: make(map[osm.RelationID]struct{}, len(relations)),
		out:     make(chan osm.RelationID, 10+len(relations)),
	}

	// start at all parts of cycle
	// basically should not infinite loop
	path := make([]osm.RelationID, 0, 100)
	for _, r := range relations {
		err := ordering.walk(r.ID, path)
		if err != nil {
			t.Errorf("should process cycle without problem: %v", err)
		}
	}
}

func TestChildFirstOrdering_missingRelation(t *testing.T) {
	relations := osm.Relations{
		{ID: 2, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 3},
		}},
		{ID: 4, Members: osm.Members{
			{Type: osm.TypeRelation, Ref: 2},
		}},
	}

	ordering := &ChildFirstOrdering{
		ctx:     context.Background(),
		ds:      (&osm.OSM{Relations: relations}).HistoryDatasource(),
		visited: make(map[osm.RelationID]struct{}, len(relations)),
		out:     make(chan osm.RelationID, 10+len(relations)),
	}

	// start at all parts of cycle
	// basically should not infinite loop
	path := make([]osm.RelationID, 0, 100)
	for _, r := range relations {
		err := ordering.walk(r.ID, path)
		if err != nil {
			t.Errorf("should process cycle without problem: %v", err)
		}
	}
}

func indexOf(s []osm.RelationID, id osm.RelationID) int {
	for i, sid := range s {
		if sid == id {
			return i
		}
	}

	panic("id not found")
}
