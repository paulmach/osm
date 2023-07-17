package osmxml

import (
	"bytes"
	"compress/bzip2"
	"context"
	"io"
	"os"
	"testing"

	"github.com/onXmaps/osm"
)

func TestScanner(t *testing.T) {
	r := changesetReader()
	scanner := New(context.Background(), r)

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if cs := scanner.Object().(*osm.Changeset); cs.ID != 41226352 {
		t.Fatalf("did not scan correctly, got %v", cs)
	}

	if v := scanner.Scan(); !v {
		t.Fatalf("should read second scan: %v", scanner.Err())
	}

	if cs := scanner.Object().(*osm.Changeset); cs.ID != 41227987 {
		t.Fatalf("did not scan correctly, got %v", cs)
	}

	if et := scanner.Object().ObjectID().Type(); et != osm.TypeChangeset {
		t.Fatalf("did not set type correctly, got %v", et)
	}

	if cs := scanner.Object().ObjectID().Ref(); cs != 41227987 {
		t.Fatalf("did not set id correctly, got %v", cs)
	}

	if v := scanner.Scan(); v {
		t.Fatalf("should be finished scanning")
	}
}

func TestScanner_context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r := changesetReader()

	scanner := New(ctx, r)

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if cs := scanner.Object().(*osm.Changeset); cs.ID != 41226352 {
		t.Fatalf("did not scan correctly, got %v", cs)
	}

	cancel()

	if v := scanner.Scan(); v {
		t.Fatalf("should be closed for second scan: %v", scanner.Err())
	}

	if v := scanner.Err(); v != ctx.Err() {
		t.Errorf("incorrect error, got %v", v)
	}
}

func TestScanner_Close(t *testing.T) {
	r := changesetReader()
	scanner := New(context.Background(), r)

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if cs := scanner.Object().(*osm.Changeset); cs.ID != 41226352 {
		t.Fatalf("did not scan correctly, got %v", cs)
	}

	scanner.Close()

	if v := scanner.Scan(); v {
		t.Fatalf("should be closed for second scan: %v", scanner.Err())
	}

	if v := scanner.Err(); v != osm.ErrScannerClosed {
		t.Errorf("incorrect error, got %v", v)
	}
}

func TestScanner_Err(t *testing.T) {
	r := changesetReaderErr()
	scanner := New(context.Background(), r)

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if cs := scanner.Object().(*osm.Changeset); cs.ID != 41226352 {
		t.Fatalf("did not scan correctly, got %v", cs)
	}

	if v := scanner.Scan(); v {
		t.Fatalf("should be closed for second scan: %v", scanner.Err())
	}

	if v := scanner.Scan(); v {
		t.Fatalf("should continue to be closed: %v", scanner.Err())
	}

	if v := scanner.Err(); v == nil {
		t.Errorf("incorrect error, got %v", v)
	}

	scanner.Close()
	if v := scanner.Err(); v == osm.ErrScannerClosed {
		t.Errorf("should return xml error not closed error, got %v", v)
	}
}

func TestScanner_userNote(t *testing.T) {
	r := userNoteReader()
	scanner := New(context.Background(), r)
	defer scanner.Close()

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if u := scanner.Object().(*osm.User); u.ID != 1 {
		t.Fatalf("did not scan correctly, got %v", u)
	}

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	if n := scanner.Object().(*osm.Note); n.ID != 2 {
		t.Fatalf("did not scan correctly, got %v", n)
	}
}

func TestScanner_bounds(t *testing.T) {
	r := boundsReader()
	scanner := New(context.Background(), r)
	defer scanner.Close()

	if v := scanner.Scan(); !v {
		t.Fatalf("should read first scan: %v", scanner.Err())
	}

	b := scanner.Object().(*osm.Bounds)
	if b.MinLat != 1 || b.MinLon != 2 || b.MaxLat != 3 || b.MaxLon != 4 {
		t.Fatalf("did not scan correctly, got: %v", b)
	}
}

