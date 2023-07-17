package annotate

import (
	"fmt"
	"time"

	"github.com/onXmaps/osm"
	"github.com/onXmaps/osm/annotate/internal/core"
)

// NoHistoryError is returned if there is no entry in the history
// map for a specific child.
type NoHistoryError struct {
	ID osm.FeatureID
}

// Error returns a pretty string of the error.
func (e *NoHistoryError) Error() string {
	return fmt.Sprintf("element history not found for %v", e.ID)
}

// NoVisibleChildError is returned if there are no visible children
// for a parent at a given time.
type NoVisibleChildError struct {
	ID        osm.FeatureID
	Timestamp time.Time
}

// Error returns a pretty string of the error.
func (e *NoVisibleChildError) Error() string {
	return fmt.Sprintf("no visible child for %v at %v", e.ID, e.Timestamp)
}

// UnsupportedMemberTypeError is returned if a relation member is not a
// node, way or relation.
type UnsupportedMemberTypeError struct {
	RelationID osm.RelationID
	MemberType osm.Type
	Index      int
}

// Error returns a pretty string of the error.
func (e *UnsupportedMemberTypeError) Error() string {
	return fmt.Sprintf("unsupported member type %v for relation %d at %d", e.MemberType, e.RelationID, e.Index)
}

func mapErrors(err error) error {
	switch t := err.(type) {
	case *core.NoHistoryError:
		return &NoHistoryError{
			ID: t.ChildID,
		}
	case *core.NoVisibleChildError:
		return &NoVisibleChildError{
			ID:        t.ChildID,
			Timestamp: t.Timestamp,
		}
	}

	return err
}
