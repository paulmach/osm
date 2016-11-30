package augment

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
			t.Fatalf("compute error: %v", err)
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
