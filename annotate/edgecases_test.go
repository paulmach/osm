package annotate

import (
	"context"
	"testing"
	"time"

	"github.com/onXmaps/osm"
)

func TestEdgeCase_childCreatedAfterParent(t *testing.T) {
	// example: way 43680701, node 370250076
	//          way 4708608, node 29974559
	// Way's first version is 4 days after node's first version.
	// I think this is an artifact of the no-versions early history of OSM.

	t.Run("1 node after 1 way", func(t *testing.T) {
		nodes := osm.Nodes{
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 2},
		}

		ways := osm.Ways{
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
		}

		ds := (&osm.OSM{Nodes: nodes}).HistoryDatasource()
		err := Ways(context.Background(), ways, ds, IgnoreInconsistency(true))
		if err != nil {
			t.Fatalf("compute error: %v", err)
		}

		// should not annotate the way, but add an update when the node comes online.

		node := ways[0].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}

		if l := len(ways[0].Updates); l != 1 {
			t.Fatalf("way should have 1 update: %d", l)
		}

		update := ways[0].Updates[0]
		if update.Lat != 1 || update.Lon != 2 && !update.Timestamp.Equal(nodes[0].Timestamp) {
			t.Errorf("should have update: %v", update)
		}
	})

	t.Run("2 ways around 1 node", func(t *testing.T) {
		nodes := osm.Nodes{
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 2},
		}

		ways := osm.Ways{
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
			{ID: 1, Visible: true, Version: 2, Timestamp: time.Date(2012, 3, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
		}

		ds := (&osm.OSM{Nodes: nodes}).HistoryDatasource()
		err := Ways(context.Background(), ways, ds, IgnoreInconsistency(true))
		if err != nil {
			t.Fatalf("compute error: %v", err)
		}

		// should not annotate the way, but add an update when the node comes online.

		node := ways[0].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}

		if l := len(ways[0].Updates); l != 1 {
			t.Fatalf("way should have 1 update: %d", l)
		}

		update := ways[0].Updates[0]
		if update.Lat != 1 || update.Lon != 2 && !update.Timestamp.Equal(nodes[0].Timestamp) {
			t.Errorf("should have update: %v", update)
		}

		if l := len(ways[1].Updates); l != 0 {
			t.Fatalf("way should have 0 updates: %d", l)
		}

		// should annotate second way just fine.
		node = ways[1].Nodes[0]
		if node.Lat != 1 || node.Lon != 2 {
			t.Errorf("should annotate node in second way: %v", node)
		}
	})

	t.Run("2 node between 2 ways", func(t *testing.T) {
		nodes := osm.Nodes{
			{ID: 1, Visible: false, Version: 1, Timestamp: time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 2},
			{ID: 1, Visible: true, Version: 2, Timestamp: time.Date(2012, 4, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 3},
		}

		ways := osm.Ways{
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
			{ID: 1, Visible: true, Version: 2, Timestamp: time.Date(2012, 3, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
		}

		ds := (&osm.OSM{Nodes: nodes}).HistoryDatasource()
		err := Ways(context.Background(), ways, ds, IgnoreInconsistency(true))
		if err != nil {
			t.Fatalf("compute error: %v", err)
		}

		// should not annotate the way, but add an update when the node comes online.

		node := ways[0].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}

		if l := len(ways[0].Updates); l != 0 {
			t.Fatalf("way should have 0 updates: %d", l)
		}

		// should annotate second way just fine.
		node = ways[1].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}

		if l := len(ways[1].Updates); l != 1 {
			t.Fatalf("way should have 1 update: %d", l)
		}

		update := ways[1].Updates[0]
		if update.Version != 2 {
			t.Errorf("incorrect update: %v", update)
		}
	})
}

