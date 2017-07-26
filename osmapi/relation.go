package osmapi

import (
	"context"
	"fmt"
	"strconv"

	"github.com/paulmach/osm"
)

// Relation returns the latest version of the relation from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Relation(ctx context.Context, id osm.RelationID) (*osm.Relation, error) {
	return DefaultDatasource.Relation(ctx, id)
}

// Relation returns the latest version of the relation from the osm rest api.
func (ds *Datasource) Relation(ctx context.Context, id osm.RelationID) (*osm.Relation, error) {
	url := fmt.Sprintf("%s/relation/%d", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Relations); l != 1 {
		return nil, fmt.Errorf("wrong number of relations, expected 1, got %v", l)
	}

	return o.Relations[0], nil
}

// Relations returns the latest version of the relations from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Relations(ctx context.Context, ids []osm.RelationID) (osm.Relations, error) {
	return DefaultDatasource.Relations(ctx, ids)
}

// Relations returns the latest version of the relations from the osm rest api.
// Will return 404 if any node is missing.
func (ds *Datasource) Relations(ctx context.Context, ids []osm.RelationID) (osm.Relations, error) {
	data := make([]byte, 0, 11*len(ids))
	for i, id := range ids {
		if i != 0 {
			data = append(data, byte(','))
		}
		data = strconv.AppendInt(data, int64(id), 10)
	}
	url := ds.baseURL() + "/relations?relations=" + string(data)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}

// RelationVersion returns the specific version of the relation from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func RelationVersion(ctx context.Context, id osm.RelationID, v int) (*osm.Relation, error) {
	return DefaultDatasource.RelationVersion(ctx, id, v)
}

// RelationVersion returns the specific version of the relation from the osm rest api.
func (ds *Datasource) RelationVersion(ctx context.Context, id osm.RelationID, v int) (*osm.Relation, error) {
	url := fmt.Sprintf("%s/relation/%d/%d", ds.baseURL(), id, v)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Relations); l != 1 {
		return nil, fmt.Errorf("wrong number of relations, expected 1, got %v", l)
	}

	return o.Relations[0], nil
}

// RelationRelations returns all relations a relation is part of.
// There is no error if the element does not exist.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func RelationRelations(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	return DefaultDatasource.RelationRelations(ctx, id)
}

// RelationRelations returns all relations a relation is part of.
// There is no error if the element does not exist.
func (ds *Datasource) RelationRelations(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	url := fmt.Sprintf("%s/relation/%d/relations", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}

// RelationHistory returns all the versions of the relation from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func RelationHistory(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	return DefaultDatasource.RelationHistory(ctx, id)
}

// RelationHistory returns all the versions of the relation from the osm rest api.
func (ds *Datasource) RelationHistory(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	url := fmt.Sprintf("%s/relation/%d/history", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}

// RelationFull returns the relation and its nodes for the latest version the relation.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func RelationFull(ctx context.Context, id osm.RelationID) (*osm.OSM, error) {
	return DefaultDatasource.RelationFull(ctx, id)
}

// RelationFull returns the relation and its nodes for the latest version the relation.
func (ds *Datasource) RelationFull(ctx context.Context, id osm.RelationID) (*osm.OSM, error) {
	url := fmt.Sprintf("%s/relation/%d/full", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o, nil
}
