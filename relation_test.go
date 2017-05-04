package osm

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

func TestRelationMarshalJSON(t *testing.T) {
	r := Relation{
		ID: 123,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`{"type":"relation","id":123,"visible":false,"timestamp":"0001-01-01T00:00:00Z","members":[]}`)) {
		t.Errorf("incorrect json: %v", string(data))
	}

	// with members
	r = Relation{
		ID:      123,
		Members: []Member{{Type: "node", Ref: 123, Role: "outer", Version: 1}},
	}

	data, err = json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`{"type":"relation","id":123,"visible":false,"timestamp":"0001-01-01T00:00:00Z","members":[{"type":"node","ref":123,"role":"outer","version":1}]}`)) {
		t.Errorf("incorrect json: %v", string(data))
	}
}

func TestRelationApplyUpdatesUpTo(t *testing.T) {
	r := Relation{
		ID:      123,
		Members: []Member{{Version: 1}, {Version: 2}, {Version: 3}},
		Updates: Updates{
			{Index: 0, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Version: 11},
			{Index: 1, Timestamp: time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC), Version: 12},
			{Index: 2, Timestamp: time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC), Version: 13, Lat: 10, Lon: 20},
		},
	}

	r.ApplyUpdatesUpTo(time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC))
	if r.Members[0].Version != 1 || r.Members[1].Version != 2 || r.Members[2].Version != 3 {
		t.Errorf("incorrect members, got %v", r.Members)
	}

	r.ApplyUpdatesUpTo(time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC))
	if r.Members[0].Version != 11 || r.Members[1].Version != 2 || r.Members[2].Version != 13 {
		t.Errorf("incorrect members, got %v", r.Members)
	}

	if r.Members[2].Lat != 10 {
		t.Errorf("did not apply lat data")
	}

	if r.Members[2].Lon != 20 {
		t.Errorf("did not apply lon data")
	}
}

func TestRelationApplyUpdate(t *testing.T) {
	r := Relation{
		ID:      123,
		Members: []Member{{Ref: 1, Type: TypeNode}},
	}

	err := r.applyUpdate(Update{
		Index:       0,
		Version:     1,
		ChangesetID: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := Member{
		Ref:         1,
		Type:        TypeNode,
		Version:     1,
		ChangesetID: 2,
	}

	if expected != r.Members[0] {
		t.Errorf("incorrect update, got %v", r.Members[0])
	}
}

func TestRelationApplyUpdateError(t *testing.T) {
	r := Relation{
		ID:      123,
		Members: []Member{{Ref: 1, Type: TypeNode}},
	}

	err := r.applyUpdate(Update{
		Index: 1,
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if e, ok := err.(*UpdateIndexOutOfRangeError); !ok {
		t.Errorf("incorrect error, got %v", e)
	}
}

func TestRelationMarshalXML(t *testing.T) {
	r := Relation{
		ID: 123,
	}

	data, err := xml.Marshal(r)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	expected := `<relation id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"></relation>`
	if !bytes.Equal(data, []byte(expected)) {
		t.Errorf("incorrect marshal, got: %s", string(data))
	}

	// members
	r.Members = []Member{{Type: "node", Ref: 123, Role: "child"}}
	data, err = xml.Marshal(r)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`<relation id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><member type="node" ref="123" role="child"></member></relation>`)) {
		t.Errorf("not marshalled correctly: %s", string(data))
	}

	// minor relation
	r.Members = nil
	r.Updates = []Update{
		{
			Index:       0,
			Version:     1,
			ChangesetID: 123,
			Timestamp:   time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	data, err = xml.Marshal(r)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`<relation id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><update index="0" version="1" timestamp="2012-01-01T00:00:00Z" changeset="123"></update></relation>`)) {
		t.Errorf("not marshalled correctly: %s", string(data))
	}

	// blanket xml test
	data = readFile(t, "testdata/relation-updates.osm")

	osm := &OSM{}
	err = xml.Unmarshal(data, &osm)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	relation := osm.Relations[0]

	var i1 interface{}
	err = xml.Unmarshal(data, &i1)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	data, err = xml.Marshal(relation)
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

func TestRelationsMarshal(t *testing.T) {
	rs := Relations{
		{ID: 123},
		{ID: 321},
	}

	data, err := rs.Marshal()
	if err != nil {
		t.Fatalf("relations marshal error: %v", err)
	}

	rs2, err := UnmarshalRelations(data)
	if err != nil {
		t.Fatalf("relations unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(rs, rs2) {
		t.Errorf("relations not equal")
		t.Logf("%+v", rs)
		t.Logf("%+v", rs2)
	}

	// empty relations
	rs = Relations{}

	data, err = rs.Marshal()
	if err != nil {
		t.Fatalf("relations marshal error: %v", err)
	}

	if l := len(data); l != 0 {
		t.Errorf("length of relation data should be 0, got %v", l)
	}

	rs2, err = UnmarshalRelations(data)
	if err != nil {
		t.Fatalf("relations unmarshal error: %v", err)
	}

	if rs2 != nil {
		t.Errorf("should return nil Relations for empty data, got %v", rs2)
	}
}
