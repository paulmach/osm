package osmapi

import (
	"fmt"

	"github.com/paulmach/go.osm"
	"golang.org/x/net/context"
)

// Relation returns the latest version of the relation from the osm rest api.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func Relation(ctx context.Context, id osm.RelationID) (*osm.Relation, error) {
	return DefaultDataSource.Relation(ctx, id)
}

// Relation returns the latest version of the relation from the osm rest api.
func (ds *DataSource) Relation(ctx context.Context, id osm.RelationID) (*osm.Relation, error) {
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

// RelationVersion returns the specific version of the relation from the osm rest api.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func RelationVersion(ctx context.Context, id osm.RelationID, v int) (*osm.Relation, error) {
	return DefaultDataSource.RelationVersion(ctx, id, v)
}

// RelationVersion returns the specific version of the relation from the osm rest api.
func (ds *DataSource) RelationVersion(ctx context.Context, id osm.RelationID, v int) (*osm.Relation, error) {
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
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func RelationRelations(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	return DefaultDataSource.RelationRelations(ctx, id)
}

// RelationRelations returns all relations a relation is part of.
// There is no error if the element does not exist.
func (ds *DataSource) RelationRelations(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	url := fmt.Sprintf("%s/relation/%d/relations", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}

// RelationHistory returns all the versions of the relation from the osm rest api.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func RelationHistory(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	return DefaultDataSource.RelationHistory(ctx, id)
}

// RelationHistory returns all the versions of the relation from the osm rest api.
func (ds *DataSource) RelationHistory(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	url := fmt.Sprintf("%s/relation/%d/history", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}

// RelationFull returns the relation and its nodes for the latest version the relation.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func RelationFull(ctx context.Context, id osm.RelationID) (*osm.OSM, error) {
	return DefaultDataSource.RelationFull(ctx, id)
}

// RelationFull returns the relation and its nodes for the latest version the relation.
func (ds *DataSource) RelationFull(ctx context.Context, id osm.RelationID) (*osm.OSM, error) {
	url := fmt.Sprintf("%s/relation/%d/full", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o, nil
}
