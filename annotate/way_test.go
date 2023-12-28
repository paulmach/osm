package annotate

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/paulmach/osm"
)

func TestWays(t *testing.T) {
	ids := []osm.WayID{
		6394949,
		230391153,
	}

	for _, id := range ids {
		o := loadTestdata(t, fmt.Sprintf("testdata/way_%d.osm", id))

		ds := (&osm.OSM{Nodes: o.Nodes}).HistoryDatasource()
		err := Ways(context.Background(), o.Ways, ds)
		if err != nil {
			t.Fatalf("compute error: %v", err)
		}

		filename := fmt.Sprintf("testdata/way_%d_expected.osm", id)
		expected := loadTestdata(t, filename)

		if !reflect.DeepEqual(o.Ways, expected.Ways) {
			t.Errorf("expected way for id %d not equal", id)

			filename := fmt.Sprintf("testdata/way_%d_got.osm", id)
			t.Errorf("expected way not equal, file saved to %s", filename)

			data, _ := xml.MarshalIndent(&osm.OSM{Ways: o.Ways}, "", " ")
			err := os.WriteFile(filename, data, 0644)
			if err != nil {
				t.Fatalf("write error: %v", err)
			}
		}
	}
}

func TestWays_childFilter(t *testing.T) {
	nodes := osm.Nodes{
		{ID: 1, Version: 1, Lat: 1, Lon: 1, Visible: true},
		{ID: 1, Version: 2, Lat: 2, Lon: 2, Visible: true},
		{ID: 1, Version: 3, Lat: 3, Lon: 3, Visible: true},
		{ID: 2, Version: 1, Lat: 1, Lon: 1, Visible: true},
		{ID: 2, Version: 2, Lat: 2, Lon: 2, Visible: true},
		{ID: 2, Version: 3, Lat: 3, Lon: 3, Visible: true},
		{ID: 3, Version: 1, Lat: 1, Lon: 1, Visible: true},
	}

	ways := osm.Ways{
		{
			ID:      1,
			Version: 1,
			Visible: true,
			Nodes: osm.WayNodes{
				{ID: 1, Version: 1},
				{ID: 2, Version: 1}, // filter says no annotate
				{ID: 3},             // annotate because not
			},
		},
	}

	ds := (&osm.OSM{Nodes: nodes}).HistoryDatasource()
	err := Ways(
		context.Background(),
		ways,
		ds,
		Threshold(0),
		ChildFilter(func(fid osm.FeatureID) bool {
			return fid == osm.NodeID(1).FeatureID()
		}),
	)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	if ways[0].Nodes[0].Lat == 0 {
		t.Errorf("should annotate first node")
	}

	if ways[0].Nodes[1].Lat != 0 {
		t.Errorf("should not annotate second node")
	}

	if ways[0].Nodes[2].Lat == 0 {
		t.Errorf("should annotate third node")
	}
}

func BenchmarkWay(b *testing.B) {
	o := loadTestdata(b, "testdata/way_6394949.osm")
	ds := (&osm.OSM{Nodes: o.Nodes}).HistoryDatasource()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := Ways(context.Background(), o.Ways, ds)
		if err != nil {
			b.Fatalf("compute error: %v", err)
		}
	}
}

func BenchmarkWays(b *testing.B) {
	o := loadTestdata(b, "testdata/relation_2714790.osm")
	ds := o.HistoryDatasource()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; {
		for id, ways := range ds.Ways {
			err := Ways(context.Background(), ways, ds)
			if err != nil {
				b.Fatalf("compute error for way %d: %v", id, err)
			}

			n++
			if n >= b.N {
				break
			}
		}
	}
}

func loadTestdata(tb testing.TB, filename string) *osm.OSM {
	data, err := os.ReadFile(filename)
	if err != nil {
		tb.Fatalf("unable to open file: %v", err)
	}

	o := &osm.OSM{}
	err = xml.Unmarshal(data, o)
	if err != nil {
		tb.Fatalf("unable to unmarshal data: %v", err)
	}

	return o
}
