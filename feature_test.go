package osm

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestFeatureIDsSort(t *testing.T) {
	ids := FeatureIDs{
		{RelationType, 1},
		{ChangesetType, 1},
		{NodeType, 1},
		{WayType, 2},
		{WayType, 1},
		{ChangesetType, 3},
		{ChangesetType, 1},
	}

	expected := FeatureIDs{
		{NodeType, 1},
		{WayType, 1},
		{WayType, 2},
		{RelationType, 1},
		{ChangesetType, 1},
		{ChangesetType, 1},
		{ChangesetType, 3},
	}

	ids.Sort()
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("not sorted correctly")
		for i := range ids {
			t.Logf("%d: %v", i, ids[i])
		}
	}
}

func BenchmarkFeatureIDsSort(b *testing.B) {
	rand.Seed(1024)

	n2t := map[int]Type{
		0: NodeType,
		1: WayType,
		2: RelationType,
		3: ChangesetType,
	}

	tests := make([]FeatureIDs, b.N)
	for i := range tests {
		ids := make(FeatureIDs, 10000)

		for j := range ids {
			ids[j] = FeatureID{
				Type: n2t[rand.Intn(len(n2t))],
				Ref:  rand.Int63n(int64(len(ids) / 10)),
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
