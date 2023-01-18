package osmapi

import (
	"context"
	"fmt"
	"strconv"

	"github.com/onXmaps/osm"
)

// Way returns the latest version of the way from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Way(ctx context.Context, id osm.WayID, opts ...FeatureOption) (*osm.Way, error) {
	return DefaultDatasource.Way(ctx, id, opts...)
}

// Way returns the latest version of the way from the osm rest api.
func (ds *Datasource) Way(ctx context.Context, id osm.WayID, opts ...FeatureOption) (*osm.Way, error) {
	params, err := featureOptions(opts)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/way/%d?%s", ds.baseURL(), id, params)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Ways); l != 1 {
		return nil, fmt.Errorf("wrong number of ways, expected 1, got %v", l)
	}

	return o.Ways[0], nil
}

// Ways returns the latest version of the ways from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Ways(ctx context.Context, ids []osm.WayID, opts ...FeatureOption) (osm.Ways, error) {
	return DefaultDatasource.Ways(ctx, ids, opts...)
}

// Ways returns the latest version of the ways from the osm rest api.
// Will return 404 if any way is missing.
func (ds *Datasource) Ways(ctx context.Context, ids []osm.WayID, opts ...FeatureOption) (osm.Ways, error) {
	params, err := featureOptions(opts)
	if err != nil {
		return nil, err
	}
	data := make([]byte, 0, 11*len(ids))
	for i, id := range ids {
		if i != 0 {
			data = append(data, byte(','))
		}
		data = strconv.AppendInt(data, int64(id), 10)
	}
	url := ds.baseURL() + "/ways?ways=" + string(data)
	if len(params) > 0 {
		url += "&" + params
	}

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Ways, nil
}

// WayVersion returns the specific version of the way from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func WayVersion(ctx context.Context, id osm.WayID, v int) (*osm.Way, error) {
	return DefaultDatasource.WayVersion(ctx, id, v)
}

// WayVersion returns the specific version of the way from the osm rest api.
func (ds *Datasource) WayVersion(ctx context.Context, id osm.WayID, v int) (*osm.Way, error) {
	url := fmt.Sprintf("%s/way/%d/%d", ds.baseURL(), id, v)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Ways); l != 1 {
		return nil, fmt.Errorf("wrong number of ways, expected 1, got %v", l)
	}

	return o.Ways[0], nil
}

// WayHistory returns all the versions of the way from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func WayHistory(ctx context.Context, id osm.WayID) (osm.Ways, error) {
	return DefaultDatasource.WayHistory(ctx, id)
}

// WayHistory returns all the versions of the way from the osm rest api.
func (ds *Datasource) WayHistory(ctx context.Context, id osm.WayID) (osm.Ways, error) {
	url := fmt.Sprintf("%s/way/%d/history", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Ways, nil
}

// WayRelations returns all relations a way is part of.
// There is no error if the element does not exist.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func WayRelations(ctx context.Context, id osm.WayID, opts ...FeatureOption) (osm.Relations, error) {
	return DefaultDatasource.WayRelations(ctx, id, opts...)
}

// WayRelations returns all relations a way is part of.
// There is no error if the element does not exist.
func (ds *Datasource) WayRelations(ctx context.Context, id osm.WayID, opts ...FeatureOption) (osm.Relations, error) {
	params, err := featureOptions(opts)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/way/%d/relations?%s", ds.baseURL(), id, params)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}

// WayFull returns the way and its nodes for the latest version the way.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func WayFull(ctx context.Context, id osm.WayID, opts ...FeatureOption) (*osm.OSM, error) {
	return DefaultDatasource.WayFull(ctx, id, opts...)
}

// WayFull returns the way and its nodes for the latest version the way.
func (ds *Datasource) WayFull(ctx context.Context, id osm.WayID, opts ...FeatureOption) (*osm.OSM, error) {
	params, err := featureOptions(opts)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/way/%d/full?%s", ds.baseURL(), id, params)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o, nil
}
