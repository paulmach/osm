package core

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	osm "github.com/paulmach/go.osm"
)

var (
	start  = time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	child1 = osm.FeatureID{Type: osm.NodeType, Ref: 1}
	child2 = osm.FeatureID{Type: osm.NodeType, Ref: 2}
)

func TestCompute(t *testing.T) {
	// this is the basic test case where the first parent version
	// gets an update because there is a node update before the parents next version.
	histories := &Histories{}
	histories.Set(child1, ChildList{
		&testChild{childID: child1, versionIndex: 0, timestamp: start.Add(0 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(1 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 2, timestamp: start.Add(2 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 3, timestamp: start.Add(3 * time.Hour), visible: true},
	})

	parents := []Parent{
		&testParent{
			version: 1, visible: true,
			timestamp: start.Add(0 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 2, visible: true,
			timestamp: start.Add(2 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 3, visible: true,
			timestamp: start.Add(3 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 4, visible: true,
			timestamp: start.Add(4 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
	}

	updates, err := Compute(parents, histories, time.Minute)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{histories.Get(child1)[0]}, // index skip requires an update
		{histories.Get(child1)[2]},
		{histories.Get(child1)[3]},
		{histories.Get(child1)[3]},
	}

	expUpdates := []osm.Updates{
		{
			{Index: 0, Version: 1, Timestamp: start.Add(1 * time.Hour)},
		},
		nil, nil, nil, // no updates
	}

	compareParents(t, parents, expected)
	compareUpdates(t, updates, expUpdates)
}

func TestComputeDeletedParent(t *testing.T) {
	// Verifies behavior when a parent is deleted and then recreated.
	// The last living parent needs to have updates up to the deleted time.
	// When the parent comes back it needs to start from where its at.
	histories := &Histories{}
	histories.Set(child1, ChildList{
		&testChild{childID: child1, versionIndex: 0, timestamp: start.Add(0 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(1 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 2, timestamp: start.Add(2 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 3, timestamp: start.Add(3 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 4, timestamp: start.Add(4 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 5, timestamp: start.Add(5 * time.Hour), visible: true},
	})

	parents := []Parent{
		&testParent{
			version: 0, visible: true,
			timestamp: start.Add(0 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 1, visible: false,
			timestamp: start.Add(2 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 2, visible: true,
			timestamp: start.Add(4 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 3, visible: true,
			timestamp: start.Add(6 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
	}

	updates, err := Compute(parents, histories, time.Minute)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{histories.Get(child1)[0]},
		nil,
		{histories.Get(child1)[4]}, // index 5 is created before the 4th parent
		{histories.Get(child1)[5]},
	}

	expUpdates := []osm.Updates{
		{
			{Index: 0, Version: 1, Timestamp: start.Add(1 * time.Hour)},
		},
		nil, // no updates
		{
			{Index: 0, Version: 5, Timestamp: start.Add(5 * time.Hour)},
		},
		nil,
	}

	compareParents(t, parents, expected)
	compareUpdates(t, updates, expUpdates)
}

func TestComputeChildUpdateAfterLastParentVersion(t *testing.T) {
	// If a child is updated after the only version of a parent,
	// an update should be created.
	histories := &Histories{}
	histories.Set(child1, ChildList{
		&testChild{childID: child1, versionIndex: 0, timestamp: start.Add(0 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(1 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 2, timestamp: start.Add(2 * time.Hour), visible: true},
	})

	parents := []Parent{
		&testParent{
			version: 0, visible: true,
			timestamp: start.Add(0 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
	}

	updates, err := Compute(parents, histories, time.Minute)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{histories.Get(child1)[0]},
	}

	expUpdates := []osm.Updates{
		{
			{Index: 0, Version: 1, Timestamp: start.Add(1 * time.Hour)},
			{Index: 0, Version: 2, Timestamp: start.Add(2 * time.Hour)},
		},
	}

	compareParents(t, parents, expected)
	compareUpdates(t, updates, expUpdates)
}

func TestComputeChildUpdateRightBeforeParentDelete(t *testing.T) {
	// If a child is updated right before (based on threshold) a parent is DELETED,
	// this should create an update. This also tests a child is missing in
	// the next version. All of these histories should returned the same results.
	histories := []*Histories{}
	h := &Histories{}
	h.Set(child1, ChildList{
		&testChild{childID: child1, versionIndex: 0, timestamp: start, visible: true},
		// child is updated within threshold
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(30 * time.Second), visible: true},
	})
	histories = append(histories, h)
	h = &Histories{}
	h.Set(child1, ChildList{
		// initial child is created BEFORE parent
		&testChild{childID: child1, versionIndex: 0, timestamp: start.Add(-time.Second), visible: true},
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(30 * time.Second), visible: true},
	})
	histories = append(histories, h)
	h = &Histories{}
	h.Set(child1, ChildList{
		// initial child is created AFTER parent
		&testChild{childID: child1, versionIndex: 0, timestamp: start.Add(time.Second), visible: true},
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(30 * time.Second), visible: true},
	})
	histories = append(histories, h)

	parents := []Parent{
		&testParent{
			version: 0, visible: true,
			timestamp: start,
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 1, visible: false,
			timestamp: start.Add(time.Minute - time.Second),
		},
	}

	expUpdates := []osm.Updates{nil, nil}

	for i, h := range histories {
		t.Run(fmt.Sprintf("history %d", i), func(t *testing.T) {
			expected := []ChildList{
				{h.Get(child1)[0]},
				nil,
			}

			updates, err := Compute(parents, h, time.Minute)
			if err != nil {
				t.Fatalf("compute error: %v", err)
			}

			compareParents(t, parents, expected)
			compareUpdates(t, updates, expUpdates)
		})
	}
}

func TestComputeChildUpdateRightBeforeParentUpdated(t *testing.T) {
	// If a child is updated right before (based on threshold) a parent is UPDATED,
	// this should not trigger an updates.
	histories := &Histories{}
	histories.Set(child1, ChildList{
		&testChild{childID: child1, versionIndex: 0, timestamp: start.Add(0 * time.Hour), visible: true},
		// updated exactly 1 threshold before next parent does not create an update.
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(1*time.Hour - time.Minute), visible: true},
	})

	parents := []Parent{
		&testParent{
			version: 0, visible: true,
			timestamp: start.Add(0 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 1, visible: true,
			timestamp: start.Add(1 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
	}

	updates, err := Compute(parents, histories, time.Minute)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{histories.Get(child1)[0]},
		{histories.Get(child1)[1]},
	}

	expUpdates := []osm.Updates{nil, nil}

	compareParents(t, parents, expected)
	compareUpdates(t, updates, expUpdates)
}

func TestComputeMultipleChildren(t *testing.T) {
	// A parent with multiple children should handle each child independently.
	histories := &Histories{}
	histories.Set(child1, ChildList{
		&testChild{childID: child1, versionIndex: 0, timestamp: start.Add(0 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(1 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 2, timestamp: start.Add(5 * time.Hour), visible: true},
	})
	histories.Set(child2, ChildList{
		&testChild{childID: child2, versionIndex: 0, timestamp: start.Add(0 * time.Hour), visible: true},
		&testChild{childID: child2, versionIndex: 1, timestamp: start.Add(2 * time.Hour), visible: true},
		&testChild{childID: child2, versionIndex: 2, timestamp: start.Add(4 * time.Hour), visible: true},
	})

	parents := []Parent{
		&testParent{
			version: 0, visible: true,
			timestamp: start.Add(0 * time.Hour),
			refs:      osm.FeatureIDs{child1, child2},
		},
		&testParent{
			version: 1, visible: true,
			timestamp: start.Add(3 * time.Hour),
			refs:      osm.FeatureIDs{child1, child2},
		},
	}

	updates, err := Compute(parents, histories, time.Minute)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{histories.Get(child1)[0], histories.Get(child2)[0]},
		{histories.Get(child1)[1], histories.Get(child2)[1]},
	}

	expUpdates := []osm.Updates{
		{
			{Index: 0, Version: 1, Timestamp: start.Add(1 * time.Hour)},
			{Index: 1, Version: 1, Timestamp: start.Add(2 * time.Hour)},
		},
		{
			{Index: 0, Version: 2, Timestamp: start.Add(5 * time.Hour)},
			{Index: 1, Version: 2, Timestamp: start.Add(4 * time.Hour)},
		},
	}

	compareParents(t, parents, expected)
	compareUpdates(t, updates, expUpdates)
}

func TestComputeChangedChildList(t *testing.T) {
	// A change in the child list should be supported.
	histories := &Histories{}
	histories.Set(child1, ChildList{
		&testChild{childID: child1, versionIndex: 0, timestamp: start.Add(0 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(1 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 2, timestamp: start.Add(4 * time.Hour), visible: true},
	})
	histories.Set(child2, ChildList{
		&testChild{childID: child2, versionIndex: 0, timestamp: start.Add(0 * time.Hour), visible: true},
		&testChild{childID: child2, versionIndex: 1, timestamp: start.Add(2 * time.Hour), visible: true},
		&testChild{childID: child2, versionIndex: 2, timestamp: start.Add(3 * time.Hour), visible: true},
	})

	parents := []Parent{
		&testParent{
			version: 0, visible: true,
			timestamp: start.Add(0 * time.Hour),
			refs:      osm.FeatureIDs{child1, child2},
		},
		&testParent{
			version: 1, visible: true,
			timestamp: start.Add(2 * time.Hour),
			refs:      osm.FeatureIDs{child2},
		},
		&testParent{
			version: 2, visible: true,
			timestamp: start.Add(5 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
	}

	updates, err := Compute(parents, histories, time.Minute)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{histories.Get(child1)[0], histories.Get(child2)[0]},
		{histories.Get(child2)[1]},
		{histories.Get(child1)[2]},
	}

	expUpdates := []osm.Updates{
		{
			{Index: 0, Version: 1, Timestamp: start.Add(1 * time.Hour)},
		},
		{
			{Index: 0, Version: 2, Timestamp: start.Add(3 * time.Hour)},
		},
		nil,
	}

	compareParents(t, parents, expected)
	compareUpdates(t, updates, expUpdates)
}

func TestSetupMajorChildren(t *testing.T) {
	histories := &Histories{}
	histories.Set(child1, ChildList{
		&testChild{childID: child1, versionIndex: 0, timestamp: start, visible: true},
		&testChild{childID: child1, versionIndex: 1, timestamp: start.Add(1 * time.Hour), visible: true},
		&testChild{childID: child1, versionIndex: 2, timestamp: start.Add(2 * time.Hour), visible: false},
		&testChild{childID: child1, versionIndex: 3, timestamp: start.Add(3 * time.Hour), visible: true},
	})

	parents := []Parent{
		&testParent{
			version: 1, visible: true,
			timestamp: start.Add(0 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 2, visible: false,
			timestamp: start.Add(3 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
		&testParent{
			version: 3, visible: true,
			timestamp: start.Add(4 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
	}

	_, err := setupMajorChildren(parents, histories, time.Minute)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	expected := []ChildList{
		{histories.Get(child1)[0]},
		nil,
		{histories.Get(child1)[3]},
	}

	compareParents(t, parents, expected)

	// child is not visible for this parent's timestamp
	parents[0].(*testParent).timestamp = start.Add(-time.Hour)

	_, err = setupMajorChildren(parents, histories, time.Minute)
	if _, ok := err.(*NoVisibleChildError); !ok {
		t.Errorf("did not return correct error, got %v", err)
	}

	// one of the child's histories was not provided
	parents[0].(*testParent).timestamp = start

	histories.Set(osm.FeatureID{Type: osm.NodeType, Ref: 2}, histories.Get(child1))
	histories.Set(child1, nil)

	_, err = setupMajorChildren(parents, histories, time.Minute)
	if _, ok := err.(*NoHistoryError); !ok {
		t.Errorf("Did not return correct error, got %v", err)
	}
}

func compareParents(t *testing.T, parents []Parent, expected []ChildList) {
	for i, p := range parents {
		parent := p.(*testParent)

		if expected[i] == nil {
			if parent.children != nil {
				t.Errorf("expected no children for %d", i)
				t.Logf("got: %+v", parent.children)
			}

			continue
		}

		if expected[i] != nil && parent.children == nil {
			t.Errorf("got no children for %d", i)
			t.Logf("expected: %+v", expected[i])

			continue
		}

		if parent.children[0] != expected[i][0] {
			t.Errorf("incorrect at parent %d", i)
			t.Logf("%+v", parent.children)
			t.Logf("%+v", expected[i])
		}
	}
}

func compareUpdates(t *testing.T, updates, expected []osm.Updates) {
	if !reflect.DeepEqual(updates, expected) {
		t.Errorf("updates not equal")

		if len(updates) != len(expected) {
			// length should be the length of the parents
			t.Fatalf("length of updates mismatch, %d != %d", len(updates), len(expected))
		}

		for i := range updates {
			if !reflect.DeepEqual(updates[i], expected[i]) {
				t.Errorf("index %d not equal", i)

				for j := range updates[i] {
					if !reflect.DeepEqual(updates[i][j], expected[i][j]) {
						t.Errorf("sub-index %d not equal", j)
						t.Logf("%+v", updates[i][j])
						t.Logf("%+v", expected[i][j])
					}
				}
			}
		}
	}
}
