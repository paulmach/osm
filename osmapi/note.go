package osmapi

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/onXmaps/osm"
)

// Note returns the note from the osm rest api.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Note(ctx context.Context, id osm.NoteID) (*osm.Note, error) {
	return DefaultDatasource.Note(ctx, id)
}

// Note returns the note from the osm rest api.
func (ds *Datasource) Note(ctx context.Context, id osm.NoteID) (*osm.Note, error) {
	url := fmt.Sprintf("%s/notes/%d", ds.baseURL(), id)

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	if l := len(o.Notes); l != 1 {
		return nil, fmt.Errorf("wrong number of notes, expected 1, got %v", l)
	}

	return o.Notes[0], nil
}

// Notes returns the notes in a bounding box. Can provide options to limit the results
// or change what it means to be "closed". See the options or osm api v0.6 docs for details.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Notes(ctx context.Context, bounds *osm.Bounds, opts ...NotesOption) (osm.Notes, error) {
	return DefaultDatasource.Notes(ctx, bounds, opts...)
}

var _ NotesOption = Limit(1)
var _ NotesOption = MaxDaysClosed(1)

// Notes returns the notes in a bounding box. Can provide options to limit the results
// or change what it means to be "closed". See the options or osm api v0.6 docs for details.
func (ds *Datasource) Notes(ctx context.Context, bounds *osm.Bounds, opts ...NotesOption) (osm.Notes, error) {
	params := make([]string, 0, 1+len(opts))
	params = append(params, fmt.Sprintf("bbox=%f,%f,%f,%f",
		bounds.MinLon, bounds.MinLat,
		bounds.MaxLon, bounds.MaxLat))

	var err error
	for _, o := range opts {
		params, err = o.applyNotes(params)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/notes?%s", ds.baseURL(), strings.Join(params, "&"))

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Notes, nil
}

// NotesSearch returns the notes in a bounding box whose text matches the query.
// Can provide options to limit the results or change what it means to be "closed".
// See the options or osm api v0.6 docs for details.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func NotesSearch(ctx context.Context, query string, opts ...NotesOption) (osm.Notes, error) {
	return DefaultDatasource.NotesSearch(ctx, query, opts...)
}

// NotesSearch returns the notes whose text matches the query.
// Can provide options to limit the results or change what it means to be "closed".
// See the options or osm api v0.6 docs for details.
func (ds *Datasource) NotesSearch(ctx context.Context, query string, opts ...NotesOption) (osm.Notes, error) {
	params := make([]string, 0, 1+len(opts))
	params = append(params, fmt.Sprintf("q=%s", url.QueryEscape(query)))

	var err error
	for _, o := range opts {
		params, err = o.applyNotes(params)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/notes/search?%s", ds.baseURL(), strings.Join(params, "&"))

	o := &osm.OSM{}
	if err := ds.getFromAPI(ctx, url, &o); err != nil {
		return nil, err
	}

	return o.Notes, nil
}
