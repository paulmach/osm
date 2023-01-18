package osmapi

import (
	"context"
	"fmt"

	"github.com/onXmaps/osm"
)

// User returns the user from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func User(ctx context.Context, id osm.UserID) (*osm.User, error) {
	return DefaultDatasource.User(ctx, id)
}

// User returns the user from the osm rest api.
func (ds *Datasource) User(ctx context.Context, id osm.UserID) (*osm.User, error) {
	url := fmt.Sprintf("%s/user/%d", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Users); l != 1 {
		return nil, fmt.Errorf("wrong number of users, expected 1, got %v", l)
	}

	return o.Users[0], nil
}
