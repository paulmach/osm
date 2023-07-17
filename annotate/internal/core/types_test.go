package core

import (
	"testing"
	"time"

	"github.com/onXmaps/osm"
	"github.com/onXmaps/osm/annotate/shared"
)

type findVisibleTestCase struct {
	name      string
	timestamp time.Time
	threshold time.Duration
	index     int
}

type lastVisibleTestCase struct {
	name      string
	timestamp time.Time
	index     int
}

func TestChildListFindVisible(t *testing.T) {
	cl := ChildList{
		&shared.Child{
			VersionIndex: 0, Visible: true,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 30, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 1, Visible: true,
			Timestamp: time.Date(2016, 1, 2, 0, 0, 30, 0, time.UTC)},
	}

	cases := []findVisibleTestCase{
		{
			timestamp: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     -1,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
			threshold: time.Second,
			index:     -1,
		}, {
			timestamp: time.Date(2016, 1, 1, 23, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 10, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		}, {
			timestamp: time.Date(2016, 1, 2, 0, 0, 30, 0, time.UTC),
			threshold: 0,
			index:     1,
		}, {
			timestamp: time.Date(2016, 1, 2, 0, 0, 30, 0, time.UTC),
			threshold: time.Minute,
			index:     1,
		}, {
			timestamp: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     1,
		},
	}

	checkChildListFindVisible(t, 0, cl, cases)
}

func TestChildListFindVisibleCommitted(t *testing.T) {
	cl := ChildList{
		&shared.Child{
			VersionIndex: 0, Visible: true,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 30, 0, time.UTC),
			Committed: time.Date(2016, 1, 1, 5, 0, 30, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 1, Visible: true,
			Timestamp: time.Date(2016, 1, 2, 0, 0, 30, 0, time.UTC),
			Committed: time.Date(2016, 1, 2, 5, 0, 30, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 2, Visible: false,
			Timestamp: time.Date(2016, 1, 3, 0, 0, 30, 0, time.UTC),
			Committed: time.Date(2016, 1, 3, 5, 0, 30, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 3, Visible: true,
			Timestamp: time.Date(2016, 1, 4, 0, 0, 30, 0, time.UTC),
			Committed: time.Date(2016, 1, 4, 5, 0, 30, 0, time.UTC)},
	}

	cases := []findVisibleTestCase{
		{
			name:      "before all should return first",
			timestamp: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     -1,
		}, {
			name:      "before within threshold should not matter",
			timestamp: time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     -1,
		}, {
			name:      "before within threshold of committed should not matter",
			timestamp: time.Date(2016, 1, 1, 5, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     -1,
		}, {
			name:      "on committed should return child",
			timestamp: time.Date(2016, 1, 1, 5, 0, 30, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		}, {
			name:      "after hidden timestamp but not committed",
			timestamp: time.Date(2016, 1, 3, 0, 0, 30, 0, time.UTC),
			threshold: time.Minute,
			index:     1,
		}, {
			name:      "after non-visible's committed time",
			timestamp: time.Date(2016, 1, 3, 9, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     -1,
		}, {
			name:      "after last element should return latest",
			timestamp: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     3,
		},
	}

	checkChildListFindVisible(t, 0, cl, cases)
}

func TestChildListFindVisibleWithHidden(t *testing.T) {
	cl := ChildList{
		&shared.Child{
			VersionIndex: 0, Visible: false,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 30, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 1, Visible: true,
			Timestamp: time.Date(2016, 1, 2, 0, 0, 30, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 2, Visible: false,
			Timestamp: time.Date(2016, 1, 3, 0, 0, 30, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 3, Visible: true,
			Timestamp: time.Date(2016, 1, 4, 0, 0, 30, 0, time.UTC)},
	}

	cases := []findVisibleTestCase{
		{
			name:      "before all",
			timestamp: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     -1,
		}, {
			name:      "after first hidden element",
			timestamp: time.Date(2016, 1, 1, 10, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     -1,
		}, {
			name:      "within threshold of first visible",
			timestamp: time.Date(2016, 1, 2, 0, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     1,
		}, {
			timestamp: time.Date(2016, 1, 3, 0, 0, 0, 0, time.UTC),
			threshold: 0,
			index:     1,
		}, {
			name:      "within threshold of hidden",
			timestamp: time.Date(2016, 1, 3, 0, 0, 30, 0, time.UTC),
			threshold: time.Second,
			index:     1,
		}, {
			name:      "all of threshold internval on or after hidden",
			timestamp: time.Date(2016, 1, 3, 0, 0, 31, 0, time.UTC),
			threshold: time.Second,
			index:     -1,
		}, {
			name:      "well after hidden should return nil/-1",
			timestamp: time.Date(2016, 1, 3, 23, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     -1,
		}, {
			name:      "well after last visible element",
			timestamp: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			threshold: time.Minute,
			index:     3,
		},
	}

	checkChildListFindVisible(t, 0, cl, cases)
}

func TestChildListFindVisibleWithinThreshold(t *testing.T) {
	cl := ChildList{
		&shared.Child{
			VersionIndex: 0, Visible: true,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 10, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 1, Visible: true,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 20, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 2, Visible: true,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 30, 0, time.UTC)},
	}

	cases := []findVisibleTestCase{
		{
			name:      "nearest one within threshold",
			timestamp: time.Date(2016, 1, 1, 0, 0, 14, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		}, {
			name:      "if equal distant, should be later",
			timestamp: time.Date(2016, 1, 1, 0, 0, 15, 0, time.UTC),
			threshold: time.Minute,
			index:     1,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 16, 0, time.UTC),
			threshold: time.Minute,
			index:     1,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 21, 0, time.UTC),
			threshold: time.Minute,
			index:     1,
		}, {
			name:      "nothing within threshold should find previous visible",
			timestamp: time.Date(2016, 1, 1, 0, 0, 29, 0, time.UTC),
			threshold: 0,
			index:     1,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 29, 0, time.UTC),
			threshold: time.Second,
			index:     2,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 50, 0, time.UTC),
			threshold: 0,
			index:     2,
		},
	}

	checkChildListFindVisible(t, 0, cl, cases)
}

func TestChildListFindVisibleHiddenWithinThreshold(t *testing.T) {
	cl := ChildList{
		&shared.Child{
			VersionIndex: 0, Visible: true,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 10, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 1, Visible: false,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 20, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 2, Visible: true,
			Timestamp: time.Date(2016, 1, 1, 0, 2, 0, 0, time.UTC)},
	}

	cases := []findVisibleTestCase{
		{
			timestamp: time.Date(2016, 1, 1, 0, 0, 14, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 15, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 16, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 16, 0, time.UTC),
			threshold: time.Second,
			index:     0,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 21, 0, time.UTC),
			threshold: time.Second,
			index:     -1,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 21, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		},

		// with larger thresholds
		{
			timestamp: time.Date(2016, 1, 1, 0, 0, 14, 0, time.UTC),
			threshold: 10 * time.Minute,
			index:     0,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 1, 30, 0, time.UTC),
			threshold: 10 * time.Minute,
			index:     2,
		}, {
			timestamp: time.Date(2016, 1, 1, 0, 0, 23, 0, time.UTC),
			threshold: 10 * time.Minute,
			index:     0,
		},
	}

	checkChildListFindVisible(t, 0, cl, cases)
}

func TestChildListFindVisibleDifferentChangeset(t *testing.T) {
	cl := ChildList{
		&shared.Child{
			VersionIndex: 0, Visible: true, ChangesetID: 1,
			Timestamp: time.Date(2010, 1, 1, 0, 0, 10, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 1, Visible: true, ChangesetID: 2,
			Timestamp: time.Date(2010, 1, 1, 0, 0, 20, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 2, Visible: true, ChangesetID: 1,
			Timestamp: time.Date(2010, 1, 1, 0, 0, 30, 0, time.UTC)},
	}

	cases := []findVisibleTestCase{
		{
			name:      "not match next if different changeset",
			timestamp: time.Date(2010, 1, 1, 0, 0, 19, 0, time.UTC),
			threshold: time.Minute,
			index:     0,
		}, {
			name:      "match previous is different changeset",
			timestamp: time.Date(2010, 1, 1, 0, 0, 21, 0, time.UTC),
			threshold: time.Second,
			index:     1,
		}, {
			name:      "should consider only closest change",
			timestamp: time.Date(2010, 1, 1, 0, 0, 21, 0, time.UTC),
			threshold: time.Minute,
			index:     1,
		}, {
			name:      "include next if same changeset",
			timestamp: time.Date(2010, 1, 1, 0, 0, 29, 0, time.UTC),
			threshold: time.Minute,
			index:     2,
		},
	}

	checkChildListFindVisible(t, 1, cl, cases)
}

func checkChildListFindVisible(t *testing.T, id osm.ChangesetID, cl ChildList, cases []findVisibleTestCase) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := cl.FindVisible(id, tc.timestamp, tc.threshold)
			if c == nil {
				if tc.index != -1 {
					t.Errorf("should not be nil, should be %d", tc.index)
					t.Logf("%+v", tc)
				}
			} else if c != nil && tc.index == -1 {
				t.Errorf("should be nil, got %v", c.VersionIndex)
				t.Logf("%+v", tc)
			} else if idx := c.VersionIndex; idx != tc.index {
				t.Errorf("should be %d, got %v", tc.index, idx)
				t.Logf("%+v", tc)
			}
		})
	}
}

func TestChildListVersionBefore(t *testing.T) {
	cl := ChildList{
		&shared.Child{
			VersionIndex: 0, Visible: false,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 10, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 1, Visible: true,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 20, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 2, Visible: false,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 30, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 3, Visible: true,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 30, 0, time.UTC),
			Committed: time.Date(2016, 1, 1, 0, 0, 40, 0, time.UTC)},
		&shared.Child{
			VersionIndex: 4, Visible: false,
			Timestamp: time.Date(2016, 1, 1, 0, 0, 50, 0, time.UTC)},
	}

	cases := []lastVisibleTestCase{
		{
			name:      "before first time",
			timestamp: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
			index:     -1,
		}, {
			name:      "right on time should return previous",
			timestamp: time.Date(2016, 1, 1, 0, 0, 20, 0, time.UTC),
			index:     0,
		}, {
			name:      "after first visible",
			timestamp: time.Date(2016, 1, 1, 0, 0, 20, 1, time.UTC),
			index:     1,
		}, {
			name:      "after all should return last",
			timestamp: time.Date(2016, 1, 1, 0, 0, 70, 0, time.UTC),
			index:     4,
		}, {
			name:      "after timestamp but before committed should return previous",
			timestamp: time.Date(2016, 1, 1, 0, 0, 40, 0, time.UTC),
			index:     2,
		}, {
			name:      "right after committed time should return element",
			timestamp: time.Date(2016, 1, 1, 0, 0, 40, 1, time.UTC),
			index:     3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			c := cl.VersionBefore(tc.timestamp)
			if c == nil {
				if tc.index != -1 {
					t.Errorf("should not be nil, should be %d", tc.index)
					t.Logf("%+v", tc)
				}
			} else if c != nil && tc.index == -1 {
				t.Errorf("should be nil, got %v", c.VersionIndex)
				t.Logf("%+v", tc)
			} else if idx := c.VersionIndex; idx != tc.index {
				t.Errorf("should be %d, got %v", tc.index, idx)
				t.Logf("%+v", tc)
			}
		})
	}
}

var _ Parent = &testParent{}

type testParent struct {
	changesetID osm.ChangesetID
	version     int
	visible     bool
	timestamp   time.Time
	committed   time.Time
	refs        osm.FeatureIDs
	children    ChildList
}

func (t testParent) ID() osm.FeatureID {
	return osm.FeatureID(0) // this is only used for logging.
}

func (t testParent) ChangesetID() osm.ChangesetID {
	return t.changesetID
}

func (t testParent) Version() int {
	return t.version
}

func (t testParent) Visible() bool {
	return t.visible
}

func (t testParent) Timestamp() time.Time {
	return t.timestamp
}

func (t testParent) Committed() time.Time {
	return t.committed
}

func (t testParent) Refs() (osm.FeatureIDs, []bool) {
	annotated := make([]bool, len(t.refs))
	for i := range annotated {
		annotated[i] = true
	}
	return t.refs, annotated
}

func (t *testParent) SetChild(idx int, c *shared.Child) {
	if idx >= len(t.children) {
		nc := make(ChildList, idx+1)
		copy(nc, t.children)
		t.children = nc
	}
	t.children[idx] = c
}
