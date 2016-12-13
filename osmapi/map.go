package osmapi

import (
	"context"
	"fmt"

	osm "github.com/paulmach/go.osm"
)

// Map returns the latest elements in the given bounding box.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Map(ctx context.Context, left, bottom, right, top float64) (*osm.OSM, error) {
	return DefaultDatasource.Map(ctx, left, bottom, right, top)
}

// Map returns the latest elements in the given bounding box.
func (ds *Datasource) Map(ctx context.Context, left, bottom, right, top float64) (*osm.OSM, error) {
	url := fmt.Sprintf("%s/map?bbox=%f,%f,%f,%f", ds.baseURL(), left, bottom, right, top)
	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o, nil
}
