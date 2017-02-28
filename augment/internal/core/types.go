package core

import (
	"time"

	"github.com/paulmach/go.osm"
)

// A Parent is something that holds children. ie. ways have nodes as children
// and relations can have nodes, ways and relations as children.
type Parent interface {
	ID() osm.ElementID // used for logging
	ChangesetID() osm.ChangesetID

	Version() int
	Visible() bool
	Timestamp() time.Time
	Committed() time.Time

	Refs() osm.ElementIDs
	Children() ChildList
	SetChildren(ChildList)
}

// A Child a thing contained by parents such as nodes for ways or nodes, ways
// and/or relations for relations.
type Child interface {
	ID() osm.ElementID
	ChangesetID() osm.ChangesetID

	// VersionIndex is the index of the version if sorted from lowest to highest.
	// This is necessary since version don't have to start at 1 or be sequential.
	VersionIndex() int
	Visible() bool
	Timestamp() time.Time
	Committed() time.Time
	Update() osm.Update
}

// A ChildList is a set
type ChildList []Child

// FindVisible locates the child visible at the given time.
// If 'at' is on or after osm.CommitInfoStart the committed
// time is used to determine visiblity. If 'at' is before, a range +-eps
// around the give time. Will return the closes visible node within that
// range, or the previous node if visible. Children after 'at' but within
// the eps must have the same changeset id as provided (the parent's).
// If the previous node is not visible, or does not exits will return nil.
func (cl ChildList) FindVisible(cid osm.ChangesetID, at time.Time, eps time.Duration) Child {
	var (
		diff    time.Duration = -1
		nearest Child
	)

	start := at.Add(-eps)
	for _, c := range cl {
		if c.Committed().Before(osm.CommitInfoStart) {
			// more complicated logic for early data.
			offset := c.Timestamp().Sub(start)
			visible := c.Visible()

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
					} else if c.ChangesetID() == cid {
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
			if c.Committed().After(at) {
				break
			}

			if c.Visible() {
				nearest = c
			} else {
				nearest = nil
			}
		}
	}

	return nearest
}

// LastVisibleBefore finds the last visible child before a given time.
func (cl ChildList) LastVisibleBefore(end time.Time) Child {
	var latest Child

	for _, c := range cl {
		if !timeThreshold(c, 0).Before(end) {
			break
		}

		if c.Visible() {
			latest = c
		}
	}

	return latest
}

type timer interface {
	Timestamp() time.Time
	Committed() time.Time
}

func timeThreshold(c timer, esp time.Duration) time.Time {
	if c.Committed().Before(osm.CommitInfoStart) {
		return c.Timestamp().Add(esp)
	}

	return c.Committed()
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}

	return d
}
