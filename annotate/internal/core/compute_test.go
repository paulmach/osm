package core

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/onXmaps/osm"
	"github.com/onXmaps/osm/annotate/shared"
)

var (
	start  = time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	child1 = osm.NodeID(1).FeatureID()
	child2 = osm.NodeID(2).FeatureID()
)

func TestCompute(t *testing.T) {
	ctx := context.Background()

	// this is the basic test case where the first parent version
	// gets an update because there is a node update before the parents next version.
	ds := &TestDS{}
	ds.Set(child1, ChildList{
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start.Add(0 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(1 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 2, VersionIndex: 2, Timestamp: start.Add(2 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 3, VersionIndex: 3, Timestamp: start.Add(3 * time.Hour), Visible: true},
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

	updates, err := Compute(ctx, parents, ds, nil)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{ds.MustGet(child1)[0]}, // index skip requires an update
		{ds.MustGet(child1)[2]},
		{ds.MustGet(child1)[3]},
		{ds.MustGet(child1)[3]},
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

func TestCompute_MissingChildren(t *testing.T) {
	ctx := context.Background()

	ds := &TestDS{}
	ds.Set(child1, ChildList{
		&shared.Child{ID: child1, VersionIndex: 0, Timestamp: start.Add(0 * time.Hour), Visible: true},
	})

	parents := []Parent{
		&testParent{
			version: 0, visible: true,
			timestamp: start.Add(0 * time.Hour),
			refs:      osm.FeatureIDs{child1, osm.FeatureID(0)},
		},
	}

	opts := &Options{Threshold: time.Minute, IgnoreMissingChildren: true}
	_, err := Compute(ctx, parents, ds, opts)
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{ds.MustGet(child1)[0]},
		nil,
	}

	compareParents(t, parents, expected)
}

func TestCompute_DeletedParent(t *testing.T) {
	ctx := context.Background()

	// Verifies behavior when a parent is deleted and then recreated.
	// The last living parent needs to have updates up to the deleted time.
	// When the parent comes back it needs to start from where its at.
	ds := &TestDS{}
	ds.Set(child1, ChildList{
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start.Add(0 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(1 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 2, VersionIndex: 2, Timestamp: start.Add(2 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 3, VersionIndex: 3, Timestamp: start.Add(3 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 4, VersionIndex: 4, Timestamp: start.Add(4 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 5, VersionIndex: 5, Timestamp: start.Add(5 * time.Hour), Visible: true},
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

	updates, err := Compute(ctx, parents, ds, &Options{Threshold: time.Minute})
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{ds.MustGet(child1)[0]},
		nil,
		{ds.MustGet(child1)[4]}, // index 5 is created before the 4th parent
		{ds.MustGet(child1)[5]},
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

func TestCompute_ChildUpdateAfterLastParentVersion(t *testing.T) {
	ctx := context.Background()

	// If a child is updated after the only version of a parent,
	// an update should be created.
	ds := &TestDS{}
	ds.Set(child1, ChildList{
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start.Add(0 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(1 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 2, VersionIndex: 2, Timestamp: start.Add(2 * time.Hour), Visible: true},
	})

	parents := []Parent{
		&testParent{
			version: 0, visible: true,
			timestamp: start.Add(0 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
	}

	updates, err := Compute(ctx, parents, ds, &Options{Threshold: time.Minute})
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{ds.MustGet(child1)[0]},
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

func TestCompute_ChildUpdateRightBeforeParentDelete(t *testing.T) {
	ctx := context.Background()

	// If a child is updated right before (based on threshold) a parent is DELETED,
	// this should create an update. This also tests a child is missing in
	// the next version. All of these histories should returned the same results.
	dss := []*TestDS{}
	ds := &TestDS{}
	ds.Set(child1, ChildList{
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start, Visible: true},
		// child is updated within threshold
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(30 * time.Second), Visible: true},
	})
	dss = append(dss, ds)
	ds = &TestDS{}
	ds.Set(child1, ChildList{
		// initial child is created BEFORE parent
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start.Add(-time.Second), Visible: true},
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(30 * time.Second), Visible: true},
	})
	dss = append(dss, ds)
	ds = &TestDS{}
	ds.Set(child1, ChildList{
		// initial child is created AFTER parent
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start.Add(time.Second), Visible: true},
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(30 * time.Second), Visible: true},
	})
	dss = append(dss, ds)

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

	for i, ds := range dss {
		t.Run(fmt.Sprintf("history %d", i), func(t *testing.T) {
			expected := []ChildList{
				{ds.MustGet(child1)[0]},
				nil,
			}

			updates, err := Compute(ctx, parents, ds, &Options{Threshold: time.Minute})
			if err != nil {
				t.Fatalf("compute error: %v", err)
			}

			compareParents(t, parents, expected)
			compareUpdates(t, updates, expUpdates)
		})
	}
}

func TestCompute_ChildUpdateRightBeforeParentUpdated(t *testing.T) {
	ctx := context.Background()

	// If a child is updated right before (based on threshold) a parent is UPDATED,
	// this should not trigger an updates.
	ds := &TestDS{}
	ds.Set(child1, ChildList{
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start.Add(0 * time.Hour), Visible: true},
		// updated exactly 1 threshold before next parent does not create an update.
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(1*time.Hour - time.Minute), Visible: true},
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

	updates, err := Compute(ctx, parents, ds, &Options{Threshold: time.Minute})
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{ds.MustGet(child1)[0]},
		{ds.MustGet(child1)[1]},
	}

	expUpdates := []osm.Updates{nil, nil}

	compareParents(t, parents, expected)
	compareUpdates(t, updates, expUpdates)
}

func TestCompute_MultipleChildren(t *testing.T) {
	ctx := context.Background()

	// A parent with multiple children should handle each child independently.
	ds := &TestDS{}
	ds.Set(child1, ChildList{
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start.Add(0 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(1 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 2, VersionIndex: 2, Timestamp: start.Add(5 * time.Hour), Visible: true},
	})
	ds.Set(child2, ChildList{
		&shared.Child{ID: child2, Version: 0, VersionIndex: 0, Timestamp: start.Add(0 * time.Hour), Visible: true},
		&shared.Child{ID: child2, Version: 1, VersionIndex: 1, Timestamp: start.Add(2 * time.Hour), Visible: true},
		&shared.Child{ID: child2, Version: 2, VersionIndex: 2, Timestamp: start.Add(4 * time.Hour), Visible: true},
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

	updates, err := Compute(ctx, parents, ds, &Options{Threshold: time.Minute})
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{ds.MustGet(child1)[0], ds.MustGet(child2)[0]},
		{ds.MustGet(child1)[1], ds.MustGet(child2)[1]},
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

func TestCompute_ChangedChildList(t *testing.T) {
	ctx := context.Background()

	// A change in the child list should be supported.
	ds := &TestDS{}
	ds.Set(child1, ChildList{
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start.Add(0 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(1 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 2, VersionIndex: 2, Timestamp: start.Add(4 * time.Hour), Visible: true},
	})
	ds.Set(child2, ChildList{
		&shared.Child{ID: child2, Version: 0, VersionIndex: 0, Timestamp: start.Add(0 * time.Hour), Visible: true},
		&shared.Child{ID: child2, Version: 1, VersionIndex: 1, Timestamp: start.Add(2 * time.Hour), Visible: true},
		&shared.Child{ID: child2, Version: 2, VersionIndex: 2, Timestamp: start.Add(3 * time.Hour), Visible: true},
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

	updates, err := Compute(ctx, parents, ds, &Options{Threshold: time.Minute})
	if err != nil {
		t.Fatalf("compute error: %v", err)
	}

	expected := []ChildList{
		{ds.MustGet(child1)[0], ds.MustGet(child2)[0]},
		{ds.MustGet(child2)[1]},
		{ds.MustGet(child1)[2]},
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

func TestCompute_MajorChildren(t *testing.T) {
	ctx := context.Background()

	ds := &TestDS{}
	ds.Set(child1, ChildList{
		&shared.Child{ID: child1, Version: 0, VersionIndex: 0, Timestamp: start, Visible: true},
		&shared.Child{ID: child1, Version: 1, VersionIndex: 1, Timestamp: start.Add(1 * time.Hour), Visible: true},
		&shared.Child{ID: child1, Version: 2, VersionIndex: 2, Timestamp: start.Add(3 * time.Hour), Visible: false},
		&shared.Child{ID: child1, Version: 3, VersionIndex: 3, Timestamp: start.Add(5 * time.Hour), Visible: true},
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
			timestamp: start.Add(6 * time.Hour),
			refs:      osm.FeatureIDs{child1},
		},
	}

	_, err := Compute(ctx, parents, ds, &Options{Threshold: time.Minute})
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	expected := []ChildList{
		{ds.MustGet(child1)[0]},
		nil,
		{ds.MustGet(child1)[3]},
	}

	compareParents(t, parents, expected)

	// child is not visible for this parent's timestamp
	parents[0].(*testParent).timestamp = start.Add(-time.Hour)

	_, err = Compute(ctx, parents, ds, &Options{Threshold: time.Minute})
	if _, ok := err.(*NoVisibleChildError); !ok {
		t.Errorf("did not return correct error, got %v", err)
	}

	// one of the child's histories was not provided
	parents[0].(*testParent).timestamp = start

	ds.Set(osm.NodeID(2).FeatureID(), ds.MustGet(child1))
	ds.Set(child1, nil)

	_, err = Compute(ctx, parents, ds, &Options{Threshold: time.Minute})
	if _, ok := err.(*NoHistoryError); !ok {
		t.Errorf("Did not return correct error, got %v", err)
	}
}

func TestChildLocs_GroupByParent(t *testing.T) {
	in := childLocs{
		{Parent: 1, Index: 1},
		{Parent: 2, Index: 2},
		{Parent: 4, Index: 3},
		{Parent: 4, Index: 3},
		{Parent: 4, Index: 4},
		{Parent: 3, Index: 6},
		{Parent: 3, Index: 6},
		{Parent: 7, Index: 8},
	}

	expected := []childLocs{
		in[0:1],
		in[1:2],
		in[2:5],
		in[5:7],
		in[7:8],
	}

	out := in.GroupByParent()
	if !reflect.DeepEqual(out, expected) {
		t.Errorf("incorrect sets: %v", out)
		t.Errorf("expected:       %v", expected)
	}
}

func compareParents(t *testing.T, parents []Parent, expected []ChildList) {
	t.Helper()

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
	t.Helper()

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
