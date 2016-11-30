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

func TestWays(t *testing.T) {
	ids := []osm.WayID{
		6394949,
		230391153,
	}

	for _, id := range ids {
		o := loadTestdata(t, fmt.Sprintf("testdata/way_%d.osm", id))

		ds := NewDatasource(o.Nodes, nil, nil)
		err := Ways(context.Background(), o.Ways, ds, 30*time.Minute)
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
			ioutil.WriteFile(filename, data, 0644)
		}
	}
}

func BenchmarkWay(b *testing.B) {
	o := loadTestdata(b, "testdata/way_6394949.osm")
	ds := NewDatasource(o.Nodes, nil, nil)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := Ways(context.Background(), o.Ways, ds, 30*time.Minute)
		if err != nil {
			b.Fatalf("compute error: %v", err)
		}
	}
}

func BenchmarkWays(b *testing.B) {
	o := loadTestdata(b, "testdata/relation_2714790.osm")
	ds := NewDatasource(o.Nodes, o.Ways, o.Relations)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; {
		for id, ways := range ds.Ways {
			err := Ways(context.Background(), ways, ds, 30*time.Minute)
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
	data, err := ioutil.ReadFile(filename)
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
