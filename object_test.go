package osm

import "testing"

func TestObjectIDImplementations(t *testing.T) {
	type oid interface {
		ObjectID() ObjectID
	}

	var _ oid = FeatureID(0)
	var _ oid = ElementID(0)

	var _ oid = &Node{}
	var _ oid = &Way{}
	var _ oid = &Relation{}
	var _ oid = &Changeset{}
	var _ oid = &Note{}
	var _ oid = &User{}

	var _ oid = NodeID(0)
	var _ oid = WayID(0)
	var _ oid = RelationID(0)
	var _ oid = ChangesetID(0)
	var _ oid = NoteID(0)
	var _ oid = UserID(0)

	// These should not implement the ObjectID methods
	noImplement := []interface{}{
		WayNode{},
		Member{},
	}

	for _, ni := range noImplement {
		if _, ok := ni.(oid); ok {
			t.Errorf("%T should not have ObjectID() method", ni)
		}
	}
}

func TestObjectImplementations(t *testing.T) {
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
