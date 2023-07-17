package osmapi

import (
	"context"
	"fmt"

	"github.com/onXmaps/osm"
)

// Map returns the latest elements in the given bounding box.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Map(ctx context.Context, bounds *osm.Bounds, opts ...FeatureOption) (*osm.OSM, error) {
	return DefaultDatasource.Map(ctx, bounds, opts...)
}

// Map returns the latest elements in the given bounding box.
func (ds *Datasource) Map(ctx context.Context, bounds *osm.Bounds, opts ...FeatureOption) (*osm.OSM, error) {
	params, err := featureOptions(opts)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/map?bbox=%f,%f,%f,%f&%s", ds.baseURL(),
		bounds.MinLon, bounds.MinLat,
		bounds.MaxLon, bounds.MaxLat,
		params)
	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o, nil
}
