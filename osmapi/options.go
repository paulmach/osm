package osmapi

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// FeatureOption can be used when fetching a feature or a set of different features.
type FeatureOption interface {
	applyFeature([]string) ([]string, error)
}

// At adds an `at=2006-01-02T15:04:05Z` parameter to the request.
// The osm.fyi supports requesting features and maps as they were at the given time.
func At(t time.Time) FeatureOption {
	return &at{t}
}

type at struct{ t time.Time }

func (o *at) applyFeature(p []string) ([]string, error) {
	return append(p, "at="+o.t.UTC().Format("2006-01-02T15:04:05Z")), nil
}

// NotesOption defines a valid option for the osmapi.Notes by bounding box api.
type NotesOption interface {
	applyNotes([]string) ([]string, error)
}

// Limit indicates the number of results to return valid values [1,10000].
// Default is 100.
func Limit(num int) NotesOption {
	return &limit{num}
}

// MaxDaysClosed specifies the number of days a note needs to be closed to
// no longer be returned. 0 will return only open notes, -1 will return all notes.
// Default is 7.
func MaxDaysClosed(num int) NotesOption {
	return &maxDaysClosed{num}
}

type limit struct{ n int }

func (o *limit) applyNotes(p []string) ([]string, error) {
	if o.n < 1 || 10000 < o.n {
		return nil, errors.New("osmapi: limit must be between 1 and 10000")
	}
	return append(p, fmt.Sprintf("limit=%d", o.n)), nil
}

type maxDaysClosed struct{ n int }

func (o *maxDaysClosed) applyNotes(p []string) ([]string, error) {
	return append(p, fmt.Sprintf("closed=%d", o.n)), nil
}

func featureOptions(opts []FeatureOption) (string, error) {
	if len(opts) == 0 {
		return "", nil
	}

	params := make([]string, 0, len(opts))

	var err error
	for _, o := range opts {
		params, err = o.applyFeature(params)
		if err != nil {
			return "", err
		}
	}

	return strings.Join(params, "&"), nil
}
