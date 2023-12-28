package annotate

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/osm"
)

func TestRelation(t *testing.T) {
	ids := []osm.RelationID{
		2714790,
		4017808,
	}

	for _, id := range ids {
		o := loadTestdata(t, fmt.Sprintf("testdata/relation_%d.osm", id))

		ds := o.HistoryDatasource()
		for id, ways := range ds.Ways {
			err := Ways(context.Background(), ways, ds)
			if err != nil {
				t.Fatalf("compute error for way %d: %v", id, err)
			}
		}

		relations := ds.Relations[id]
		err := Relations(context.Background(), relations, ds, Threshold(30*time.Minute))
		if err != nil {
			t.Fatalf("compute error for %d: %v", id, err)
		}

		expected := loadTestdata(t, fmt.Sprintf("testdata/relation_%d_expected.osm", id))
		if !reflect.DeepEqual(relations, expected.Relations) {
			filename := fmt.Sprintf("testdata/relation_%d_got.osm", id)
			t.Errorf("expected relations not equal, file saved to %s", filename)

			data, _ := xml.MarshalIndent(&osm.OSM{Relations: relations}, "", " ")
			err := os.WriteFile(filename, data, 0644)
			if err != nil {
				t.Fatalf("write error: %v", err)
			}
		}
	}
}

func TestRelation_reverse(t *testing.T) {
	ways := osm.Ways{
		{
			ID: 1, Version: 1, Visible: true, Nodes: osm.WayNodes{
				{ID: 3, Lon: 3, Lat: 3},
				{ID: 2, Lon: 0, Lat: 3},
				{ID: 1, Lon: 0, Lat: 0},
			},
		},
		{
			ID: 1, Version: 2, Visible: true,
			Timestamp: time.Now().Add(-time.Hour),
			Nodes: osm.WayNodes{
				{ID: 1, Lon: 0, Lat: 0},
				{ID: 2, Lon: 0, Lat: 3},
				{ID: 3, Lon: 3, Lat: 3},
			},
		},
		{
			ID: 2, Version: 1, Visible: true, Nodes: osm.WayNodes{
				{ID: 3, Lon: 3, Lat: 3},
				{ID: 2, Lon: 0, Lat: 3},
				{ID: 1, Lon: 0, Lat: 0},
				{ID: 3, Lon: 3, Lat: 3},
			},
		},
		{
			ID: 2, Version: 2, Visible: true,
			Timestamp: time.Now().Add(-time.Hour),
			Nodes: osm.WayNodes{
				{ID: 3, Lon: 3, Lat: 3},
				{ID: 1, Lon: 0, Lat: 0},
				{ID: 2, Lon: 0, Lat: 3},
				{ID: 3, Lon: 3, Lat: 3},
			},
		},
	}

	t.Run("segment reverse", func(t *testing.T) {
		r := &osm.Relation{
			ID:      1,
			Visible: true,
			Tags:    osm.Tags{{Key: "type", Value: "multipolygon"}},
			Members: osm.Members{
				{Type: osm.TypeWay, Ref: 1, Role: "outer"},
			},
		}

		err := Relations(
			context.Background(),
			osm.Relations{r},
			(&osm.OSM{Ways: ways}).HistoryDatasource(),
			Threshold(time.Hour),
		)
		if err != nil {
			t.Fatalf("annotation error: %v", err)
		}

		if !r.Updates[0].Reverse {
			t.Errorf("incorrect reverse")
		}
	})

	t.Run("closed ring not a reverse", func(t *testing.T) {
		r := &osm.Relation{
			ID:      1,
			Visible: true,
			Tags:    osm.Tags{{Key: "type", Value: "multipolygon"}},
			Members: osm.Members{
				{Type: osm.TypeWay, Ref: 2, Role: "outer"},
			},
		}

		err := Relations(
			context.Background(),
			osm.Relations{r},
			(&osm.OSM{Ways: ways}).HistoryDatasource(),
			Threshold(time.Hour),
		)
		if err != nil {
			t.Fatalf("annotation error: %v", err)
		}

		if r.Updates[0].Reverse {
			t.Errorf("incorrect reverse")
		}
	})
}

func TestRelation_polygon(t *testing.T) {
	ways := osm.Ways{
		{
			ID:      1,
			Version: 1,
			Visible: true,
			Nodes: osm.WayNodes{
				{ID: 3, Lon: 3, Lat: 3},
				{ID: 2, Lon: 0, Lat: 3},
				{ID: 1, Lon: 0, Lat: 0},
			},
		},
		{
			ID:        1,
			Version:   2,
			Visible:   true,
			Timestamp: time.Now().Add(-time.Hour),
			Nodes: osm.WayNodes{
				{ID: 1, Lon: 0, Lat: 0},
				{ID: 2, Lon: 0, Lat: 3},
				{ID: 3, Lon: 3, Lat: 3},
			},
		},
		{
			ID:      2,
			Version: 1,
			Visible: true,
			Nodes: osm.WayNodes{
				{ID: 3, Lon: 3, Lat: 3},
				{ID: 4, Lon: 3, Lat: 0},
				{ID: 1, Lon: 0, Lat: 0},
			},
		},
		{
			ID:        2,
			Version:   2,
			Visible:   true,
			Timestamp: time.Now().Add(-time.Minute),
			Nodes: osm.WayNodes{
				{ID: 3, Lon: 3, Lat: 3},
				{ID: 4, Lon: 3, Lat: 0.1},
				{ID: 1, Lon: 0, Lat: 0},
			},
		},
		{
			ID:      3,
			Visible: true,
			Nodes: osm.WayNodes{
				{ID: 5, Lon: 1, Lat: 1},
				{ID: 6, Lon: 2, Lat: 1},
				{ID: 7, Lon: 2, Lat: 2},
			},
		},
		{
			ID:      4,
			Visible: true,
			Nodes: osm.WayNodes{
				{ID: 5, Lon: 1, Lat: 1},
				{ID: 8, Lon: 1, Lat: 2},
				{ID: 7, Lon: 2, Lat: 2},
			},
		},
	}
	r := &osm.Relation{
		ID:      1,
		Visible: true,
		Tags:    osm.Tags{{Key: "type", Value: "multipolygon"}},
		Members: osm.Members{
			{Type: osm.TypeWay, Ref: 1, Role: "outer"},
			{Type: osm.TypeWay, Ref: 2, Role: "outer"},
			{Type: osm.TypeWay, Ref: 3, Role: "inner"},
			{Type: osm.TypeWay, Ref: 4, Role: "inner"},
		},
	}

	if !r.Polygon() {
		t.Fatalf("test relation must be a polygon")
	}

	err := Relations(
		context.Background(),
		osm.Relations{r},
		(&osm.OSM{Ways: ways}).HistoryDatasource(),
		Threshold(time.Hour),
	)

	if err != nil {
		t.Fatalf("annotation error: %v", err)
	}

	expected := []orb.Orientation{orb.CCW, orb.CW, orb.CCW, orb.CW}
	for i, m := range r.Members {
		if m.Orientation != expected[i] {
			t.Errorf("member %d: %v != %v", i, m.Orientation, expected[i])
		}
	}

	if !r.Updates[0].Reverse {
		t.Errorf("incorrect reverse")
	}

	if r.Updates[1].Reverse {
		t.Errorf("incorrect reverse")
	}
}

