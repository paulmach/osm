package osmpbf

import (
	"os"
	"testing"

	"golang.org/x/net/context"
)

func TestScanner(t *testing.T) {
	file := os.Getenv("OSMPBF_BENCHMARK_FILE")
	if file == "" {
		file = London
	}

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	scanner := New(context.Background(), f, 1)
	defer scanner.Close()

	if v := scanner.Scan(); v == false {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if n := scanner.Element().Node; n.ID != 44 {
		t.Fatalf("did not scan correctly, got %v", n)
	}

	if v := scanner.Scan(); v == false {
		t.Fatalf("should read second scan: %v", scanner.Err())
	}

	if n := scanner.Element().Node; n.ID != 47 {
		t.Fatalf("did not scan correctly, got %v", n)
	}
}

func TestChangesetScannerContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	file := os.Getenv("OSMPBF_BENCHMARK_FILE")
	if file == "" {
		file = London
	}

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	scanner := New(ctx, f, 1)
	defer scanner.Close()

	if v := scanner.Scan(); v == false {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if n := scanner.Element().Node; n.ID != 44 {
		t.Fatalf("did not scan correctly, got %v", n)
	}

	cancel()

	if v := scanner.Scan(); v == true {
		t.Fatalf("should be closed for second scan: %v", scanner.Err())
	}

	if v := scanner.Err(); v != ctx.Err() {
		t.Errorf("incorrect error, got %v", v)
	}
}

func BenchmarkLondon(b *testing.B) {
	f, err := os.Open("greater-london-140324.osm.pbf")
	if err != nil {
		b.Fatalf("could not open file: %v", err)
	}

	scanner := New(context.Background(), f, 4)

	var (
		nodes     int
		ways      int
		relations int
	)

	b.ReportAllocs()
	b.ResetTimer()
	for scanner.Scan() {
		e := scanner.Element()
		if e.Node != nil {
			nodes++
		}

		if e.Way != nil {
			ways++
		}

		if e.Relation != nil {
			relations++
		}
	}

	if scanner.Err() != nil {
		b.Errorf("scanner returned error: %v", err)
	}

	b.Logf("nodes %d, ways %d, relations %d", nodes, ways, relations)
}
