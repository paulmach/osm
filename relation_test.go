package osm

import (
	"bytes"
	"encoding/xml"
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