func TestAndorra(t *testing.T) {
	f, err := os.Open("../testdata/andorra-latest.osm.bz2")
	if err != nil {
		t.Fatalf("could not open file: %v", err)
	}

	r := bzip2.NewReader(f)
	scanner := New(context.Background(), r)

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

	if scanner.Err() != nil {
		t.Errorf("scanner returned error: %v", err)
	}

	if nodes != 203265 {
		t.Errorf("incorrect number of nodes: %v", nodes)
	}

	if ways != 9080 {
		t.Errorf("incorrect number of ways: %v", ways)
	}

	if relations != 233 {
		t.Errorf("incorrect number of relations: %v", relations)
	}
}

func BenchmarkAndorra(b *testing.B) {
	f, err := os.Open("../testdata/andorra-latest.osm.bz2")
	if err != nil {
		b.Fatalf("could not open file: %v", err)
	}

	r := bzip2.NewReader(f)
	scanner := New(context.Background(), r)

	var (
		nodes     int
		ways      int
		relations int
	)

	b.ReportAllocs()
	b.ResetTimer()
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

	if scanner.Err() != nil {
		b.Errorf("scanner returned error: %v", err)
	}

	b.Logf("nodes %d, ways %d, relations %d", nodes, ways, relations)
}

func userNoteReader() io.Reader {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<osm>
  <user id="1"></user>
  <note><id>2</id></note>
</osm>`)

	return bytes.NewReader(data)
}

func boundsReader() io.Reader {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<osm>
	<bounds minlat="1" minlon="2" maxlat="3" maxlon="4"/>
</osm>`)

	return bytes.NewReader(data)
}

func changesetReader() io.Reader {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6" generator="replicate_changesets.rb" copyright="OpenStreetMap and contributors" attribution="http://www.openstreetmap.org/copyright" license="http://opendatacommons.org/licenses/odbl/1-0/">
  <changeset id="41226352" created_at="2016-08-03T22:40:15Z" closed_at="2016-08-04T01:41:27Z" open="false" num_changes="112" user="dragon_ear" uid="321257" min_lat="36.496286" max_lat="36.6110983" min_lon="136.6138636" max_lon="136.644462" comments_count="0">
    <tag k="comment" v="updated fire hydrant details with OsmHydrant"/>
    <tag k="created_by" v="OsmHydrant / http://yapafo.net v0.3"/>
  </changeset>
  <changeset id="41227987" created_at="2016-08-04T01:41:04Z" closed_at="2016-08-04T01:41:07Z" open="false" num_changes="7" user="MapAnalyser465" uid="3077038" min_lat="-33.7963817" max_lat="-33.7881945" min_lon="151.2527542" max_lon="151.2667464" comments_count="0">
    <tag k="comment" v="Updated Burnt Creek Deviation to Motorway Standard"/>
    <tag k="locale" v="en"/>
    <tag k="host" v="https://www.openstreetmap.org/id"/>
    <tag k="imagery_used" v="Bing"/>
    <tag k="created_by" v="iD 1.9.7"/>
  </changeset>
</osm>`)

	return bytes.NewReader(data)
}

func changesetReaderErr() io.Reader {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6" generator="replicate_changesets.rb" copyright="OpenStreetMap and contributors" attribution="http://www.openstreetmap.org/copyright" license="http://opendatacommons.org/licenses/odbl/1-0/">
  <changeset id="41226352" created_at="2016-08-03T22:40:15Z" closed_at="2016-08-04T01:41:27Z" open="false" num_changes="112" user="dragon_ear" uid="321257" min_lat="36.496286" max_lat="36.6110983" min_lon="136.6138636" max_lon="136.644462" comments_count="0">
    <tag k="comment" v="updated fire hydrant details with OsmHydrant"/>
    <tag k="created_by" v="OsmHydrant / http://yapafo.net v0.3"/>
  </changeset>
  <changeset id="41227987" created_at="2016-08-04T01:41:04Z" closed_at="2016-08-04T01:41:07Z" open="false" num_changes="7" user="MapAnalyser465" uid="3077038" min_lat="-33.7963817" max_lat="-33.7881945" min_lon="151.2527542" max_lon="151.2667464" comments_count="0">
    <tag k="comment" v="Updated Burnt Creek Deviation to Motorway Standard"/>`)

	return bytes.NewReader(data)
}
