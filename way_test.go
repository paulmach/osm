package osm

import (
	"bytes"
	"encoding/xml"
	"reflect"
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

func TestWaysMarshal(t *testing.T) {
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
