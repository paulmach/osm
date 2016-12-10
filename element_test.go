package osm

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestElementIDsSort(t *testing.T) {
	ids := []ElementID{
		{RelationType, 1, 1},
		{ChangesetType, 1, 5},
		{NodeType, 1, 1},
		{WayType, 2, 1},
		{WayType, 1, 1},
		{ChangesetType, 3, 1},
		{ChangesetType, 1, 1},
	}

	expected := []ElementID{
		{NodeType, 1, 1},
		{WayType, 1, 1},
		{WayType, 2, 1},
		{RelationType, 1, 1},
		{ChangesetType, 1, 1},
		{ChangesetType, 1, 5},
		{ChangesetType, 3, 1},
	}

	ElementIDs(ids).Sort()
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("not sorted correctly")
		for i := range ids {
			t.Logf("%d: %v", i, ids[i])
		}
	}
}

func BenchmarkElementIDSort(b *testing.B) {
	rand.Seed(1024)

	n2t := map[int]ElementType{
		0: NodeType,
		1: WayType,
		2: RelationType,
		3: ChangesetType,
	}

	tests := make([][]ElementID, b.N)
	for i := range tests {
		ids := make([]ElementID, 10000)

		for j := range ids {
			ids[j] = ElementID{
				Type:    n2t[rand.Intn(len(n2t))],
				ID:      rand.Int63n(int64(len(ids) / 10)),
				Version: rand.Intn(20),
			}
		}
		tests[i] = ids
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ElementIDs(tests[n]).Sort()
	}
}
