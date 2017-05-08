package osmgeojson

import (
	"encoding/xml"
	"io/ioutil"
	"testing"

	osm "github.com/paulmach/go.osm"
)

func BenchmarkConvert(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")

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
		Convert(o, NoID)
	}
}

func BenchmarkConvert_NoMeta(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Convert(o, NoMeta)
	}
}

func BenchmarkConvert_NoRelationMembership(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Convert(o, NoRelationMembership)
	}
}

func BenchmarkConvert_NoIDsMetaMembership(b *testing.B) {
	o := parseFile(b, "testdata/benchmark.osm")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Convert(o, NoID, NoMeta, NoRelationMembership)
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
