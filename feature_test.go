package osm

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestFeatureID_ids(t *testing.T) {
	id := NodeID(1).FeatureID()

	oid := id.ObjectID(10)
	if v := oid.Type(); v != TypeNode {
		t.Errorf("incorrect type: %v", v)
	}

	if v := oid.Ref(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	if v := oid.Version(); v != 10 {
		t.Errorf("incorrect version: %v", v)
	}

	eid := id.ElementID(10)
	if v := eid.Type(); v != TypeNode {
		t.Errorf("incorrect type: %v", v)
	}

	if v := eid.Ref(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	if v := eid.Version(); v != 10 {
		t.Errorf("incorrect version: %v", v)
	}

	if v := NodeID(1).FeatureID().NodeID(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	if v := WayID(1).FeatureID().WayID(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	if v := RelationID(1).FeatureID().RelationID(); v != 1 {
		t.Errorf("incorrect id: %v", v)
	}

	t.Run("not a node", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("should panic?")
			}
		}()

		id := WayID(1).FeatureID()
		id.NodeID()
	})

	t.Run("not a way", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("should panic?")
			}
		}()

		id := NodeID(1).FeatureID()
		id.WayID()
	})

	t.Run("not a relation", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("should panic?")
			}
		}()

		id := WayID(1).FeatureID()
		id.RelationID()
	})

	t.Run("should not panic if invalid type", func(t *testing.T) {
		var id FeatureID
		if v := id.Type(); v != "" {
			t.Errorf("should return empty string for invalid type: %v", v)
		}
	})
}

func TestFeature_String(t *testing.T) {
	cases := []struct {
		name     string
		id       FeatureID
		expected string
	}{
		{
			name:     "node",
			id:       NodeID(1).FeatureID(),
			expected: "node/1",
		},
		{
			name:     "way",
			id:       WayID(3).FeatureID(),
			expected: "way/3",
		},
		{
			name:     "relation",
			id:       RelationID(1000).FeatureID(),
			expected: "relation/1000",
		},
		{
			name:     "unknown",
			id:       0,
			expected: "unknown/0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if v := tc.id.String(); v != tc.expected {
				t.Errorf("incorrect string: %v", v)
			}
		})
	}
}

func TestParseFeatureID(t *testing.T) {
	cases := []struct {
		name string
		id   FeatureID
	}{
		{
			name: "node",
			id:   NodeID(0).FeatureID(),
		},
		{
			name: "way",
			id:   WayID(10).FeatureID(),
		},
		{
			name: "relation",
			id:   RelationID(100).FeatureID(),
		},
		{
			name: "changeset",
			id:   RelationID(1000).FeatureID(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := ParseFeatureID(tc.id.String())
			if err != nil {
				t.Errorf("parse error: %v", err)
			}

			if id != tc.id {
				t.Errorf("incorrect id: %v != %v", id, tc.id)
			}
		})
	}

	// errors
	if _, err := ParseFeatureID("123"); err == nil {
		t.Errorf("should return error if only one part")
	}

	if _, err := ParseFeatureID("node/abc"); err == nil {
		t.Errorf("should return error if id not a number")
	}

	if _, err := ParseFeatureID("lake/1"); err == nil {
		t.Errorf("should return error if not a valid type")
	}
}

func TestFeatureIDs_Counts(t *testing.T) {
	ids := FeatureIDs{
		RelationID(1).FeatureID(),
		NodeID(1).FeatureID(),
		WayID(2).FeatureID(),
		WayID(1).FeatureID(),
		RelationID(1).FeatureID(),
		WayID(1).FeatureID(),
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

func TestFeatureIDs_Sort(t *testing.T) {
	ids := FeatureIDs{
		RelationID(1).FeatureID(),
		NodeID(1).FeatureID(),
		WayID(2).FeatureID(),
		WayID(1).FeatureID(),
	}

	expected := FeatureIDs{
		NodeID(1).FeatureID(),
		WayID(1).FeatureID(),
		WayID(2).FeatureID(),
		RelationID(1).FeatureID(),
	}

	ids.Sort()
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("not sorted correctly")
		for i := range ids {
			t.Logf("%d: %v", i, ids[i])
		}
	}
}

func BenchmarkFeatureIDs_Sort(b *testing.B) {
	rand.Seed(1024)

	tests := make([]FeatureIDs, b.N)
	for i := range tests {
		ids := make(FeatureIDs, 10000)

		for j := range ids {
			switch rand.Intn(4) {
			case 0:
				ids[j] = NodeID(rand.Int63n(int64(len(ids) / 10))).FeatureID()
			case 1:
				ids[j] = WayID(rand.Int63n(int64(len(ids) / 10))).FeatureID()
			case 2:
				ids[j] = RelationID(rand.Int63n(int64(len(ids) / 10))).FeatureID()
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
