package annotate

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	osm "github.com/paulmach/go.osm"
)

func TestRelation(t *testing.T) {
	ids := []osm.RelationID{
		2714790,
		4017808,
	}

	for _, id := range ids {
		o := loadTestdata(t, fmt.Sprintf("testdata/relation_%d.osm", id))

		ds := NewDatasource(o.Nodes, o.Ways, o.Relations)
		for id, ways := range ds.Ways {
			err := Ways(context.Background(), ways, ds, 30*time.Minute)
			if err != nil {
				t.Fatalf("compute error for way %d: %v", id, err)
			}
		}

		relations := ds.Relations[id]
		err := Relations(context.Background(), relations, ds, 30*time.Minute)
		if err != nil {
			t.Fatalf("compute error for %d: %v", id, err)
		}

		expected := loadTestdata(t, fmt.Sprintf("testdata/relation_%d_expected.osm", id))
		if !reflect.DeepEqual(relations, expected.Relations) {
			filename := fmt.Sprintf("testdata/relation_%d_got.osm", id)
			t.Errorf("expected relations not equal, file saved to %s", filename)

			data, _ := xml.MarshalIndent(&osm.OSM{Relations: relations}, "", " ")
			ioutil.WriteFile(filename, data, 0644)
		}
	}
}

func TestRelationCircular(t *testing.T) {
	relations := osm.Relations{
		&osm.Relation{ID: 1, Version: 1, Visible: true, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 2},
				osm.Member{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 1, Version: 2, Visible: true, Timestamp: time.Date(2012, 1, 2, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 2},
				osm.Member{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 1, Version: 3, Visible: true, Timestamp: time.Date(2012, 1, 3, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 2},
				osm.Member{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 2, Version: 1, Visible: true, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 2, Version: 2, Visible: true, Timestamp: time.Date(2012, 1, 4, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 3},
			}},
		&osm.Relation{ID: 3, Version: 1, Visible: true, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 1},
			}},
		&osm.Relation{ID: 3, Version: 2, Visible: true, Timestamp: time.Date(2012, 1, 1, 10, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 1},
			}},
	}

	ds := NewDatasource(nil, nil, relations)
	rs := ds.Relations[1]
	err := Relations(context.Background(), rs, ds, 30*time.Minute)
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

func TestRelationSelfCircular(t *testing.T) {
	rs := osm.Relations{
		&osm.Relation{ID: 1, Version: 1, Visible: true, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 1},
			}},
		&osm.Relation{ID: 1, Version: 2, Visible: true, Timestamp: time.Date(2012, 1, 2, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 1},
			}},
		&osm.Relation{ID: 1, Version: 3, Visible: true, Timestamp: time.Date(2012, 1, 3, 0, 0, 0, 0, time.UTC),
			Members: osm.Members{
				osm.Member{Type: osm.TypeRelation, Ref: 1},
			}},
	}

	ds := NewDatasource(nil, nil, rs)
	err := Relations(context.Background(), rs, ds, 30*time.Minute)
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

func BenchmarkRelation(b *testing.B) {
	id := osm.RelationID(2714790)
	filename := fmt.Sprintf("testdata/relation_%d.osm", id)

	o := loadTestdata(b, filename)
	ds := NewDatasource(o.Nodes, o.Ways, o.Relations)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := Relations(context.Background(), ds.Relations[id], ds, 30*time.Minute)
		if err != nil {
			b.Fatalf("compute error: %v", err)
		}
	}
}
