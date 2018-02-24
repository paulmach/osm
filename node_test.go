package osm

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

func TestNode(t *testing.T) {
	data := []byte(`<node id="123" changeset="456" timestamp="2014-04-10T00:43:05Z" version="1" visible="true" user="user" uid="1357" lat="50.7107023" lon="6.0043943"/>`)

	n := Node{}
	err := xml.Unmarshal(data, &n)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if v := n.ID; v != 123 {
		t.Errorf("incorrect id, got %v", v)
	}

	if v := n.ChangesetID; v != 456 {
		t.Errorf("incorrect changeset, got %v", v)
	}

	if v := n.Timestamp; v != time.Date(2014, 4, 10, 0, 43, 05, 0, time.UTC) {
		t.Errorf("incorrect timestamp, got %v", v)
	}

	if v := n.Version; v != 1 {
		t.Errorf("incorrect version, got %v", v)
	}

	if v := n.Visible; !v {
		t.Errorf("incorrect visible, got %v", v)
	}

	if v := n.User; v != "user" {
		t.Errorf("incorrect user, got %v", v)
	}

	if v := n.UserID; v != 1357 {
		t.Errorf("incorrect user id, got %v", v)
	}

	if v := n.Lat; v != 50.7107023 {
		t.Errorf("incorrect lat, got %v", v)
	}

	if v := n.Lon; v != 6.0043943 {
		t.Errorf("incorrect lon, got %v", v)
	}
}

func TestNode_MarshalJSON(t *testing.T) {
	n := Node{
		ID: 123,
	}

	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`{"type":"node","id":123,"lat":0,"lon":0,"visible":false,"timestamp":"0001-01-01T00:00:00Z"}`)) {
		t.Errorf("incorrect json: %v", string(data))
	}
}

func TestNode_MarshalXML(t *testing.T) {
	n := Node{
		ID: 123,
	}

	data, err := xml.Marshal(n)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	expected := `<node id="123" lat="0" lon="0" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"></node>`
	if !bytes.Equal(data, []byte(expected)) {
		t.Errorf("incorrect marshal, got: %s", string(data))
	}
}

func TestUnmarshalNodes(t *testing.T) {
	ns := Nodes{
		{ID: 123},
		{ID: 321},
	}

	data, err := ns.Marshal()
	if err != nil {
		t.Fatalf("nodes marshal error: %v", err)
	}

	ns2, err := UnmarshalNodes(data)
	if err != nil {
		t.Fatalf("nodes unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(ns, ns2) {
		t.Errorf("nodes not equal")
		t.Logf("%+v", ns)
		t.Logf("%+v", ns2)
	}

	// empty nodes
	ns = Nodes{}

	data, err = ns.Marshal()
	if err != nil {
		t.Fatalf("nodes marshal error: %v", err)
	}

	if l := len(data); l != 0 {
		t.Errorf("length of node data should be 0, got %v", l)
	}

	ns2, err = UnmarshalNodes(data)
	if err != nil {
		t.Fatalf("nodes unmarshal error: %v", err)
	}

	if ns2 != nil {
		t.Errorf("should return nil Nodes for empty data, got %v", ns2)
	}
}