func TestRelation_circular(t *testing.T) {
	relations := osm.Relations{
		&osm.Relation{ID: 1, Version: 1, Visible: true, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 2},
				{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 1, Version: 2, Visible: true, Timestamp: time.Date(2012, 1, 2, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 2},
				{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 1, Version: 3, Visible: true, Timestamp: time.Date(2012, 1, 3, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 2},
				{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 2, Version: 1, Visible: true, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 2, Version: 2, Visible: true, Timestamp: time.Date(2012, 1, 4, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 3, Version: 1, Visible: true, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 1},
			}},
		&osm.Relation{ID: 3, Version: 2, Visible: true, Timestamp: time.Date(2012, 1, 1, 10, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 1},
			}},
	}

	ds := (&osm.OSM{Relations: relations}).HistoryDatasource()
	rs := ds.Relations[1]
	err := Relations(context.Background(), rs, ds, Threshold(30*time.Minute))
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	// verify the members were annotated with the version numbers
	expected := osm.Members{
		{Type: osm.TypeRelation, Ref: 2, Version: 1},
		{Type: osm.TypeRelation, Ref: 3, Version: 1},
	}
	if !reflect.DeepEqual(rs[0].Members, expected) {
		t.Errorf("incorrect members: %v", rs[0].Members)
	}

	// should not have any updates
	if l := len(rs[0].Updates); l != 1 {
		t.Errorf("should have one update: %v", rs[0].Updates)
	}

	// version 2
	expected = osm.Members{
		{Type: osm.TypeRelation, Ref: 2, Version: 1},
		{Type: osm.TypeRelation, Ref: 3, Version: 2},
	}
	if !reflect.DeepEqual(rs[1].Members, expected) {
		t.Errorf("incorrect members: %v", rs[1].Members)
	}

	if l := len(rs[1].Updates); l != 0 {
		t.Errorf("should have no updates: %v", rs[1].Updates)
	}

	// version 3
	expected = osm.Members{
		{Type: osm.TypeRelation, Ref: 2, Version: 1},
		{Type: osm.TypeRelation, Ref: 3, Version: 2},
	}
	if !reflect.DeepEqual(rs[2].Members, expected) {
		t.Errorf("incorrect members: %v", rs[2].Members)
	}

	if l := len(rs[2].Updates); l != 1 {
		t.Errorf("should have one update: %v", rs[2].Updates)
	}
}

func TestRelation_selfCircular(t *testing.T) {
	rs := osm.Relations{
		{ID: 1, Version: 1, Visible: true, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 1},
			}},
		{ID: 1, Version: 2, Visible: true, Timestamp: time.Date(2012, 1, 2, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 1},
			}},
		{ID: 1, Version: 3, Visible: true, Timestamp: time.Date(2012, 1, 3, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				{Type: osm.TypeRelation, Ref: 1},
			}},
	}

	ds := (&osm.OSM{Relations: rs}).HistoryDatasource()
	err := Relations(context.Background(), rs, ds)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	// should not have any updates
	if l := len(rs[0].Updates); l != 0 {
		t.Errorf("should have no updates: %v", rs[0].Updates)
	}

	if v := rs[0].Members[0].Version; v != 1 {
		t.Errorf("member version not annotated: %v", v)
	}

	if l := len(rs[1].Updates); l != 0 {
		t.Errorf("should have no updates: %v", rs[1].Updates)
	}

	if v := rs[1].Members[0].Version; v != 2 {
		t.Errorf("member version not annotated: %v", v)
	}

	if l := len(rs[2].Updates); l != 0 {
		t.Errorf("should have no updates: %v", rs[2].Updates)
	}

	if v := rs[2].Members[0].Version; v != 3 {
		t.Errorf("member version not annotated: %v", v)
	}
}

func BenchmarkRelations(b *testing.B) {
	id := osm.RelationID(2714790)
	filename := fmt.Sprintf("testdata/relation_%d.osm", id)

	o := loadTestdata(b, filename)
	ds := o.HistoryDatasource()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := Relations(context.Background(), ds.Relations[id], ds)
		if err != nil {
			b.Fatalf("compute error: %v", err)
		}
	}
}
