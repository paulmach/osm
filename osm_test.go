package osm

import (
	"encoding/xml"
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

	if v := n.ChangsetID; v != 456 {
		t.Errorf("incorrect changset, got %v", v)
	}

	if v := n.Timestamp; v != time.Date(2014, 4, 10, 0, 43, 05, 0, time.UTC) {
		t.Errorf("incorrect timestamp, got %v", v)
	}

	if v := n.Version; v != 1 {
		t.Errorf("incorrect version, got %v", v)
	}

	if v := n.Visible; v != true {
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

	if v := n.Lng; v != 6.0043943 {
		t.Errorf("incorrect lng, got %v", v)
	}
}
