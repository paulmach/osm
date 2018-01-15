package osm

import (
	"math/rand"
	"reflect"
	"testing"
)

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
			name:     "changeset",
			id:       ChangesetID(10000).FeatureID(),
			expected: "changeset/10000",
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

func TestFeatureID_ParseFeatureID(t *testing.T) {
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
}

func TestFeatureIDs_Sort(t *testing.T) {
	ids := FeatureIDs{
		RelationID(1).FeatureID(),
		ChangesetID(1).FeatureID(),
		NodeID(1).FeatureID(),
		WayID(2).FeatureID(),
		WayID(1).FeatureID(),
		ChangesetID(3).FeatureID(),
		ChangesetID(1).FeatureID(),
	}

	expected := FeatureIDs{
		NodeID(1).FeatureID(),
		WayID(1).FeatureID(),
		WayID(2).FeatureID(),
		RelationID(1).FeatureID(),
		ChangesetID(1).FeatureID(),
		ChangesetID(1).FeatureID(),
		ChangesetID(3).FeatureID(),
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
			case 3:
				ids[j] = ChangesetID(rand.Int63n(int64(len(ids) / 10))).FeatureID()
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
