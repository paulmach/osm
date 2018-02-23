package osmapi

import (
	"errors"
	"fmt"
)

// NotesOption defines a valid option for the osm.Notes by bounding box api.
type NotesOption interface {
	apply([]string) ([]string, error)
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

func (o *limit) apply(p []string) ([]string, error) {
	if o.n < 1 || 10000 < o.n {
		return nil, errors.New("osmapi: limit must be between 1 and 10000")
	}
	return append(p, fmt.Sprintf("limit=%d", o.n)), nil
}

type maxDaysClosed struct{ n int }

func (o *maxDaysClosed) apply(p []string) ([]string, error) {
	return append(p, fmt.Sprintf("closed=%d", o.n)), nil
}
