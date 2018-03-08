package osm

import (
	"context"
	"errors"
)

// A HistoryDatasourcer defines an interface to osm history data.
type HistoryDatasourcer interface {
	NodeHistory(context.Context, NodeID) (Nodes, error)
	WayHistory(context.Context, WayID) (Ways, error)
	RelationHistory(context.Context, RelationID) (Relations, error)
	NotFound(error) bool
}

var errNotFound = errors.New("osm: feature not found")

// A HistoryDatasource wraps maps to implement the HistoryDataSource interface.
type HistoryDatasource struct {
	Nodes     map[NodeID]Nodes
	Ways      map[WayID]Ways
	Relations map[RelationID]Relations
}

var _ HistoryDatasourcer = &HistoryDatasource{}

func (ds *HistoryDatasource) add(o *OSM, visible ...bool) {
	if o == nil {
		return
	}

	if len(o.Nodes) > 0 {
		if ds.Nodes == nil {
			ds.Nodes = make(map[NodeID]Nodes)
		}

		for _, n := range o.Nodes {
			if len(visible) == 1 {
				n.Visible = visible[0]
			}
			ds.Nodes[n.ID] = append(ds.Nodes[n.ID], n)
		}
	}

	if len(o.Ways) > 0 {
		if ds.Ways == nil {
			ds.Ways = make(map[WayID]Ways)
		}

		for _, w := range o.Ways {
			if len(visible) == 1 {
				w.Visible = visible[0]
			}
			ds.Ways[w.ID] = append(ds.Ways[w.ID], w)
		}
	}

	if len(o.Relations) > 0 {
		if ds.Relations == nil {
			ds.Relations = make(map[RelationID]Relations)
		}

		for _, r := range o.Relations {
			if len(visible) == 1 {
				r.Visible = visible[0]
			}
			ds.Relations[r.ID] = append(ds.Relations[r.ID], r)
		}
	}
}

// NodeHistory returns the history for the given id from the map.
func (ds *HistoryDatasource) NodeHistory(ctx context.Context, id NodeID) (Nodes, error) {
	if ds.Nodes == nil {
		return nil, errNotFound
	}

	v := ds.Nodes[id]
	if v == nil {
		return nil, errNotFound
	}

	return v, nil
}

// WayHistory returns the history for the given id from the map.
func (ds *HistoryDatasource) WayHistory(ctx context.Context, id WayID) (Ways, error) {
	if ds.Ways == nil {
		return nil, errNotFound
	}

	v := ds.Ways[id]
	if v == nil {
		return nil, errNotFound
	}

	return v, nil
}

// RelationHistory returns the history for the given id from the map.
func (ds *HistoryDatasource) RelationHistory(ctx context.Context, id RelationID) (Relations, error) {
	if ds.Relations == nil {
		return nil, errNotFound
	}

	v := ds.Relations[id]
	if v == nil {
		return nil, errNotFound
	}

	return v, nil
}

// NotFound returns true if the error returned is a not found error.
func (ds *HistoryDatasource) NotFound(err error) bool {
	return err == errNotFound
}
