package osmapi

import (
	"fmt"

	"github.com/paulmach/go.osm"
	"golang.org/x/net/context"
)

// Changeset returns a given changeset from the osm rest api.
func Changeset(ctx context.Context, id osm.ChangesetID) (*osm.Changeset, error) {
	return getChangeset(ctx, id, false)
}

// ChangesetWithDiscussion returns a changeset and its discussion from the osm rest api.
func ChangesetWithDiscussion(ctx context.Context, id osm.ChangesetID) (*osm.Changeset, error) {
	return getChangeset(ctx, id, true)
}

func getChangeset(ctx context.Context, id osm.ChangesetID, disc bool) (*osm.Changeset, error) {
	var url string
	if disc {
		url = fmt.Sprintf("%s/changeset/%d?include_discussion=true", host, id)
	} else {
		url = fmt.Sprintf("%s/changeset/%d", host, id)
	}

	css := &osm.OSMChangesets{}
	if err := getFromAPI(ctx, url, &css); err != nil {
		return nil, err
	}

	if l := len(css.Changesets); l != 1 {
		return nil, fmt.Errorf("wrong number of changesets, expected 1, got %v", l)
	}

	return css.Changesets[0], nil
}

// ChangesetDownload returns the full osmchange for the changeset using the osm rest api.
func ChangesetDownload(ctx context.Context, id osm.ChangesetID) (*osm.Change, error) {
	url := fmt.Sprintf("%s/changeset/%d/download", host, id)

	change := &osm.Change{}
	if err := getFromAPI(ctx, url, &change); err != nil {
		return nil, err
	}

	return change, nil
}
