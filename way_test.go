package osm

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

func TestWay_ApplyUpdatesUpTo(t *testing.T) {
	w := Way{
		ID:    123,
		Nodes: WayNodes{{Lat: 1}, {Lat: 2}, {Lat: 3}},
		Updates: Updates{
			{Index: 0, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 11},
			{Index: 2, Timestamp: time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 13},
			{Index: 1, Timestamp: time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 12},
		},
	}

	w.ApplyUpdatesUpTo(time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC))
	if w.Nodes[0].Lat != 1 || w.Nodes[1].Lat != 2 || w.Nodes[2].Lat != 3 {
		t.Errorf("incorrect way nodes, got %v", w.Nodes)
	}

	w.ApplyUpdatesUpTo(time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC))
	if w.Nodes[0].Lat != 11 || w.Nodes[1].Lat != 2 || w.Nodes[2].Lat != 13 {
		t.Errorf("incorrect way nodes, got %v", w.Nodes)
	}
}

func TestWay_ApplyUpdate(t *testing.T) {
	w := Way{
		ID:    123,
		Nodes: WayNodes{{Lat: 1, Lon: 2}},
	}

	err := w.applyUpdate(Update{
		Index:       0,
		Version:     1,
		ChangesetID: 2,
		Lat:         3,
		Lon:         4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := WayNode{
		ID:          0,
		Version:     1,
		ChangesetID: 2,
		Lat:         3,
		Lon:         4,
	}

	if !reflect.DeepEqual(expected, w.Nodes[0]) {
		t.Errorf("incorrect update, got %+v", w.Nodes[0])
	}
}

func TestWay_ApplyUpdate_error(t *testing.T) {
	w := Way{
		ID:    123,
		Nodes: WayNodes{{Lat: 1, Lon: 2}},
	}

	err := w.applyUpdate(Update{
		Index: 1,
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if e, ok := err.(*UpdateIndexOutOfRangeError); !ok {
		t.Errorf("incorrect error, got %v", e)
	}
}

func TestWayNodes_Bounds(t *testing.T) {
	wn := WayNodes{
		{Lat: 1, Lon: 2},
		{Lat: 3, Lon: 4},
		{Lat: 2, Lon: 3},
	}

	b := wn.Bounds()
	if !reflect.DeepEqual(b, &Bounds{1, 3, 2, 4}) {
		t.Errorf("incorrect bounds: %v", b)
	}
}

func TestWay_MarshalJSON(t *testing.T) {
	w := Way{
		ID:    123,
		Nodes: WayNodes{{ID: 1}, {ID: 2}, {ID: 4}},
	}

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`{"type":"way","id":123,"visible":false,"timestamp":"0001-01-01T00:00:00Z","nodes":[1,2,4]}`)) {
		t.Errorf("incorrect json: %v", string(data))
	}
}

func TestWay_MarshalXML(t *testing.T) {
	w := Way{
		ID: 123,
	}

	data, err := xml.Marshal(w)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	expected := `<way id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"></way>`
	if !bytes.Equal(data, []byte(expected)) {
		t.Errorf("incorrect marshal, got: %s", string(data))
	}

	// node refs
	w.Nodes = WayNodes{{ID: 123}}
	data, err = xml.Marshal(w)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`<way id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><nd ref="123"></nd></way>`)) {
		t.Errorf("not marshalled correctly: %s", string(data))
	}

	// node with lat/lon
	w.Nodes[0] = WayNode{Lat: 1, Lon: 2}
	data, err = xml.Marshal(w)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`<way id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><nd ref="0" lat="1" lon="2"></nd></way>`)) {
		t.Errorf("not marshalled correctly: %s", string(data))
	}

	// minor way
	w.Nodes = nil
	w.Updates = []Update{
		{
			Index:     0,
			Version:   2,
			Lat:       100.0,
			Lon:       200.0,
			Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	data, err = xml.Marshal(w)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`<way id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><update index="0" version="2" timestamp="2012-01-01T00:00:00Z" lat="100" lon="200"></update></way>`)) {
		t.Errorf("not marshalled correctly: %s", string(data))
	}

	// blanket xml test
	data = readFile(t, "testdata/way-updates.osm")

	osm := &OSM{}
	err = xml.Unmarshal(data, &osm)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	way := osm.Ways[0]

	var i1 interface{}
	err = xml.Unmarshal(data, &i1)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	data, err = xml.Marshal(way)
	if err != nil {
		t.Errorf("marshal error: %v", err)
	}

	var i2 interface{}
	err = xml.Unmarshal(data, &i2)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(i1, i2) {
		t.Errorf("interfaces not equal")
		t.Logf("%+v", i1)
		t.Logf("%+v", i2)
	}
}

func TestWays_Marshal(t *testing.T) {
	ws := Ways{
		{ID: 123},
		{ID: 321},
	}

	data, err := ws.Marshal()
	if err != nil {
		t.Fatalf("ways marshal error: %v", err)
	}

	ws2, err := UnmarshalWays(data)
	if err != nil {
		t.Fatalf("ways unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(ws, ws2) {
		t.Errorf("ways not equal")
		t.Logf("%+v", ws)
		t.Logf("%+v", ws2)
	}

	// empty ways
	ws = Ways{}

	data, err = ws.Marshal()
	if err != nil {
		t.Fatalf("ways marshal error: %v", err)
	}

	if l := len(data); l != 0 {
		t.Errorf("length of way data should be 0, got %v", l)
	}

	ws2, err = UnmarshalWays(data)
	if err != nil {
		t.Fatalf("ways unmarshal error: %v", err)
	}

	if ws2 != nil {
		t.Errorf("should return nil Ways for empty data, got %v", ws2)
	}
}
