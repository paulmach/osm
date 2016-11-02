package osm

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

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

	if !bytes.Equal(data, []byte(`<relation id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><update index="0" version="1" minor="false" timestamp="2012-01-01T00:00:00Z" changeset="123"></update></relation>`)) {
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
