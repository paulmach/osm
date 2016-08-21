package osmapi

import (
	"fmt"

	"github.com/paulmach/go.osm"
	"golang.org/x/net/context"
)

// Node returns the latest version of the node from the osm rest api.
func Node(ctx context.Context, id osm.NodeID) (*osm.Node, error) {
	url := fmt.Sprintf("%s/node/%d", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Nodes); l != 1 {
		return nil, fmt.Errorf("wrong number of nodes, expected 1, got %v", l)
	}

	return o.Nodes[0], nil
}

// NodeVersion returns the specific version of the node from the osm rest api.
func NodeVersion(ctx context.Context, id osm.NodeID, v int) (*osm.Node, error) {
	url := fmt.Sprintf("%s/node/%d/%d", host, id, v)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Nodes); l != 1 {
		return nil, fmt.Errorf("wrong number of nodes, expected 1, got %v", l)
	}

	return o.Nodes[0], nil
}

// NodeHistory returns all the versions of the node from the osm rest api.
func NodeHistory(ctx context.Context, id osm.NodeID) (osm.Nodes, error) {
	url := fmt.Sprintf("%s/node/%d/history", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Nodes, nil
}

// NodeWays returns all ways a node is part of.
// There is no error if the element does not exist.
func NodeWays(ctx context.Context, id osm.NodeID) (osm.Ways, error) {
	url := fmt.Sprintf("%s/node/%d/ways", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Ways, nil
}

// NodeRelations returns all relations a node is part of.
// There is no error if the element does not exist.
func NodeRelations(ctx context.Context, id osm.NodeID) (osm.Relations, error) {
	url := fmt.Sprintf("%s/node/%d/relations", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}
