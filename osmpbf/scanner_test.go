package osmpbf

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/paulmach/osm"
)

var (
	Delaware = "../testdata/delaware-latest.osm.pbf"
)

func TestScanner(t *testing.T) {
	f, err := os.Open(Delaware)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	scanner := New(context.Background(), f, 1)
	defer scanner.Close()

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if n := scanner.Object().(*osm.Node); n.ID != 75385503 {
		t.Fatalf("did not scan correctly, got %v", n)
	}

	if v := scanner.Scan(); !v {
		t.Fatalf("should read second scan: %v", scanner.Err())
	}

	if n := scanner.Object().(*osm.Node); n.ID != 75390099 {
		t.Fatalf("did not scan correctly, got %v", n)
	}
}

func TestScanner_intermediateStart(t *testing.T) {
	f, err := os.Open(Delaware)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	scanner := New(context.Background(), f, 1)

	target := osm.NodeID(178592359) // first object in last partially scanned block
	indexOfTarget := 0
	for i := 0; i < 30000; i++ {
		scanner.Scan()
		if scanner.Object().(*osm.Node).ID == target {
			indexOfTarget = i
		}
	}

	// verifies the target is less than 1 block length from the end.
	if 30000-indexOfTarget > 8000 {
		t.Errorf("target is not near the end, index %v", indexOfTarget)
	}
	scanner.Close()

	// move the file pointer past all the fully scanned bytes,
	// to the start of the not-fully scanned block.
	f.Seek(scanner.FullyScannedBytes(), 0)
	scanner = New(context.Background(), f, 1)

	scanner.Scan()
	if id := scanner.Object().(*osm.Node).ID; id != target {
		t.Errorf("incorrect first id, got %v", id)
	}
	scanner.Close()
}

func TestScanner_context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	f, err := os.Open(Delaware)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	scanner := New(ctx, f, 1)
	defer scanner.Close()

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if n := scanner.Object().(*osm.Node); n.ID != 75385503 {
		t.Fatalf("did not scan correctly, got %v", n)
	}

	cancel()

	if v := scanner.Scan(); v {
		t.Fatalf("should be closed for second scan: %v", scanner.Err())
	}

	if v := scanner.Err(); v != ctx.Err() {
		t.Errorf("incorrect error, got %v", v)
	}
}

func TestScanner_Header(t *testing.T) {
	f, err := os.Open(Delaware)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	scanner := New(context.Background(), f, 1)

	header, err := scanner.Header()
	if err != nil {
		t.Fatalf("error reading header: %v", err)
	}

	expected := &osm.Bounds{
		MinLat: 38.450430000000004,
		MaxLat: 40.03221,
		MinLon: -75.78974000000001,
		MaxLon: -74.96121000000001,
	}
	if !reflect.DeepEqual(header.Bounds, expected) {
		t.Errorf("incorrect bounds: %v", header.Bounds)
	}

	if !reflect.DeepEqual(header.RequiredFeatures, []string{"OsmSchema-V0.6", "DenseNodes"}) {
		t.Errorf("incorrect required features: %v", header.RequiredFeatures)
	}

	if !reflect.DeepEqual(header.WritingProgram, "Osmium (http://wiki.openstreetmap.org/wiki/Osmium)") {
		t.Errorf("incorrect writing program: %v", header.WritingProgram)
	}

	if !reflect.DeepEqual(header.ReplicationTimestamp, time.Date(2016, 8, 10, 19, 28, 3, 0, time.UTC)) {
		t.Errorf("incorrect timestamp: %v", header.ReplicationTimestamp)
	}
}

func TestScanner_Close(t *testing.T) {
	f, err := os.Open(Delaware)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	scanner := New(context.Background(), f, 1)

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if n := scanner.Object().(*osm.Node); n.ID != 75385503 {
		t.Fatalf("did not scan correctly, got %v", n)
	}

	scanner.Close()

	if v := scanner.Scan(); v {
		t.Fatalf("should be closed for second scan: %v", scanner.Err())
	}

	if v := scanner.Err(); v != osm.ErrScannerClosed {
		t.Errorf("incorrect error, got %v", v)
	}
}

