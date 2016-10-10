package osmapi

import (
	"fmt"

	"github.com/paulmach/go.osm"
	"golang.org/x/net/context"
)

// Node returns the latest version of the node from the osm rest api.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func Node(ctx context.Context, id osm.NodeID) (*osm.Node, error) {
	return DefaultDataSource.Node(ctx, id)
}

// Node returns the latest version of the node from the osm rest api.
func (ds *DataSource) Node(ctx context.Context, id osm.NodeID) (*osm.Node, error) {
	url := fmt.Sprintf("%s/node/%d", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Nodes); l != 1 {
		return nil, fmt.Errorf("wrong number of nodes, expected 1, got %v", l)
	}

	return o.Nodes[0], nil
}

// NodeVersion returns the specific version of the node from the osm rest api.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func NodeVersion(ctx context.Context, id osm.NodeID, v int) (*osm.Node, error) {
	return DefaultDataSource.NodeVersion(ctx, id, v)
}

// NodeVersion returns the specific version of the node from the osm rest api.
func (ds *DataSource) NodeVersion(ctx context.Context, id osm.NodeID, v int) (*osm.Node, error) {
	url := fmt.Sprintf("%s/node/%d/%d", ds.baseURL(), id, v)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Nodes); l != 1 {
		return nil, fmt.Errorf("wrong number of nodes, expected 1, got %v", l)
	}

	return o.Nodes[0], nil
}

// NodeHistory returns all the versions of the node from the osm rest api.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func NodeHistory(ctx context.Context, id osm.NodeID) (osm.Nodes, error) {
	return DefaultDataSource.NodeHistory(ctx, id)
}

// NodeHistory returns all the versions of the node from the osm rest api.
func (ds *DataSource) NodeHistory(ctx context.Context, id osm.NodeID) (osm.Nodes, error) {
	url := fmt.Sprintf("%s/node/%d/history", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Nodes, nil
}

// NodeWays returns all ways a node is part of.
// There is no error if the element does not exist.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func NodeWays(ctx context.Context, id osm.NodeID) (osm.Ways, error) {
	return DefaultDataSource.NodeWays(ctx, id)
}

// NodeWays returns all ways a node is part of.
// There is no error if the element does not exist.
func (ds *DataSource) NodeWays(ctx context.Context, id osm.NodeID) (osm.Ways, error) {
	url := fmt.Sprintf("%s/node/%d/ways", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Ways, nil
}

// NodeRelations returns all relations a node is part of.
// There is no error if the element does not exist.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func NodeRelations(ctx context.Context, id osm.NodeID) (osm.Relations, error) {
	return DefaultDataSource.NodeRelations(ctx, id)
}

// NodeRelations returns all relations a node is part of.
// There is no error if the element does not exist.
func (ds *DataSource) NodeRelations(ctx context.Context, id osm.NodeID) (osm.Relations, error) {
	url := fmt.Sprintf("%s/node/%d/relations", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}
