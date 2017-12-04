package annotate

import (
	"context"
	"fmt"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/annotate/internal/core"
)

// Relations computes the updates for the given relations
// and annotate members with stuff like changeset and lon/lat data.
// The input relations are modified to include this information.
func Relations(
	ctx context.Context,
	relations osm.Relations,
	datasource osm.HistoryDatasourcer,
	threshold time.Duration,
	opts ...Option,
) error {
	computeOpts := &core.Options{}
	for _, o := range opts {
		err := o(computeOpts)
		if err != nil {
			return err
		}
	}
	computeOpts.Threshold = threshold

	parents := make([]core.Parent, len(relations))
	for i, r := range relations {
		parents[i] = &parentRelation{Relation: r}
	}

	rds := &relationDatasource{datasource}
	updatesForParents, err := core.Compute(ctx, parents, rds, computeOpts)
	if err != nil {
		return mapErrors(err)
	}

	for i, updates := range updatesForParents {
		relations[i].Updates = updates
	}

	return nil
}

// A parentRelation wraps a osm.Relation into the core.Parent interface
// so that updates can be computed.
type parentRelation struct {
	Relation *osm.Relation
	children core.ChildList
	refs     osm.FeatureIDs
}

func (r parentRelation) ID() osm.FeatureID {
	return r.Relation.FeatureID()
}

func (r parentRelation) ChangesetID() osm.ChangesetID {
	return r.Relation.ChangesetID
}

func (r parentRelation) Version() int {
	return r.Relation.Version
}

func (r parentRelation) Visible() bool {
	return r.Relation.Visible
}

func (r parentRelation) Timestamp() time.Time {
	return r.Relation.Timestamp
}

func (r parentRelation) Committed() time.Time {
	if r.Relation.Committed == nil {
		return time.Time{}
	}

	return *r.Relation.Committed
}

func (r parentRelation) Refs() osm.FeatureIDs {
	if r.refs == nil {
		r.refs = r.Relation.Members.FeatureIDs()
	}

	return r.refs
}

func (r parentRelation) Children() core.ChildList {
	return r.children
}

func (r *parentRelation) SetChildren(list core.ChildList) {
	r.children = list

	var ways map[osm.WayID]*osm.Way
	if r.Relation.Polygon() {
		ways = make(map[osm.WayID]*osm.Way, len(r.Relation.Members))
	}

	for i, child := range list {
		if child == nil {
			continue
		}

		switch t := child.(type) {
		case *childNode:
			r.Relation.Members[i].Version = t.Node.Version
			r.Relation.Members[i].ChangesetID = t.Node.ChangesetID
			r.Relation.Members[i].Lat = t.Node.Lat
			r.Relation.Members[i].Lon = t.Node.Lon
		case *childWay:
			r.Relation.Members[i].Version = t.Way.Version
			r.Relation.Members[i].ChangesetID = t.Way.ChangesetID

			if ways != nil {
				ways[t.Way.ID] = t.Way
			}
		case *childRelation:
			r.Relation.Members[i].Version = t.Relation.Version
			r.Relation.Members[i].ChangesetID = t.Relation.ChangesetID
		default:
			panic(fmt.Sprintf("unsupported type %T", child))
		}
	}

	if r.Relation.Polygon() {
		orientation(r.Relation.Members, ways, r.Relation.CommittedAt())
	}
}