func TestScanner_FullyScannedBytes(t *testing.T) {
	t.Run("should be non-zero after reading whole file", func(t *testing.T) {
		f, err := os.Open(Delaware)
		if err != nil {
			t.Fatalf("unable to open file: %v", err)
		}
		defer f.Close()

		scanner := New(context.Background(), f, 1)
		for i := 0; i < 30000; i++ {
			scanner.Scan()
		}

		if v := scanner.FullyScannedBytes(); v != 214162 {
			t.Errorf("incorrect scanned bytes: %v", v)
		}

		for scanner.Scan() {
			// scan the whole thing
		}

		if v := scanner.FullyScannedBytes(); v != 7488871 {
			t.Errorf("incorrect scanned bytes: %v", v)
		}
	})

	t.Run("should be non-zero if cancel context", func(t *testing.T) {
		f, err := os.Open(Delaware)
		if err != nil {
			t.Fatalf("unable to open file: %v", err)
		}
		defer f.Close()

		scanner := New(context.Background(), f, 1)
		for i := 0; i < 30000; i++ {
			scanner.Scan()
		}

		if v := scanner.FullyScannedBytes(); v != 214162 {
			t.Errorf("incorrect scanned bytes: %v", v)
		}

		for scanner.Scan() {
			// scan the whole thing
		}

		if v := scanner.FullyScannedBytes(); v != 7488871 {
			t.Errorf("incorrect scanned bytes: %v", v)
		}
	})

	t.Run("should always be increasing", func(t *testing.T) {
		f, err := os.Open(Delaware)
		if err != nil {
			t.Fatalf("unable to open file: %v", err)
		}
		defer f.Close()

		scanner := New(context.Background(), f, 2)

		var previouslyScanned int64
		for scanner.Scan() {
			if v := scanner.FullyScannedBytes(); v < previouslyScanned {
				t.Errorf("scanned bytes decreased: %v < %v", v, previouslyScanned)
			}

			previouslyScanned = scanner.FullyScannedBytes()
		}

		if v := scanner.FullyScannedBytes(); v != 7488871 {
			t.Errorf("incorrect scanned bytes: %v", v)
		}
	})

	t.Run("should always be increasing after restart", func(t *testing.T) {
		f, err := os.Open(Delaware)
		if err != nil {
			t.Fatalf("unable to open file: %v", err)
		}
		defer f.Close()

		_, err = f.Seek(214162, 0)
		if err != nil {
			t.Fatalf("seek failed: %v", err)
		}

		scanner := New(context.Background(), f, 2)

		var previouslyScanned int64
		for scanner.Scan() {
			if v := scanner.FullyScannedBytes(); v < previouslyScanned {
				t.Errorf("scanned bytes decreased: %v < %v", v, previouslyScanned)
			}

			previouslyScanned = scanner.FullyScannedBytes()
		}

		if v := scanner.FullyScannedBytes(); v != 7274709 {
			t.Errorf("incorrect scanned bytes: %v", v)
		}
	})
}

func BenchmarkLondon(b *testing.B) {
	f, err := os.Open(London)
	if err != nil {
		b.Fatalf("could not open file: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Seek(0, 0)

		scanner := New(context.Background(), f, 4)
		nodes, ways, relations := benchmarkScanner(b, scanner)

		if nodes != 2729006 {
			b.Errorf("wrong number of nodes, got %v", nodes)
		}

		if ways != 459055 {
			b.Errorf("wrong number of ways, got %v", ways)
		}

		if relations != 12833 {
			b.Errorf("wrong number of relations, got %v", relations)
		}

		scanner.Close()
	}
}

func BenchmarkLondon_nodes(b *testing.B) {
	f, err := os.Open(London)
	if err != nil {
		b.Fatalf("could not open file: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Seek(0, 0)

		scanner := New(context.Background(), f, 4)
		scanner.SkipWays = true
		scanner.SkipRelations = true

		nodes, ways, relations := benchmarkScanner(b, scanner)

		if nodes != 2729006 {
			b.Errorf("wrong number of nodes, got %v", nodes)
		}

		if ways != 0 {
			b.Errorf("wrong number of ways, got %v", ways)
		}

		if relations != 0 {
			b.Errorf("wrong number of relations, got %v", relations)
		}

		scanner.Close()
	}
}

func BenchmarkLondon_ways(b *testing.B) {
	f, err := os.Open(London)
	if err != nil {
		b.Fatalf("could not open file: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Seek(0, 0)

		scanner := New(context.Background(), f, 4)
		scanner.SkipNodes = true
		scanner.SkipRelations = true

		nodes, ways, relations := benchmarkScanner(b, scanner)

		if nodes != 0 {
			b.Errorf("wrong number of nodes, got %v", nodes)
		}

		if ways != 459055 {
			b.Errorf("wrong number of ways, got %v", ways)
		}

		if relations != 0 {
			b.Errorf("wrong number of relations, got %v", relations)
		}

		scanner.Close()
	}
}

func BenchmarkLondon_relations(b *testing.B) {
	f, err := os.Open(London)
	if err != nil {
		b.Fatalf("could not open file: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Seek(0, 0)

		scanner := New(context.Background(), f, 4)
		scanner.SkipNodes = true
		scanner.SkipWays = true

		nodes, ways, relations := benchmarkScanner(b, scanner)

		if nodes != 0 {
			b.Errorf("wrong number of nodes, got %v", nodes)
		}

		if ways != 0 {
			b.Errorf("wrong number of ways, got %v", ways)
		}

		if relations != 12833 {
			b.Errorf("wrong number of relations, got %v", relations)
		}

		scanner.Close()
	}
}

func benchmarkScanner(b *testing.B, scanner osm.Scanner) (int, int, int) {
	var (
		nodes     int
		ways      int
		relations int
	)

	for scanner.Scan() {
		switch scanner.Object().(type) {
		case *osm.Node:
			nodes++
		case *osm.Way:
			ways++
		case *osm.Relation:
			relations++
		}
	}

	if err := scanner.Err(); err != nil {
		b.Errorf("scanner returned error: %v", err)
	}

	return nodes, ways, relations
}
