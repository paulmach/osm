package annotate

import (
	"context"
	"time"

	"github.com/onXmaps/osm"
	"github.com/onXmaps/osm/annotate/internal/core"
	"github.com/onXmaps/osm/annotate/shared"
)

// HistoryAsChildrenDatasourcer is an advanced data source that
// returns the needed elements as children directly.
type HistoryAsChildrenDatasourcer interface {
	osm.HistoryDatasourcer
	NodeHistoryAsChildren(context.Context, osm.NodeID) ([]*shared.Child, error)
	WayHistoryAsChildren(context.Context, osm.WayID) ([]*shared.Child, error)
	RelationHistoryAsChildren(context.Context, osm.RelationID) ([]*shared.Child, error)
}

// Relations computes the updates for the given relations
// and annotate members with stuff like changeset and lon/lat data.
// The input relations are modified to include this information.
func Relations(
	ctx context.Context,
	relations osm.Relations,
	datasource osm.HistoryDatasourcer,
	opts ...Option,
) error {
	computeOpts := &core.Options{
		Threshold: defaultThreshold,
	}
	for _, o := range opts {
		err := o(computeOpts)
		if err != nil {
			return err
		}
	}

	parents := make([]core.Parent, len(relations))
	for i, r := range relations {
		parents[i] = &parentRelation{Relation: r}
	}

	rds := newRelationDatasourcer(datasource)
	updatesForParents, err := core.Compute(ctx, parents, rds, computeOpts)
	if err != nil {
		return mapErrors(err)
	}

	for _, p := range parents {
		r := p.(*parentRelation)
		if r.Relation.Polygon() {
			orientation(r.Relation.Members, r.ways, r.Relation.CommittedAt())
		}
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
	ways     map[osm.WayID]*osm.Way
}

func (r *parentRelation) ID() osm.FeatureID {
	return r.Relation.FeatureID()
}

func (r *parentRelation) ChangesetID() osm.ChangesetID {
	return r.Relation.ChangesetID
}

func (r *parentRelation) Version() int {
	return r.Relation.Version
}

func (r *parentRelation) Visible() bool {
	return r.Relation.Visible
}

func (r *parentRelation) Timestamp() time.Time {
	return r.Relation.Timestamp
}

func (r *parentRelation) Committed() time.Time {
	if r.Relation.Committed == nil {
		return time.Time{}
	}

	return *r.Relation.Committed
}

func (r *parentRelation) Refs() (osm.FeatureIDs, []bool) {
	ids := make(osm.FeatureIDs, len(r.Relation.Members))
	annotated := make([]bool, len(r.Relation.Members))

	for i := range r.Relation.Members {
		ids[i] = r.Relation.Members[i].FeatureID()
		annotated[i] = r.Relation.Members[i].Version != 0
	}

	return ids, annotated
}

func (r *parentRelation) SetChild(idx int, child *shared.Child) {
	if r.Relation.Polygon() && r.ways == nil {
		r.ways = make(map[osm.WayID]*osm.Way, len(r.Relation.Members))
	}

	if child == nil {
		return
	}

	r.Relation.Members[idx].Version = child.Version
	r.Relation.Members[idx].ChangesetID = child.ChangesetID
	r.Relation.Members[idx].Lat = child.Lat
	r.Relation.Members[idx].Lon = child.Lon

	if r.ways != nil && child.Way != nil {
		r.ways[child.Way.ID] = child.Way
	}
}
