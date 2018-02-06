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

	if n := scanner.Element().(*osm.Node); n.ID != 75385503 {
		t.Fatalf("did not scan correctly, got %v", n)
	}

	if v := scanner.Scan(); !v {
		t.Fatalf("should read second scan: %v", scanner.Err())
	}

	if n := scanner.Element().(*osm.Node); n.ID != 75390099 {
		t.Fatalf("did not scan correctly, got %v", n)
	}
}

func TestScannerIntermediateStart(t *testing.T) {
	f, err := os.Open(Delaware)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	scanner := New(context.Background(), f, 1)

	target := osm.NodeID(178592359) // first element in last partially scanned block
	indexOfTarget := 0
	for i := 0; i < 30000; i++ {
		scanner.Scan()
		if scanner.Element().(*osm.Node).ID == target {
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
	if id := scanner.Element().(*osm.Node).ID; id != target {
		t.Errorf("incorrect first id, got %v", id)
	}
	scanner.Close()
}

func TestChangesetScannerContext(t *testing.T) {
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

	if n := scanner.Element().(*osm.Node); n.ID != 75385503 {
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

func TestScannerHeader(t *testing.T) {
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

func TestChangesetScannerClose(t *testing.T) {
	f, err := os.Open(Delaware)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	scanner := New(context.Background(), f, 1)

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if n := scanner.Element().(*osm.Node); n.ID != 75385503 {
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

func benchmarkScanner(b *testing.B, scanner osm.Scanner) (int, int, int) {
	var (
		nodes     int
		ways      int
		relations int
	)

	for scanner.Scan() {
		switch scanner.Element().(type) {
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
