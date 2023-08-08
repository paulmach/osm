// Package shared is used by annotate and the internal core.
// External usage of this package is for advanced use only.
package shared

import (
	"time"

	"github.com/paulmach/osm"
)

// A Child represents a node, way or relation that is a dependent for
// annotating ways or relations.
type Child struct {
	ID          osm.FeatureID
	Version     int
	ChangesetID osm.ChangesetID

	// VersionIndex is the index of the version if sorted from lowest to highest.
	// This is necessary since version don't have to start at 1 or be sequential.
	VersionIndex int
	Timestamp    time.Time
	Committed    time.Time

	// for nodes
	Lon, Lat float64

	// for ways
	Way               *osm.Way
	ReverseOfPrevious bool

	// moving the visible bool here decreases the struct size from
	// size 120 (size class 128) to 112 (size class 112).
	Visible bool
}

// Update generates an update from this child.
func (c *Child) Update() osm.Update {
	return osm.Update{
		Version:     c.Version,
		Timestamp:   updateTimestamp(c.Timestamp, c.Committed),
		ChangesetID: c.ChangesetID,

		Lat: c.Lat,
		Lon: c.Lon,

		Reverse: c.ReverseOfPrevious,
	}
}

// FromNode converts a node to a child.
func FromNode(n *osm.Node) *Child {
	c := &Child{
		ID:          n.FeatureID(),
		Version:     n.Version,
		ChangesetID: n.ChangesetID,
		Visible:     n.Visible,
		Timestamp:   n.Timestamp,

		Lon: n.Lon,
		Lat: n.Lat,
	}

	if n.Committed != nil {
		c.Committed = *n.Committed
	}

	return c
}

// FromWay converts a way to a child.
func FromWay(w *osm.Way) *Child {
	c := &Child{
		ID:          w.FeatureID(),
		Version:     w.Version,
		ChangesetID: w.ChangesetID,
		Visible:     w.Visible,
		Timestamp:   w.Timestamp,
		Way:         w,
	}

	if w.Committed != nil {
		c.Committed = *w.Committed
	}

	return c
}

// FromRelation converts a way to a child.
func FromRelation(r *osm.Relation) *Child {
	c := &Child{
		ID:          r.FeatureID(),
		Version:     r.Version,
		ChangesetID: r.ChangesetID,
		Visible:     r.Visible,
		Timestamp:   r.Timestamp,
	}

	if r.Committed != nil {
		c.Committed = *r.Committed
	}

	return c
}

func updateTimestamp(timestamp, committed time.Time) time.Time {
	if timestamp.Before(osm.CommitInfoStart) || committed.IsZero() {
		return timestamp
	}

	return committed
}
