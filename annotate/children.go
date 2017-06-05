package annotate

import (
	"time"

	osm "github.com/paulmach/go.osm"
	"github.com/paulmach/go.osm/annotate/internal/core"
)

// A childNode wraps a node into a core.Child interface
// so that it can be used to compute updates for ways and relations.
type childNode struct {
	Index int
	Node  *osm.Node
}

var _ core.Child = childNode{}

func (c childNode) ID() osm.FeatureID {
	return c.Node.FeatureID()
}

func (c childNode) ChangesetID() osm.ChangesetID {
	return c.Node.ChangesetID
}

func (c childNode) VersionIndex() int {
	return c.Index
}

func (c childNode) Visible() bool {
	return c.Node.Visible
}

func (c childNode) Timestamp() time.Time {
	return c.Node.Timestamp
}

func (c childNode) Committed() time.Time {
	if c.Node.Committed == nil {
		return time.Time{}
	}
	return *c.Node.Committed
}

func (c childNode) Update() osm.Update {
	return osm.Update{
		Version:     c.Node.Version,
		Timestamp:   updateTimestamp(c.Node.Timestamp, c.Node.Committed),
		ChangesetID: c.Node.ChangesetID,
		Lat:         c.Node.Lat,
		Lon:         c.Node.Lon,
	}
}

// A childWay wraps a way into a core.Child interface
// so that it can be used to compute updates for ways and relations.
type childWay struct {
	Index int
	Way   *osm.Way
}

var _ core.Child = childWay{}

func (c childWay) ID() osm.FeatureID {
	return c.Way.FeatureID()
}

func (c childWay) ChangesetID() osm.ChangesetID {
	return c.Way.ChangesetID
}

func (c childWay) VersionIndex() int {
	return c.Index
}

func (c childWay) Visible() bool {
	return c.Way.Visible
}

func (c childWay) Timestamp() time.Time {
	return c.Way.Timestamp
}

func (c childWay) Committed() time.Time {
	if c.Way.Committed == nil {
		return time.Time{}
	}
	return *c.Way.Committed
}

func (c childWay) Update() osm.Update {
	return osm.Update{
		Version:     c.Way.Version,
		Timestamp:   updateTimestamp(c.Way.Timestamp, c.Way.Committed),
		ChangesetID: c.Way.ChangesetID,
	}
}

// A childRelation wraps a way into a core.Child interface
// so that it can be used to compute the full relation version and its updates.
type childRelation struct {
	Index    int
	Relation *osm.Relation
}

var _ core.Child = childRelation{}

func (c childRelation) ID() osm.FeatureID {
	return c.Relation.FeatureID()
}

func (c childRelation) ChangesetID() osm.ChangesetID {
	return c.Relation.ChangesetID
}

func (c childRelation) VersionIndex() int {
	return c.Index
}

func (c childRelation) Visible() bool {
	return c.Relation.Visible
}

func (c childRelation) Timestamp() time.Time {
	return c.Relation.Timestamp
}

func (c childRelation) Committed() time.Time {
	if c.Relation.Committed == nil {
		return time.Time{}
	}
	return *c.Relation.Committed
}

func (c childRelation) Update() osm.Update {
	return osm.Update{
		Version:     c.Relation.Version,
		Timestamp:   updateTimestamp(c.Relation.Timestamp, c.Relation.Committed),
		ChangesetID: c.Relation.ChangesetID,
	}
}

func updateTimestamp(timestamp time.Time, committed *time.Time) time.Time {
	if timestamp.Before(osm.CommitInfoStart) || committed == nil {
		return timestamp
	}

	return *committed
}
