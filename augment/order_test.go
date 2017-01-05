package augment

import (
	"context"
	"testing"

	osm "github.com/paulmach/go.osm"
)

func TestDeptFirstOrdering(t *testing.T) {
	relations := map[osm.RelationID]osm.Relations{
		8: {
			{Members: osm.Members{{Type: osm.NodeType, Ref: 12}}},
		},
		10: {
			{Members: osm.Members{
				{Type: osm.WayType, Ref: 8},
				{Type: osm.RelationType, Ref: 8},
			}},
		},
		12: {
			{Members: osm.Members{{Type: osm.RelationType, Ref: 10}}},
		},
		14: {
			{Members: osm.Members{{Type: osm.NodeType, Ref: 12}}},
		},
	}
	ids := make([]osm.RelationID, 0, len(relations))
	for k := range relations {
		ids = append(ids, k)
	}

	ordering := NewChildFirstOrdering(context.Background(),
		ids, &MapDatasource{Relations: relations})

	ids = make([]osm.RelationID, 0, len(relations))
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

func TestDeptFirtOrderingCancel(t *testing.T) {
	relations := map[osm.RelationID]osm.Relations{
		8: {
			{Members: osm.Members{{Type: osm.NodeType, Ref: 12}}},
		},
		10: {
			{Members: osm.Members{{Type: osm.RelationType, Ref: 8}}},
		},
		12: {
			{Members: osm.Members{{Type: osm.RelationType, Ref: 10}}},
		},
	}
	ids := make([]osm.RelationID, 0, len(relations))
	for k := range relations {
		ids = append(ids, k)
	}

	ctx, done := context.WithCancel(context.Background())
	ordering := NewChildFirstOrdering(ctx,
		ids, &MapDatasource{Relations: relations})

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

func TestDeptFirtOrderingClose(t *testing.T) {
	relations := map[osm.RelationID]osm.Relations{
		8: {
			{Members: osm.Members{{Type: osm.NodeType, Ref: 12}}},
		},
		10: {
			{Members: osm.Members{{Type: osm.RelationType, Ref: 8}}},
		},
		12: {
			{Members: osm.Members{{Type: osm.RelationType, Ref: 10}}},
		},
	}
	ids := make([]osm.RelationID, 0, len(relations))
	for k := range relations {
		ids = append(ids, k)
	}

	ordering := NewChildFirstOrdering(context.Background(),
		ids, &MapDatasource{Relations: relations})

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
func TestDeptFirstOrderingWalk(t *testing.T) {
	relations := map[osm.RelationID]osm.Relations{
		2: {
			{Members: osm.Members{
				{Type: osm.RelationType, Ref: 4},
			}},
		},
		4: {
			{Members: osm.Members{
				{Type: osm.RelationType, Ref: 6},
			}},
		},
		6: {
			{Members: osm.Members{
				{Type: osm.RelationType, Ref: 10},
			}},
		},
		8: {
			{Members: osm.Members{
				{Type: osm.RelationType, Ref: 6},
			}},
		},
		10: {
			{Members: osm.Members{
				{Type: osm.RelationType, Ref: 8},
			}},
		},
	}

	ordering := &ChildFirstOrdering{
		ctx:     context.Background(),
		ds:      &MapDatasource{Relations: relations},
		visited: make(map[osm.RelationID]struct{}, len(relations)),
		out:     make(chan osm.RelationID, 10+len(relations)),
	}

	// start at all parts of cycle
	path := make([]osm.RelationID, 0, 100)
	for k := range relations {
		err := ordering.walk(k, path)
		if err == nil {
			t.Errorf("should get cycle error starting at %v", k)
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
