package osm

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestElementID_ids(t *testing.T) {
	id := NodeID(1).ElementID(1)

	oid := id.ObjectID()
	if v := oid.Type(); v != TypeNode {
		t.Errorf("incorrect type: %v", v)
	}

	if v := oid.Ref(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	fid := id.FeatureID()
	if v := fid.Type(); v != TypeNode {
		t.Errorf("incorrect type: %v", v)
	}

	if v := fid.Ref(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	if v := NodeID(1).ElementID(1).NodeID(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	if v := WayID(1).ElementID(1).WayID(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	if v := RelationID(1).ElementID(1).RelationID(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	t.Run("not a node", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("should panic?")
			}
		}()

		id := WayID(1).ElementID(1)
		id.NodeID()
	})

	t.Run("not a way", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("should panic?")
			}
		}()

		id := NodeID(1).ElementID(1)
		id.WayID()
	})

	t.Run("not a relation", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("should panic?")
			}
		}()

		id := WayID(1).ElementID(1)
		id.RelationID()
	})
}

func TestParseElementID(t *testing.T) {
	cases := []struct {
		name   string
		string string
		id     ElementID
	}{
		{
			name: "node",
			id:   NodeID(0).ElementID(1),
		},
		{
			name: "zero version node",
			id:   NodeID(3).ElementID(0),
		},
		{
			name: "way",
			id:   WayID(10).ElementID(2),
		},
		{
			name: "relation",
			id:   RelationID(100).ElementID(3),
		},
		{
			name:   "node feature",
			string: "node/100",
			id:     NodeID(100).ElementID(0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				id  ElementID
				err error
			)

			if tc.string == "" {
				id, err = ParseElementID(tc.id.String())
				if err != nil {
					t.Errorf("parse error: %v", err)
					return
				}
			} else {
				id, err = ParseElementID(tc.string)
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
	if _, err := ParseElementID("123"); err == nil {
		t.Errorf("should return error if only one part")
	}

	if _, err := ParseElementID("node/1:1:1"); err == nil {
		t.Errorf("should return error if multiple :")
	}

	if _, err := ParseElementID("node/abc:1"); err == nil {
		t.Errorf("should return error if id not a number")
	}

	if _, err := ParseElementID("node/1:abc"); err == nil {
		t.Errorf("should return error if version not a number")
	}

	if _, err := ParseElementID("lake/1:1"); err == nil {
		t.Errorf("should return error if not a valid type")
	}
}

func TestElement_implementations(t *testing.T) {
	var _ Element = &Node{}
	var _ Element = &Way{}
	var _ Element = &Relation{}

	// These should not implement the Element interface
	noImplement := []interface{}{
		ObjectID(0),
		FeatureID(0),
		ElementID(0),
		&Changeset{},
		&Note{},
		&User{},
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
		if _, ok := ni.(Element); ok {
			t.Errorf("%T should not be an element", ni)
		}
	}
}

func TestElements_ElementIDs(t *testing.T) {
	es := Elements{
		&Node{ID: 1, Version: 5},
		&Way{ID: 2, Version: 6},
		&Relation{ID: 3, Version: 7},
		&Node{ID: 4, Version: 8},
	}

	expected := ElementIDs{
		NodeID(1).ElementID(5),
		WayID(2).ElementID(6),
		RelationID(3).ElementID(7),
		NodeID(4).ElementID(8),
	}

	if ids := es.ElementIDs(); !reflect.DeepEqual(ids, expected) {
		t.Errorf("incorrect ids: %v", ids)
	}
}

func TestElements_FeatureIDs(t *testing.T) {
	es := Elements{
		&Node{ID: 1, Version: 5},
		&Way{ID: 2, Version: 6},
		&Relation{ID: 3, Version: 7},
		&Node{ID: 4, Version: 8},
	}

	expected := FeatureIDs{
		NodeID(1).FeatureID(),
		WayID(2).FeatureID(),
		RelationID(3).FeatureID(),
		NodeID(4).FeatureID(),
	}

	if ids := es.FeatureIDs(); !reflect.DeepEqual(ids, expected) {
		t.Errorf("incorrect ids: %v", ids)
	}
}

func TestElements_Sort(t *testing.T) {
	es := Elements{
		&Node{ID: 1, Version: 4},
		&Node{ID: 1, Version: 5},
		&Way{ID: 2, Version: 6},
		&Relation{ID: 3, Version: 7},
		&Way{ID: 2, Version: 5},
		&Node{ID: 4, Version: 8},
	}
	es.Sort()

	expected := ElementIDs{
		NodeID(1).ElementID(4),
		NodeID(1).ElementID(5),
		NodeID(4).ElementID(8),
		WayID(2).ElementID(5),
		WayID(2).ElementID(6),
		RelationID(3).ElementID(7),
	}

	if ids := es.ElementIDs(); !reflect.DeepEqual(ids, expected) {
		t.Errorf("incorrect sort: %v", ids)
	}
}

func TestElementIDs_Counts(t *testing.T) {
	ids := ElementIDs{
		RelationID(1).ElementID(1),
		NodeID(1).ElementID(2),
		WayID(2).ElementID(3),
		WayID(1).ElementID(2),
		RelationID(1).ElementID(1),
		WayID(1).ElementID(1),
	}

	n, w, r := ids.Counts()
	if n != 1 {
		t.Errorf("incorrect nodes: %v", n)
	}
	if w != 3 {
		t.Errorf("incorrect nodes: %v", w)
	}
	if r != 2 {
		t.Errorf("incorrect nodes: %v", r)
	}
}

func TestElementIDs_Sort(t *testing.T) {
	ids := ElementIDs{
		RelationID(1).ElementID(1),
		NodeID(1).ElementID(2),
		WayID(2).ElementID(3),
		WayID(1).ElementID(2),
		WayID(1).ElementID(1),
	}

	expected := ElementIDs{
		NodeID(1).ElementID(2),
		WayID(1).ElementID(1),
		WayID(1).ElementID(2),
		WayID(2).ElementID(3),
		RelationID(1).ElementID(1),
	}

	ids.Sort()
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("not sorted correctly")
		for i := range ids {
			t.Logf("%d: %v", i, ids[i])
		}
	}
}

func BenchmarkElementID_Sort(b *testing.B) {
	rand.Seed(1024)

	tests := make([]ElementIDs, b.N)
	for i := range tests {
		ids := make(ElementIDs, 10000)

		for j := range ids {
			v := rand.Intn(20)
			switch rand.Intn(4) {
			case 0:
				ids[j] = NodeID(rand.Int63n(int64(len(ids) / 10))).ElementID(v)
			case 1:
				ids[j] = WayID(rand.Int63n(int64(len(ids) / 10))).ElementID(v)
			case 2:
				ids[j] = RelationID(rand.Int63n(int64(len(ids) / 10))).ElementID(v)
			}
		}
		tests[i] = ids
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tests[n].Sort()
	}
}
