package osmgeojson

import (
	"encoding/xml"
	"io/ioutil"
	"testing"

	"github.com/onXmaps/osm"
)

func BenchmarkConvert(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Convert(o)
	}
}

func BenchmarkConvertAnnotated(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")
	annotate(o)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Convert(o)
	}
}

func BenchmarkConvert_NoID(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Convert(o, NoID(true))
	}
}

func BenchmarkConvert_NoMeta(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Convert(o, NoMeta(true))
	}
}

func BenchmarkConvert_NoRelationMembership(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Convert(o, NoRelationMembership(true))
	}
}

func BenchmarkConvert_NoIDsMetaMembership(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Convert(o, NoID(true), NoMeta(true), NoRelationMembership(true))
	}
}

func parseFile(t testing.TB, filename string) *osm.OSM {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	o := &osm.OSM{}
	err = xml.Unmarshal(data, &o)
	if err != nil {
		t.Fatalf("failed to unmarshal %s: %v", filename, err)
	}

	return o
}

func annotate(o *osm.OSM) {
	nodes := make(map[osm.NodeID]*osm.Node)
	for _, n := range o.Nodes {
		nodes[n.ID] = n
	}

	for _, w := range o.Ways {
		for i, wn := range w.Nodes {
			n := nodes[wn.ID]
			if n == nil {
				continue
			}

			w.Nodes[i].Lat = n.Lat
			w.Nodes[i].Lon = n.Lon
			w.Nodes[i].Version = n.Version
		}
	}
}
