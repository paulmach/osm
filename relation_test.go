package osm

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"testing"
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
