package core

import (
	"time"

	"github.com/onXmaps/osm"
	"github.com/onXmaps/osm/annotate/shared"
)

// A Parent is something that holds children. ie. ways have nodes as children
// and relations can have nodes, ways and relations as children.
type Parent interface {
	ID() osm.FeatureID // used for logging
	ChangesetID() osm.ChangesetID

	Version() int
	Visible() bool
	Timestamp() time.Time
	Committed() time.Time

	// Refs returns normalized information about the children.
	// Currently this is the feature ids and if it is already annotated.
	// Note: we auto-annotate all unannotated children if they would have
	// been filtered out.
	Refs() (osm.FeatureIDs, []bool)
	SetChild(idx int, c *shared.Child)
}

// A ChildList is a set
type ChildList []*shared.Child

// FindVisible locates the child visible at the given time.
// If 'at' is on or after osm.CommitInfoStart the committed
// time is used to determine visiblity. If 'at' is before, a range +-eps
// around the give time. Will return the closes visible node within that
// range, or the previous node if visible. Children after 'at' but within
// the eps must have the same changeset id as provided (the parent's).
// If the previous node is not visible, or does not exits will return nil.
func (cl ChildList) FindVisible(cid osm.ChangesetID, at time.Time, eps time.Duration) *shared.Child {
	var (
		diff    time.Duration = -1
		nearest *shared.Child
	)

	start := at.Add(-eps)
	for _, c := range cl {

		if c.Committed.Before(osm.CommitInfoStart) {
			// more complicated logic for early data.
			offset := c.Timestamp.Sub(start)
			visible := c.Visible

			// if this node is after the end then it's over
			if offset > 2*eps {
				break
			}

			// if we're before the start set with the latest node
			if offset < 0 {
				if visible {
					nearest = c
				} else {
					nearest = nil
				}

				continue
			}

			// we're in the range!!!
			d := absDuration(offset - eps)
			if diff < 0 || (d <= diff) {
				// first within range, set if not visible
				if diff == -1 && !visible && offset == 0 {
					nearest = nil
				}

				// only update nearest if visible since we want
				// the closest visible within the range.
				if visible {
					if offset <= eps {
						// if we're before at, pick it
						nearest = c
					} else if c.ChangesetID == cid {
						// if we're after at, changeset must be same
						nearest = c
					} else {
						// after at, not same changeset, ignore.
						continue
					}
				}

				diff = d
			}
		} else {
			// simpler logic, if committed is on or before 'at'
			// consider that element.
			if c.Committed.After(at) {
				break
			}

			if c.Visible {
				nearest = c
			} else {
				nearest = nil
			}
		}
	}

	return nearest
}

// VersionBefore finds the last child before a given time.
func (cl ChildList) VersionBefore(end time.Time) *shared.Child {
	var latest *shared.Child

	for _, c := range cl {
		if !timeThreshold(c, 0).Before(end) {
			break
		}

		latest = c
	}

	return latest
}

func timeThreshold(c *shared.Child, esp time.Duration) time.Time {
	if c.Committed.Before(osm.CommitInfoStart) {
		return c.Timestamp.Add(esp)
	}

	return c.Committed
}

func timeThresholdParent(p Parent, esp time.Duration) time.Time {
	if p.Committed().Before(osm.CommitInfoStart) {
		return p.Timestamp().Add(esp)
	}

	return p.Committed()
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}

	return d
}
