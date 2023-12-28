package osm

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/paulmach/orb"
)

func TestWay_ids(t *testing.T) {
	w := Way{ID: 12, Version: 2}

	if id := w.FeatureID(); id != WayID(12).FeatureID() {
		t.Errorf("incorrect feature id: %v", id)
	}

	if id := w.ElementID(); id != WayID(12).ElementID(2) {
		t.Errorf("incorrect element id: %v", id)
	}
}

func TestWay_ApplyUpdatesUpTo(t *testing.T) {
	updates := Updates{
		{Index: 0, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 11},
		{Index: 1, Timestamp: time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 12},
		{Index: 2, Timestamp: time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 13},
	}

	w := Way{
		ID:    123,
		Nodes: WayNodes{{Lat: 1}, {Lat: 2}, {Lat: 3}},
	}

	w.Updates = updates
	w.ApplyUpdatesUpTo(time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC))
	if w.Nodes[0].Lat != 1 || w.Nodes[1].Lat != 2 || w.Nodes[2].Lat != 3 {
		t.Errorf("incorrect way nodes, got %v", w.Nodes)
	}

	w.Updates = updates
	w.ApplyUpdatesUpTo(time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC))
	if w.Nodes[0].Lat != 11 || w.Nodes[1].Lat != 2 || w.Nodes[2].Lat != 13 {
		t.Errorf("incorrect way nodes, got %v", w.Nodes)
	}

	if l := len(w.Updates); l != 1 {
		t.Errorf("incorrect number of updates: %v", l)
	}

	if w.Updates[0].Index != 1 {
		t.Errorf("incorrect updates: %v", w.Updates)
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

func TestWayNode_ids(t *testing.T) {
	wn := WayNode{ID: 12, Version: 2}

	if id := wn.FeatureID(); id != NodeID(12).FeatureID() {
		t.Errorf("incorrect feature id: %v", id)
	}

	if id := wn.ElementID(); id != NodeID(12).ElementID(2) {
		t.Errorf("incorrect element id: %v", id)
	}
}

func TestWayNode_Point(t *testing.T) {
	wn := WayNode{ID: 12, Version: 2, Lon: 1, Lat: 2}

	p := wn.Point()
	if p.Lon() != 1 {
		t.Errorf("incorrect point lon: %v", p)
	}

	if p.Lat() != 2 {
		t.Errorf("incorrect point lat: %v", p)
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

func TestWayNodes_Bound(t *testing.T) {
	wn := WayNodes{
		{Lat: 1, Lon: 2},
		{Lat: 3, Lon: 4},
		{Lat: 2, Lon: 3},
	}

	b := wn.Bound()
	if !reflect.DeepEqual(b, orb.Bound{Min: orb.Point{2, 1}, Max: orb.Point{4, 3}}) {
		t.Errorf("incorrect bound: %v", b)
	}
}

func TestWay_LineString(t *testing.T) {
	w := &Way{
		ID: 1,
		Nodes: WayNodes{
			{ID: 1, Lon: 1, Lat: 2},
			{ID: 2, Lon: 0, Lat: 3},
			{ID: 3, Lon: 0, Lat: 0},
			{ID: 3, Lon: 3, Lat: 0},
			{ID: 3, Lon: 3, Lat: 4},
		},
	}

	ls := w.LineString()
	expected := orb.LineString{{1, 2}, {0, 3}, {3, 0}, {3, 4}}
	if !ls.Equal(expected) {
		t.Errorf("incorrect linestring: %v", ls)
	}

	w.Updates = Updates{
		{
			Index:     1,
			Timestamp: time.Time{},
			Lon:       10, Lat: 20,
		},
		{
			Index:     1000, // index out of range should be skipped
			Timestamp: time.Time{},
			Lon:       10, Lat: 20,
		},
		{
			Index:     0,
			Timestamp: time.Time{},
			Lon:       5, Lat: 6,
		},
		{
			Index:     4,
			Timestamp: time.Time{},
			Lon:       7, Lat: 8,
		},
		{
			Index:     2, // should be skipped
			Timestamp: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			Lon:       10, Lat: 20,
		},
	}

	ls = w.LineStringAt(time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC))
	expected = orb.LineString{{5, 6}, {10, 20}, {3, 0}, {7, 8}}

	if !ls.Equal(expected) {
		t.Errorf("incorrect line: %v", ls)
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
	w.Nodes[0] = WayNode{ID: 4, Lat: 1, Lon: 2}
	data, err = xml.Marshal(w)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`<way id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><nd ref="4" lat="1" lon="2"></nd></way>`)) {
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
	data, err = os.ReadFile("testdata/way-updates.osm")
	if err != nil {
		t.Fatalf("unable to read file: %v", err)
	}

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

func TestWayNodes_ids(t *testing.T) {
	wns := WayNodes{
		{ID: 1, Version: 3},
		{ID: 2, Version: 4},
	}

	eids := ElementIDs{NodeID(1).ElementID(3), NodeID(2).ElementID(4)}
	if ids := wns.ElementIDs(); !reflect.DeepEqual(ids, eids) {
		t.Errorf("incorrect element ids: %v", ids)
	}

	fids := FeatureIDs{NodeID(1).FeatureID(), NodeID(2).FeatureID()}
	if ids := wns.FeatureIDs(); !reflect.DeepEqual(ids, fids) {
		t.Errorf("incorrect feature ids: %v", ids)
	}

	nids := []NodeID{NodeID(1), NodeID(2)}
	if ids := wns.NodeIDs(); !reflect.DeepEqual(ids, nids) {
		t.Errorf("incorrect node ids: %v", nids)
	}
}

func TestWayNodes_UnmarshalJSON(t *testing.T) {
	wn := WayNodes{}

	if err := wn.UnmarshalJSON([]byte("[asdf,]")); err == nil {
		t.Errorf("should return error when json is invalid")
	}

	json := []byte(`[1,2,3,4]`)
	err := wn.UnmarshalJSON(json)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	expected := []NodeID{1, 2, 3, 4}
	if ids := wn.NodeIDs(); !reflect.DeepEqual(ids, expected) {
		t.Errorf("incorrect ids: %v", ids)
	}
}

func TestWays_ids(t *testing.T) {
	ws := Ways{
		{ID: 1, Version: 3},
		{ID: 2, Version: 4},
	}

	eids := ElementIDs{WayID(1).ElementID(3), WayID(2).ElementID(4)}
	if ids := ws.ElementIDs(); !reflect.DeepEqual(ids, eids) {
		t.Errorf("incorrect element ids: %v", ids)
	}

	fids := FeatureIDs{WayID(1).FeatureID(), WayID(2).FeatureID()}
	if ids := ws.FeatureIDs(); !reflect.DeepEqual(ids, fids) {
		t.Errorf("incorrect feature ids: %v", ids)
	}

	wids := []WayID{1, 2}
	if ids := ws.IDs(); !reflect.DeepEqual(ids, wids) {
		t.Errorf("incorrect way ids: %v", wids)
	}
}

func TestWays_SortByIDVersion(t *testing.T) {
	ws := Ways{
		{ID: 7, Version: 3},
		{ID: 2, Version: 4},
		{ID: 5, Version: 2},
		{ID: 5, Version: 3},
		{ID: 5, Version: 4},
		{ID: 3, Version: 4},
		{ID: 4, Version: 4},
		{ID: 9, Version: 4},
	}

	ws.SortByIDVersion()

	eids := ElementIDs{
		WayID(2).ElementID(4),
		WayID(3).ElementID(4),
		WayID(4).ElementID(4),
		WayID(5).ElementID(2),
		WayID(5).ElementID(3),
		WayID(5).ElementID(4),
		WayID(7).ElementID(3),
		WayID(9).ElementID(4),
	}

	if ids := ws.ElementIDs(); !reflect.DeepEqual(ids, eids) {
		t.Errorf("incorrect sort: %v", eids)
	}
}
