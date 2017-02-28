package augment

import (
	"context"
	"time"

	osm "github.com/paulmach/go.osm"
	"github.com/paulmach/go.osm/augment/internal/core"
)

// Ways computes the updates for the given ways
// and augments the way nodes with changeset and lon/lat data.
// The input ways are modified to include this information.
func Ways(
	ctx context.Context,
	ways osm.Ways,
	datasource NodeDatasource,
	threshold time.Duration,
) error {
	parents, histories, err := convertWayData(ctx, ways, datasource)
	if err != nil {
		return mapErrors(err)
	}

	updatesForParents, err := core.Compute(parents, histories, threshold)
	if err != nil {
		return mapErrors(err)
	}

	// fill in updates
	for i, updates := range updatesForParents {
		ways[i].Updates = updates
	}

	return nil
}

func convertWayData(
	ctx context.Context,
	ways osm.Ways,
	datasource NodeDatasource,
) ([]core.Parent, *core.Histories, error) {

	ways.SortByIDVersion()

	parents := make([]core.Parent, len(ways))
	histories := &core.Histories{}

	for i, w := range ways {
		parents[i] = &parentWay{Way: w}

		for _, n := range w.Nodes {
			childID := osm.ElementID{Type: osm.NodeType, Ref: int64(n.ID)}
			if histories.Get(childID) != nil {
				continue
			}

			nodes, err := datasource.NodeHistory(ctx, n.ID)
			if err != nil {
				return nil, nil, err
			}

			histories.Set(childID, nodesToChildList(nodes))
		}
	}

	return parents, histories, nil
}

func nodesToChildList(nodes osm.Nodes) core.ChildList {
	list := make(core.ChildList, len(nodes))

	nodes.SortByIDVersion()
	for i, n := range nodes {
		list[i] = &childNode{
			Index: i,
			Node:  n,
		}
	}

	return list
}

// A parentWay wraps a osm.Way into the core.Parent interface
// so that updates can be computed.
type parentWay struct {
	Way      *osm.Way
	children core.ChildList
}

func (w parentWay) ID() osm.ElementID {
	return w.Way.ElementID()
}

func (w parentWay) ChangesetID() osm.ChangesetID {
	return w.Way.ChangesetID
}

func (w parentWay) Version() int {
	return w.Way.Version
}

func (w parentWay) Visible() bool {
	return w.Way.Visible
}

func (w parentWay) Timestamp() time.Time {
	return w.Way.Timestamp
}

func (w parentWay) Committed() time.Time {
	if w.Way.Committed == nil {
		return time.Time{}
	}

	return *w.Way.Committed
}

func (w parentWay) Refs() osm.ElementIDs {
	return w.Way.Nodes.ElementIDs()
}

func (w parentWay) Children() core.ChildList {
	return w.children
}

func (w *parentWay) SetChildren(list core.ChildList) {
	w.children = list

	// copy back in the node information
	for i, child := range list {
		n := child.(*childNode).Node

		w.Way.Nodes[i].Version = n.Version
		w.Way.Nodes[i].ChangesetID = n.ChangesetID
		w.Way.Nodes[i].Lat = n.Lat
		w.Way.Nodes[i].Lon = n.Lon
	}
}