func TestEdgeCase_nodeDeletedBetweenParents(t *testing.T) {
	// example: node 321452894, way 28831147

	nodes := osm.Nodes{
		{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 2},
		{ID: 1, Visible: false, Version: 2, Timestamp: time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 3},
		{ID: 1, Visible: true, Version: 3, Timestamp: time.Date(2012, 3, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 4},
		{ID: 1, Visible: true, Version: 4, Timestamp: time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 5},
		{ID: 1, Visible: true, Version: 5, Timestamp: time.Date(2013, 2, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 6},
	}

	ways := osm.Ways{
		{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Nodes: osm.WayNodes{{ID: 1}}},
		{ID: 1, Visible: true, Version: 2, Timestamp: time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC),
			Nodes: osm.WayNodes{{ID: 1}}},
	}

	ds := (&osm.OSM{Nodes: nodes}).HistoryDatasource()
	err := Ways(context.Background(), ways, ds, IgnoreInconsistency(true))
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	if l := len(ways[0].Updates); l != 1 {
		t.Fatalf("first way should have 1 update: %d", l)
	}

	update := ways[0].Updates[0]
	if update.Version != 3 {
		t.Errorf("incorrect update: %v", update)
	}

	if l := len(ways[1].Updates); l != 1 {
		t.Fatalf("second way should have 1 update: %d", l)
	}

	update = ways[1].Updates[0]
	if update.Version != 5 {
		t.Errorf("incorrect update: %v", update)
	}
}

func TestEdgeCase_nodeRedacted(t *testing.T) {
	// example: way 159081205, node 376130526
	// Oh the license change. Nodes have 1 non-visible version.

	t.Run("1 node redacted before both", func(t *testing.T) {
		nodes := osm.Nodes{
			{ID: 1, Visible: false, Version: 1, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 2},
		}

		ways := osm.Ways{
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
			{ID: 1, Visible: true, Version: 2, Timestamp: time.Date(2012, 3, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
		}

		ds := (&osm.OSM{Nodes: nodes}).HistoryDatasource()
		err := Ways(context.Background(), ways, ds, IgnoreInconsistency(true))
		if err != nil {
			t.Fatalf("compute error: %v", err)
		}

		// should not annotate

		node := ways[0].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}

		node = ways[1].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}
	})

	t.Run("1 node redacted between", func(t *testing.T) {
		nodes := osm.Nodes{
			{ID: 1, Visible: false, Version: 1, Timestamp: time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 2},
		}

		ways := osm.Ways{
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
			{ID: 1, Visible: true, Version: 2, Timestamp: time.Date(2012, 3, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
		}

		ds := (&osm.OSM{Nodes: nodes}).HistoryDatasource()
		err := Ways(context.Background(), ways, ds, IgnoreInconsistency(true))
		if err != nil {
			t.Fatalf("compute error: %v", err)
		}

		// should not annotate

		node := ways[0].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}

		node = ways[1].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}
	})

	t.Run("1 node redacted after", func(t *testing.T) {
		nodes := osm.Nodes{
			{ID: 1, Visible: false, Version: 1, Timestamp: time.Date(2012, 3, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 2},
		}

		ways := osm.Ways{
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
			{ID: 1, Visible: true, Version: 2, Timestamp: time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}}},
		}

		ds := (&osm.OSM{Nodes: nodes}).HistoryDatasource()
		err := Ways(context.Background(), ways, ds, IgnoreInconsistency(true))
		if err != nil {
			t.Fatalf("compute error: %v", err)
		}

		// should not annotate

		node := ways[0].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}

		node = ways[1].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}
	})

	t.Run("2 nodes", func(t *testing.T) {
		nodes := osm.Nodes{
			{ID: 1, Visible: false, Version: 1, Timestamp: time.Date(2012, 4, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 2},
			{ID: 2, Visible: true, Version: 1, Timestamp: time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC), Lat: 1, Lon: 2},
		}

		ways := osm.Ways{
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 3, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}, {ID: 2}}},
			{ID: 1, Visible: true, Version: 1, Timestamp: time.Date(2012, 5, 1, 0, 0, 0, 0, time.UTC),
				Nodes: osm.WayNodes{{ID: 1}, {ID: 2}}},
		}

		ds := (&osm.OSM{Nodes: nodes}).HistoryDatasource()
		err := Ways(context.Background(), ways, ds, IgnoreInconsistency(true))
		if err != nil {
			t.Fatalf("compute error: %v", err)
		}

		// should not annotate the redacted node, otherwise okay.

		node := ways[0].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}

		node = ways[1].Nodes[0]
		if node.Lat != 0 || node.Lon != 0 {
			t.Errorf("should not annotate node: %v", node)
		}

		node = ways[0].Nodes[1]
		if node.Lat != 1 || node.Lon != 2 {
			t.Errorf("should annotate second node: %v", node)
		}

		node = ways[1].Nodes[1]
		if node.Lat != 1 || node.Lon != 2 {
			t.Errorf("should annotate second node: %v", node)
		}
	})
}
