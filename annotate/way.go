package annotate

import (
	"context"
	"time"

	"github.com/onXmaps/osm"
	"github.com/onXmaps/osm/annotate/internal/core"
	"github.com/onXmaps/osm/annotate/shared"
)

// NodeHistoryDatasourcer is an more strict interface for when we only need node history.
type NodeHistoryDatasourcer interface {
	NodeHistory(context.Context, osm.NodeID) (osm.Nodes, error)
	NotFound(error) bool
}

// NodeHistoryAsChildrenDatasourcer is an advanced data source that
// returns the needed nodes as children directly.
type NodeHistoryAsChildrenDatasourcer interface {
	NodeHistoryDatasourcer
	NodeHistoryAsChildren(context.Context, osm.NodeID) ([]*shared.Child, error)
}

var _ NodeHistoryDatasourcer = &osm.HistoryDatasource{}

// Ways computes the updates for the given ways
// and annotate the way nodes with changeset and lon/lat data.
// The input ways are modified to include this information.
func Ways(
	ctx context.Context,
	ways osm.Ways,
	datasource NodeHistoryDatasourcer,
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

	parents := make([]core.Parent, len(ways))
	for i, w := range ways {
		parents[i] = &parentWay{Way: w}
	}

	wds := newWayDatasourcer(datasource)
	updatesForParents, err := core.Compute(ctx, parents, wds, computeOpts)
	if err != nil {
		return mapErrors(err)
	}

	// fill in updates
	for i, updates := range updatesForParents {
		ways[i].Updates = updates
	}

	return nil
}

// A parentWay wraps a osm.Way into the core.Parent interface
// so that updates can be computed.
type parentWay struct {
	Way *osm.Way
}

func (w *parentWay) ID() osm.FeatureID {
	return w.Way.FeatureID()
}

func (w *parentWay) ChangesetID() osm.ChangesetID {
	return w.Way.ChangesetID
}

func (w *parentWay) Version() int {
	return w.Way.Version
}

func (w *parentWay) Visible() bool {
	return w.Way.Visible
}

func (w *parentWay) Timestamp() time.Time {
	return w.Way.Timestamp
}

func (w *parentWay) Committed() time.Time {
	if w.Way.Committed == nil {
		return time.Time{}
	}

	return *w.Way.Committed
}

func (w *parentWay) Refs() (osm.FeatureIDs, []bool) {
	ids := make(osm.FeatureIDs, len(w.Way.Nodes))
	annotated := make([]bool, len(w.Way.Nodes))

	for i := range w.Way.Nodes {
		ids[i] = w.Way.Nodes[i].FeatureID()
		annotated[i] = w.Way.Nodes[i].Version != 0
	}

	return ids, annotated
}

func (w *parentWay) SetChild(idx int, child *shared.Child) {
	if child == nil {
		return
	}

	w.Way.Nodes[idx].Version = child.Version
	w.Way.Nodes[idx].ChangesetID = child.ChangesetID
	w.Way.Nodes[idx].Lat = child.Lat
	w.Way.Nodes[idx].Lon = child.Lon
}
