package osmapi

import (
	"fmt"

	"github.com/paulmach/go.osm"
	"golang.org/x/net/context"
)

// Changeset returns a given changeset from the osm rest api.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func Changeset(ctx context.Context, id osm.ChangesetID) (*osm.Changeset, error) {
	return DefaultDataSource.Changeset(ctx, id)
}

// Changeset returns a given changeset from the osm rest api.
func (ds *DataSource) Changeset(ctx context.Context, id osm.ChangesetID) (*osm.Changeset, error) {
	url := fmt.Sprintf("%s/changeset/%d", ds.baseURL(), id)
	return ds.getChangeset(ctx, url)
}

// ChangesetWithDiscussion returns a changeset and its discussion from the osm rest api.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func ChangesetWithDiscussion(ctx context.Context, id osm.ChangesetID) (*osm.Changeset, error) {
	return DefaultDataSource.ChangesetWithDiscussion(ctx, id)
}

// ChangesetWithDiscussion returns a changeset and its discussion from the osm rest api.
func (ds *DataSource) ChangesetWithDiscussion(ctx context.Context, id osm.ChangesetID) (*osm.Changeset, error) {
	url := fmt.Sprintf("%s/changeset/%d?include_discussion=true", ds.baseURL(), id)
	return ds.getChangeset(ctx, url)
}

func (ds *DataSource) getChangeset(ctx context.Context, url string) (*osm.Changeset, error) {
	css := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &css); err != nil {
		return nil, err
	}

	if l := len(css.Changesets); l != 1 {
		return nil, fmt.Errorf("wrong number of changesets, expected 1, got %v", l)
	}

	return css.Changesets[0], nil
}

// ChangesetDownload returns the full osmchange for the changeset using the osm rest api.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func ChangesetDownload(ctx context.Context, id osm.ChangesetID) (*osm.Change, error) {
	return DefaultDataSource.ChangesetDownload(ctx, id)
}

// ChangesetDownload returns the full osmchange for the changeset using the osm rest api.
func (ds *DataSource) ChangesetDownload(ctx context.Context, id osm.ChangesetID) (*osm.Change, error) {
	url := fmt.Sprintf("%s/changeset/%d/download", ds.baseURL(), id)

	change := &osm.Change{}
	if err := ds.getFromAPI(ctx, url, &change); err != nil {
		return nil, err
	}

	return change, nil
}
