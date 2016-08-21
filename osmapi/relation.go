package osmapi

import (
	"fmt"

	"github.com/paulmach/go.osm"
	"golang.org/x/net/context"
)

// Relation returns the latest version of the relation from the osm rest api.
func Relation(ctx context.Context, id osm.RelationID) (*osm.Relation, error) {
	url := fmt.Sprintf("%s/relation/%d", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Relations); l != 1 {
		return nil, fmt.Errorf("wrong number of relations, expected 1, got %v", l)
	}

	return o.Relations[0], nil
}

// RelationVersion returns the specific version of the relation from the osm rest api.
func RelationVersion(ctx context.Context, id osm.RelationID, v int) (*osm.Relation, error) {
	url := fmt.Sprintf("%s/relation/%d/%d", host, id, v)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Relations); l != 1 {
		return nil, fmt.Errorf("wrong number of relations, expected 1, got %v", l)
	}

	return o.Relations[0], nil
}

// RelationRelations returns all relations a relation is part of.
// There is no error if the element does not exist.
func RelationRelations(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	url := fmt.Sprintf("%s/relation/%d/relations", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}

// RelationHistory returns all the versions of the relation from the osm rest api.
func RelationHistory(ctx context.Context, id osm.RelationID) (osm.Relations, error) {
	url := fmt.Sprintf("%s/relation/%d/history", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Relations, nil
}

// RelationFull returns the relation and its nodes for the latest version the relation.
func RelationFull(ctx context.Context, id osm.RelationID) (*osm.OSM, error) {
	url := fmt.Sprintf("%s/relation/%d/full", host, id)

	o := &osm.OSM{}
	if err := getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o, nil
}
