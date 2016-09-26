package core

import (
	"fmt"
	"time"

	"github.com/paulmach/go.osm"
)

// ChildType indicates a specific type of child, ie. node, way, relation.
// We use integers here to they are sortable.
type ChildType int

// The enumerated list of Child types.
const (
	NodeType ChildType = iota + 1
	WayType
	RelationType
)

func (ct ChildType) String() string {
	switch ct {
	case NodeType:
		return "node"
	case WayType:
		return "way"
	case RelationType:
		return "relation"
	}

	return fmt.Sprintf("unknown type %d", ct)
}

// TypeMapToOSM is here to help of conversion fo the different types.
var TypeMapToOSM = map[ChildType]osm.ElementType{
	NodeType:     osm.NodeType,
	WayType:      osm.WayType,
	RelationType: osm.RelationType,
}

// TypeMapToCore is here to help of conversion fo the different types.
var TypeMapToCore = map[osm.ElementType]ChildType{
	osm.NodeType:     NodeType,
	osm.WayType:      WayType,
	osm.RelationType: RelationType,
}

// ChildID represent a specific child by both its type and id.
type ChildID struct {
	Type ChildType
	ID   int64
}

func (cid ChildID) String() string {
	return fmt.Sprintf("%v %v", cid.Type, cid.ID)
}

// A Parent is something that holds children. ie. ways have nodes as children
// and relations can have nodes, ways and relations as children.
type Parent interface {
	// used for logging
	ID() (osm.ElementType, int64)

	Version() int
	Visible() bool
	Timestamp() time.Time
	Committed() time.Time

	Refs() []ChildID
	Children() ChildList
	SetChildren(ChildList)
}

// A Child a thing contained by parents such as nodes for ways or nodes, ways
// and/or relations for relations.
type Child interface {
	ID() ChildID

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

// FindVisible find the child visible at the given time.
// If 'at' is on or after osm.CommitInfoStart the committed
// time is used to determine visiblity. If 'at' is before, a range +-eps
// around the give time. Will return the closes visible node within that
// range, or the previous node if visible. If the previous node is not
// visible, or does not exits will return nil.
func (cl ChildList) FindVisible(at time.Time, eps time.Duration) Child {
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
					nearest = c
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
