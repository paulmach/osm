package core

import (
	"context"
	"errors"

	"github.com/onXmaps/osm"
)

var _ Datasourcer = &TestDS{}

var ErrNotFound = errors.New("not found")

// TestDS implements a datasource for testing.
type TestDS struct {
	data map[osm.FeatureID]ChildList
}

// Get returns the history in ChildList form.
func (tds *TestDS) Get(ctx context.Context, id osm.FeatureID) (ChildList, error) {
	if tds.data == nil {
		return nil, ErrNotFound
	}

	v := tds.data[id]
	if v == nil {
		return nil, ErrNotFound
	}

	return v, nil
}

// MustGet is used by tests only to simplify some code.
func (tds *TestDS) MustGet(id osm.FeatureID) ChildList {
	v, err := tds.Get(context.TODO(), id)
	if err != nil {
		panic(err)
	}

	return v
}

func (tds *TestDS) NotFound(err error) bool {
	return err == ErrNotFound
}

// Set sets the element history into the map.
// The element is deleted if list is nil.
func (tds *TestDS) Set(id osm.FeatureID, list ChildList) {
	if tds.data == nil {
		tds.data = make(map[osm.FeatureID]ChildList)
	}

	if list == nil {
		delete(tds.data, id)
	}

	tds.data[id] = list
}
