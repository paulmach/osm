package osmapi

import (
	"fmt"

	"github.com/paulmach/go.osm"
	"golang.org/x/net/context"
)

// Way returns the latest version of the way from the osm rest api.
func Way(ctx context.Context, id osm.WayID) (*osm.Way, error) {
	url := fmt.Sprintf("%s/way/%d", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Ways); l != 1 {
		return nil, fmt.Errorf("wrong number of ways, expected 1, got %v", l)
	}

	return o.Ways[0], nil
}

// WayVersion returns the specific version of the way from the osm rest api.
func WayVersion(ctx context.Context, id osm.WayID, v int) (*osm.Way, error) {
	url := fmt.Sprintf("%s/way/%d/%d", host, id, v)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Ways); l != 1 {
		return nil, fmt.Errorf("wrong number of ways, expected 1, got %v", l)
	}

	return o.Ways[0], nil
}

// WayHistory returns all the versions of the way from the osm rest api.
func WayHistory(ctx context.Context, id osm.WayID) (osm.Ways, error) {
	url := fmt.Sprintf("%s/way/%d/history", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Ways, nil
}

// WayRelations returns all relations a way is part of.
// There is no error if the element does not exist.
func WayRelations(ctx context.Context, id osm.WayID) (osm.Relations, error) {
	url := fmt.Sprintf("%s/way/%d/relations", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}

// WayFull returns the way and its nodes for the latest version the way.
func WayFull(ctx context.Context, id osm.WayID) (*osm.OSM, error) {
	url := fmt.Sprintf("%s/way/%d/full", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o, nil
}
