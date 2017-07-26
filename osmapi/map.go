package osmapi

import (
	"context"
	"fmt"

	"github.com/paulmach/osm"
)

// Map returns the latest elements in the given bounding box.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Map(ctx context.Context, bounds *osm.Bounds) (*osm.OSM, error) {
	return DefaultDatasource.Map(ctx, bounds)
}

// Map returns the latest elements in the given bounding box.
func (ds *Datasource) Map(ctx context.Context, bounds *osm.Bounds) (*osm.OSM, error) {
	url := fmt.Sprintf("%s/map?bbox=%f,%f,%f,%f", ds.baseURL(),
		bounds.MinLon, bounds.MinLat,
		bounds.MaxLon, bounds.MaxLat)
	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o, nil
}
