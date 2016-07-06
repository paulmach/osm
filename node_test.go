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

	if v := n.ChangesetID; v != 456 {
		t.Errorf("incorrect changeset, got %v", v)
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

func TestNodesActiveAt(t *testing.T) {
	nodes := Nodes{
		{ID: 1, Timestamp: time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Timestamp: time.Date(2016, 2, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Timestamp: time.Date(2016, 3, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Timestamp: time.Date(2016, 4, 1, 0, 0, 0, 0, time.UTC)},
	}

	type testCase struct {
		ID   NodeID
		Time time.Time
	}

	tests := []testCase{
		{0, time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)},
		{1, time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)},
		{1, time.Date(2016, 1, 10, 0, 0, 0, 0, time.UTC)},
		{2, time.Date(2016, 2, 5, 0, 0, 0, 0, time.UTC)},
		{4, time.Date(2017, 1, 10, 0, 0, 0, 0, time.UTC)},
	}

	for i, test := range tests {
		n := nodes.ActiveAt(test.Time)
		if n == nil && test.ID != 0 {
			t.Errorf("test %d: expected nil, got %d", i, test.ID)
			continue
		}

		if n == nil {
			continue
		}

		if test.ID != n.ID {
			t.Errorf("test %d: expect %d, got %d", i, test.ID, n.ID)
		}
	}
}
