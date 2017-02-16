package augment

import (
	"context"

	osm "github.com/paulmach/go.osm"
)

// A NodeDatasource defines where node child history data comes from.
type NodeDatasource interface {
	NodeHistory(context.Context, osm.NodeID) (osm.Nodes, error)
}

// A RelationDatasource defines where relation child history data comes from.
type RelationDatasource interface {
	RelationHistory(context.Context, osm.RelationID) (osm.Relations, error)
}

// A Datasource defines where child history data comes from.
type Datasource interface {
	NodeDatasource
	WayHistory(context.Context, osm.WayID) (osm.Ways, error)
	RelationDatasource
}

// A MapDatasource wraps maps to implement the DataSource interface.
type MapDatasource struct {
	Nodes     map[osm.NodeID]osm.Nodes
	Ways      map[osm.WayID]osm.Ways
	Relations map[osm.RelationID]osm.Relations
}

// NewDatasource createsa new MapDatasource from the arrays of elements.
// It does the conversion of array to map.
func NewDatasource(nodes osm.Nodes, ways osm.Ways, relations osm.Relations) *MapDatasource {
	mds := &MapDatasource{}

	if len(nodes) > 0 {
		mds.Nodes = make(map[osm.NodeID]osm.Nodes)
		for _, n := range nodes {
			mds.Nodes[n.ID] = append(mds.Nodes[n.ID], n)
		}
	}

	if len(ways) > 0 {
		mds.Ways = make(map[osm.WayID]osm.Ways)
		for _, w := range ways {
			mds.Ways[w.ID] = append(mds.Ways[w.ID], w)
		}
	}

	if len(relations) > 0 {
		mds.Relations = make(map[osm.RelationID]osm.Relations)
		for _, r := range relations {
			mds.Relations[r.ID] = append(mds.Relations[r.ID], r)
		}
	}

	return mds
}

// NodeHistory returns the history for the given id from the map.
func (mds MapDatasource) NodeHistory(ctx context.Context, id osm.NodeID) (osm.Nodes, error) {
	return mds.Nodes[id], nil
}

// WayHistory returns the history for the given id from the map.
func (mds MapDatasource) WayHistory(ctx context.Context, id osm.WayID) (osm.Ways, error) {
	return mds.Ways[id], nil
}

// RelationHistory returns the history for the given id from the map.
func (mds MapDatasource) RelationHistory(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	return mds.Relations[id], nil
}
