package osm

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var (
	// ErrScannerClosed is returned by scanner.Err() if the scanner is closed
	// and there are no other io or xml errors to report.
	ErrScannerClosed = errors.New("osm: scanner closed by user")
)

// ElementID is a unique key for an osm element. It contains the
// type, id and version information.
type ElementID int64

// Type returns the Type for the element.
func (id ElementID) Type() Type {
	switch id & typeMask {
	case nodeMask:
		return TypeNode
	case wayMask:
		return TypeWay
	case relationMask:
		return TypeRelation
	}

	panic("unknown type")
}

// Ref return the ID reference for the element. Not unique without the type.
func (id ElementID) Ref() int64 {
	return int64((id & refMask) >> versionBits)
}

// Version returns the version of the element.
func (id ElementID) Version() int {
	return int(id & (versionMask))
}

// ObjectID is a helper to convert the id to an object id.
func (id ElementID) ObjectID() ObjectID {
	return ObjectID(id)
}

// FeatureID returns the feature id for the element id. i.e removing the version.
func (id ElementID) FeatureID() FeatureID {
	return FeatureID(id & featureMask)
}

// NodeID returns the id of this feature as a node id.
// The function will panic if this element is not of TypeNode.
func (id ElementID) NodeID() NodeID {
	if id&nodeMask != nodeMask {
		panic(fmt.Sprintf("not a node: %v", id))
	}

	return NodeID(id.Ref())
}

// WayID returns the id of this feature as a way id.
// The function will panic if this element is not of TypeWay.
func (id ElementID) WayID() WayID {
	if id&wayMask != wayMask {
		panic(fmt.Sprintf("not a way: %v", id))
	}

	return WayID(id.Ref())
}

// RelationID returns the id of this feature as a relation id.
// The function will panic if this element is not of TypeRelation.
func (id ElementID) RelationID() RelationID {
	if int64(id)&relationMask != relationMask {
		panic(fmt.Sprintf("not a relation: %v", id))
	}

	return RelationID(id.Ref())
}

// String returns "type/ref:version" for the element.
func (id ElementID) String() string {
	if id.Version() == 0 {
		return fmt.Sprintf("%s/%d:-", id.Type(), id.Ref())
	}

	return fmt.Sprintf("%s/%d:%d", id.Type(), id.Ref(), id.Version())
}

// ParseElementID takes a string and tries to determine the element id from it.
// The string must be formatted as "type/id:version", the same as the result of the String method.
func ParseElementID(s string) (ElementID, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid element id: %v", s)
	}

	parts2 := strings.Split(parts[1], ":")
	if l := len(parts2); l != 1 && l != 2 {
		return 0, fmt.Errorf("invalid element id: %v", s)
	}

	var version int
	ref, err := strconv.ParseInt(parts2[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid element id: %v: %v", s, err)
	}

	if len(parts2) == 2 && parts2[1] != "-" {
		v, e := strconv.ParseInt(parts2[1], 10, 64)
		if e != nil {
			return 0, fmt.Errorf("invalid element id: %v: %v", s, err)
		}
		version = int(v)
	}

	fid, err := Type(parts[0]).FeatureID(ref)
	if err != nil {
		return 0, fmt.Errorf("invalid element id: %v: %v", s, err)
	}

	return fid.ElementID(version), nil
}

// An Element represents a Node, Way or Relation.
type Element interface {
	Object

	ElementID() ElementID
	FeatureID() FeatureID
	TagMap() map[string]string

	// TagMap keeps waynodes and members from matching the interface.
	// This keeps the meaning of what an element is.
}

// Elements is a collection of the Element type.
type Elements []Element

// ElementIDs returns a slice of the element ids of the elements.
func (es Elements) ElementIDs() ElementIDs {
	if len(es) == 0 {
		return nil
	}

	ids := make(ElementIDs, 0, len(es))
	for _, e := range es {
		ids = append(ids, e.ElementID())
	}

	return ids
}

// FeatureIDs returns a slice of the feature ids of the elements.
func (es Elements) FeatureIDs() FeatureIDs {
	if len(es) == 0 {
		return nil
	}

	ids := make(FeatureIDs, 0, len(es))
	for _, e := range es {
		ids = append(ids, e.FeatureID())
	}

	return ids
}

// Sort will order the elements by type, node, way, relation, changeset,
// then id and lastly the version.
func (es Elements) Sort() {
	sort.Sort(elementsSort(es))
}

type elementsSort Elements

func (es elementsSort) Len() int      { return len(es) }
func (es elementsSort) Swap(i, j int) { es[i], es[j] = es[j], es[i] }
func (es elementsSort) Less(i, j int) bool {
	return es[i].ElementID() < es[j].ElementID()
}

// ElementIDs is a list of element ids with helper functions on top.
type ElementIDs []ElementID

// Counts returns the number of each type of element in the set of ids.
func (ids ElementIDs) Counts() (nodes, ways, relations int) {
	for _, id := range ids {
		switch id & typeMask {
		case nodeMask:
			nodes++
		case wayMask:
			ways++
		case relationMask:
			relations++
		}
	}

	return
}

type elementIDsSort ElementIDs

// Sort will order the ids by type, node, way, relation, changeset,
// and then id.
func (ids ElementIDs) Sort() {
	sort.Sort(elementIDsSort(ids))
}

func (ids elementIDsSort) Len() int      { return len(ids) }
func (ids elementIDsSort) Swap(i, j int) { ids[i], ids[j] = ids[j], ids[i] }
func (ids elementIDsSort) Less(i, j int) bool {
	return ids[i] < ids[j]
}
