package osm

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestElementImplementations(t *testing.T) {
	var _ Element = &Node{}
	var _ Element = &Way{}
	var _ Element = &Relation{}
	var _ Element = &Changeset{}

	// These should not implement the Element interface
	noImplement := []interface{}{
		ElementID{},
		WayNode{},
		Member{},
	}

	for _, ni := range noImplement {
		if _, ok := ni.(Element); ok {
			t.Errorf("%T should not be an element", ni)
		}
	}
}

func TestElementIDsSort(t *testing.T) {
	ids := ElementIDs{
		{RelationType, 1},
		{ChangesetType, 1},
		{NodeType, 1},
		{WayType, 2},
		{WayType, 1},
		{ChangesetType, 3},
		{ChangesetType, 1},
	}

	expected := ElementIDs{
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

func BenchmarkElementIDSort(b *testing.B) {
	rand.Seed(1024)

	n2t := map[int]Type{
		0: NodeType,
		1: WayType,
		2: RelationType,
		3: ChangesetType,
	}

	tests := make([]ElementIDs, b.N)
	for i := range tests {
		ids := make(ElementIDs, 10000)

		for j := range ids {
			ids[j] = ElementID{
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
