package osm

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestElementID_ParseElementID(t *testing.T) {
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
			name: "way",
			id:   WayID(10).ElementID(2),
		},
		{
			name: "relation",
			id:   RelationID(100).ElementID(3),
		},
		{
			name: "changeset",
			id:   RelationID(1000).ElementID(4),
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
}

func TestElementImplementations(t *testing.T) {
	var _ Element = &Node{}
	var _ Element = &Way{}
	var _ Element = &Relation{}
	var _ Element = &Changeset{}

	// These should not implement the Element interface
	noImplement := []interface{}{
		FeatureID(0),
		ElementID(0),
		WayNode{},
		Member{},
		NodeID(0),
		WayID(0),
		RelationID(0),
		ChangesetID(0),
	}

	for _, ni := range noImplement {
		if _, ok := ni.(Element); ok {
			t.Errorf("%T should not be an element", ni)
		}
	}
}

func TestElementIDsSort(t *testing.T) {
	ids := ElementIDs{
		RelationID(1).ElementID(1),
		ChangesetID(1).ElementID(),
		NodeID(1).ElementID(2),
		WayID(2).ElementID(3),
		WayID(1).ElementID(2),
		WayID(1).ElementID(1),
		ChangesetID(3).ElementID(),
	}

	expected := ElementIDs{
		NodeID(1).ElementID(2),
		WayID(1).ElementID(1),
		WayID(1).ElementID(2),
		WayID(2).ElementID(3),
		RelationID(1).ElementID(1),
		ChangesetID(1).ElementID(),
		ChangesetID(3).ElementID(),
	}

	ids.Sort()
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("not sorted correctly")
		for i := range ids {
			t.Logf("%d: %v", i, ids[i])
		}
	}
}

func BenchmarkElementIDSort(b *testing.B) {
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
			case 3:
				ids[j] = ChangesetID(rand.Int63n(int64(len(ids) / 10))).ElementID()
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
