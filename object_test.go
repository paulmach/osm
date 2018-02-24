package osm

import "testing"

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
