package osm

import (
	"bytes"
	"encoding/xml"
	"testing"
)

func TestWayMarshalXML(t *testing.T) {
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
}
