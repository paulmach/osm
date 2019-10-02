package osm

import (
	"reflect"
	"testing"
)

func TestParseObjectID(t *testing.T) {
	cases := []struct {
		name   string
		string string
		id     ObjectID
	}{
		{
			name: "node",
			id:   NodeID(0).ObjectID(1),
		},
		{
			name: "zero version node",
			id:   NodeID(3).ObjectID(0),
		},
		{
			name: "way",
			id:   WayID(10).ObjectID(2),
		},
		{
			name: "relation",
			id:   RelationID(100).ObjectID(3),
		},
		{
			name: "changeset",
			id:   ChangesetID(1000).ObjectID(),
		},
		{
			name: "note",
			id:   NoteID(10000).ObjectID(),
		},
		{
			name: "user",
			id:   UserID(5000).ObjectID(),
		},
		{
			name:   "node feature",
			string: "node/100",
			id:     NodeID(100).ObjectID(0),
		},
		{
			name:   "bounds",
			string: "bounds/0",
			id:     (&Bounds{}).ObjectID(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				id  ObjectID
				err error
			)

			if tc.string == "" {
				id, err = ParseObjectID(tc.id.String())
				if err != nil {
					t.Errorf("parse error: %v", err)
					return
				}
			} else {
				id, err = ParseObjectID(tc.string)
				if err != nil {
					t.Errorf("parse error: %v", err)
					return
				}
			}

			if id != tc.id {
				t.Errorf("incorrect id: %v != %v", id, tc.id)
			}
		})
	}

	// errors
	if _, err := ParseObjectID("123"); err == nil {
		t.Errorf("should return error if only one part")
	}

	if _, err := ParseObjectID("node/1:1:1"); err == nil {
		t.Errorf("should return error if multiple :")
	}

	if _, err := ParseObjectID("node/abc:1"); err == nil {
		t.Errorf("should return error if id not a number")
	}

	if _, err := ParseObjectID("node/1:abc"); err == nil {
		t.Errorf("should return error if version not a number")
	}

	if _, err := ParseObjectID("lake/1:1"); err == nil {
		t.Errorf("should return error if not a valid type")
	}
}

func TestObjects_ObjectIDs(t *testing.T) {
	es := Objects{
		&Node{ID: 1, Version: 5},
		&Way{ID: 2, Version: 6},
		&Relation{ID: 3, Version: 7},
		&Node{ID: 4, Version: 8},
		&User{ID: 5},
		&Note{ID: 6},
	}

	expected := ObjectIDs{
		NodeID(1).ObjectID(5),
		WayID(2).ObjectID(6),
		RelationID(3).ObjectID(7),
		NodeID(4).ObjectID(8),
		UserID(5).ObjectID(),
		NoteID(6).ObjectID(),
	}

	if ids := es.ObjectIDs(); !reflect.DeepEqual(ids, expected) {
		t.Errorf("incorrect ids: %v", ids)
	}
}

func TestObjectID_implementations(t *testing.T) {
	type oid interface {
		ObjectID() ObjectID
	}

	var _ oid = ElementID(0)

	var _ oid = &Node{}
	var _ oid = &Way{}
	var _ oid = &Relation{}
	var _ oid = &Changeset{}
	var _ oid = &Note{}
	var _ oid = &User{}

	var _ oid = ChangesetID(0)
	var _ oid = NoteID(0)
	var _ oid = UserID(0)

	type oidv interface {
		ObjectID(v int) ObjectID
	}

	var _ oidv = FeatureID(0)
	var _ oidv = NodeID(0)
	var _ oidv = WayID(0)
	var _ oidv = RelationID(0)

	// These should not implement the ObjectID methods
	noImplement := []interface{}{
		WayNode{},
		Member{},
	}

	for _, ni := range noImplement {
		if _, ok := ni.(oid); ok {
			t.Errorf("%T should not have ObjectID() method", ni)
		}

		if _, ok := ni.(oidv); ok {
			t.Errorf("%T should not have ObjectID(v int) method", ni)
		}
	}
}

func TestObject_implementations(t *testing.T) {
	var _ Object = &Node{}
	var _ Object = &Way{}
	var _ Object = &Relation{}
	var _ Object = &Changeset{}
	var _ Object = &Note{}
	var _ Object = &User{}

	// These should not implement the Object interface
	noImplement := []interface{}{
		ObjectID(0),
		FeatureID(0),
		ElementID(0),
		WayNode{},
		Member{},

		NodeID(0),
		WayID(0),
		RelationID(0),
		ChangesetID(0),
		NoteID(0),
		UserID(0),

		Nodes{},
		Ways{},
		Relations{},
		Changesets{},
		Notes{},
		Users{},
	}

	for _, ni := range noImplement {
		if _, ok := ni.(Object); ok {
			t.Errorf("%T should not be an object", ni)
		}
	}
}
