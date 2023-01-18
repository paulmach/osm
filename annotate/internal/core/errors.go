package core

import (
	"fmt"
	"time"

	"github.com/onXmaps/osm"
)

// NoHistoryError is returned if there is no entry in the history
// map for a specific child.
type NoHistoryError struct {
	ChildID osm.FeatureID
}

// Error returns a pretty string of the error.
func (e *NoHistoryError) Error() string {
	return fmt.Sprintf("element history not found for %v", e.ChildID)
}

// NoVisibleChildError is returned if there are no visible children
// for a parent at a given time.
type NoVisibleChildError struct {
	ChildID   osm.FeatureID
	Timestamp time.Time
}

// Error returns a pretty string of the error.
func (e *NoVisibleChildError) Error() string {
	return fmt.Sprintf("no visible child for %v at %v", e.ChildID, e.Timestamp)
}
