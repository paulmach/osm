package osm

import (
	"context"
	"testing"
)

func TestHistoryDatasource(t *testing.T) {
	ctx := context.Background()

	t.Run("empty datasource", func(t *testing.T) {
		ds := &HistoryDatasource{}

		if _, err := ds.NodeHistory(ctx, 1); !ds.NotFound(err) {
			t.Errorf("should be not found error: %v", err)
		}

		if _, err := ds.WayHistory(ctx, 1); !ds.NotFound(err) {
			t.Errorf("should be not found error: %v", err)
		}

		if _, err := ds.RelationHistory(ctx, 1); !ds.NotFound(err) {
			t.Errorf("should be not found error: %v", err)
		}
	})

	o := &OSM{
		Nodes: Nodes{
			{ID: 1, Version: 1},
			{ID: 1, Version: 2},
		},
		Ways: Ways{
			{ID: 1, Version: 1},
			{ID: 1, Version: 2},
			{ID: 1, Version: 3},
		},
		Relations: Relations{
			{ID: 1, Version: 1},
			{ID: 1, Version: 2},
			{ID: 1, Version: 3},
			{ID: 1, Version: 4},
		},
	}

	t.Run("non-empty datasource", func(t *testing.T) {
		ds := o.HistoryDatasource()

		ns, err := ds.NodeHistory(ctx, 1)
		if err != nil {
			t.Errorf("should not return error: %v", err)
		}

		if len(ns) != 2 {
			t.Errorf("incorrect nodes: %v", ns)
		}

		ws, err := ds.WayHistory(ctx, 1)
		if err != nil {
			t.Errorf("should not return error: %v", err)
		}

		if len(ws) != 3 {
			t.Errorf("incorrect ways: %v", ns)
		}

		rs, err := ds.RelationHistory(ctx, 1)
		if err != nil {
			t.Errorf("should not return error: %v", err)
		}

		if len(rs) != 4 {
			t.Errorf("incorrect relations: %v", ns)
		}
	})

	t.Run("not found non-empty datasource", func(t *testing.T) {
		ds := o.HistoryDatasource()

		if _, err := ds.NodeHistory(ctx, 2); !ds.NotFound(err) {
			t.Errorf("should be not found error: %v", err)
		}

		if _, err := ds.WayHistory(ctx, 2); !ds.NotFound(err) {
			t.Errorf("should be not found error: %v", err)
		}

		if _, err := ds.RelationHistory(ctx, 2); !ds.NotFound(err) {
			t.Errorf("should be not found error: %v", err)
		}
	})

}
