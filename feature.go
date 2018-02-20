package osm

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Type is the type of different osm objects.
// ie. node, way, relation, changeset, note, user.
type Type string

// Constants for the different object types.
const (
	TypeNode      Type = "node"
	TypeWay       Type = "way"
	TypeRelation  Type = "relation"
	TypeChangeset Type = "changeset"
	TypeNote      Type = "note"
	TypeUser      Type = "user"
)

// ObjectID returns an object id from the given type.
func (t Type) ObjectID(ref int64) (ObjectID, error) {
	switch t {
	case TypeNode:
		return NodeID(ref).ObjectID(), nil
	case TypeWay:
		return WayID(ref).ObjectID(), nil
	case TypeRelation:
		return RelationID(ref).ObjectID(), nil
	case TypeChangeset:
		return ChangesetID(ref).ObjectID(), nil
	case TypeNote:
		return NoteID(ref).ObjectID(), nil
	case TypeUser:
		return UserID(ref).ObjectID(), nil
	}

	return 0, fmt.Errorf("unknown type: %v", t)
}

// FeatureID returns a feature id from the given type.
func (t Type) FeatureID(ref int64) (FeatureID, error) {
	switch t {
	case TypeNode:
		return NodeID(ref).FeatureID(), nil
	case TypeWay:
		return WayID(ref).FeatureID(), nil
	case TypeRelation:
		return RelationID(ref).FeatureID(), nil
	}

	return 0, fmt.Errorf("unknown type: %v", t)
}

const (
	versionBits = 16
	versionMask = 0x000000000000FFFF

	refMask     = 0x0FFFFFFFFFFF0000
	featureMask = 0x7FFFFFFFFFFF0000
	typeMask    = 0x7000000000000000

	nodeMask      = 0x1000000000000000
	wayMask       = 0x2000000000000000
	relationMask  = 0x3000000000000000
	changesetMask = 0x4000000000000000
	noteMask      = 0x5000000000000000
	userMask      = 0x6000000000000000
)

// A FeatureID is a identifier for a feature in OSM.
// It is meant to represent all the versions of a given element.
type FeatureID int64

// Type returns the Type of the feature.
func (f FeatureID) Type() Type {
	switch f & typeMask {
	case nodeMask:
		return TypeNode
	case wayMask:
		return TypeWay
	case relationMask:
		return TypeRelation
	}

	panic("unknown type")
}

// Ref return the ID reference for the feature. Not unique without the type.
func (f FeatureID) Ref() int64 {
	return int64((f & refMask) >> versionBits)
}

// ObjectID is a helper to convert the id to an object id.
func (f FeatureID) ObjectID() ObjectID {
	return ObjectID(f)
}

// ElementID is a helper to convert the id to an element id.
func (f FeatureID) ElementID(v int) ElementID {
	return ElementID(f | (versionMask & FeatureID(v)))
}

// NodeID returns the id of this feature as a node id.
// The function will panic if this feature is not of NodeType.
func (f FeatureID) NodeID() NodeID {
	if f&nodeMask == 0 {
		panic(fmt.Sprintf("not a node: %v", f))
	}

	return NodeID(f.Ref())
}

// WayID returns the id of this feature as a way id.
// The function will panic if this feature is not of WayType.
func (f FeatureID) WayID() WayID {
	if f&wayMask == 0 {
		panic(fmt.Sprintf("not a way: %v", f))
	}

	return WayID(f.Ref())
}

// RelationID returns the id of this feature as a relation id.
// The function will panic if this feature is not of RelationType.
func (f FeatureID) RelationID() RelationID {
	if f&relationMask == 0 {
		panic(fmt.Sprintf("not a relation: %v", f))
	}

	return RelationID(f.Ref())
}

// String returns "type/ref" for the feature.
func (f FeatureID) String() string {
	t := Type("unknown")
	switch f & typeMask {
	case nodeMask:
		t = TypeNode
	case wayMask:
		t = TypeWay
	case relationMask:
		t = TypeRelation
	}
	return fmt.Sprintf("%s/%d", t, f.Ref())
}

// ParseFeatureID takes a string and tries to determine the feature id from it.
// The string must be formatted at "type/id", the same as the result of the String method.
func ParseFeatureID(s string) (FeatureID, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid feature id: %v", s)
	}

	n, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid feature id: %v: %v", s, err)
	}

	id, err := Type(parts[0]).FeatureID(n)
	if err != nil {
		return 0, fmt.Errorf("invalid feature id: %s: %v", s, err)
	}

	return id, nil
}

// FeatureIDs is a slice of FeatureIDs with some helpers on top.
type FeatureIDs []FeatureID

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

// Counts returns the number of each type of feature in the set of ids.
func (ids FeatureIDs) Counts() (nodes, ways, relations int) {
	for _, id := range ids {
		switch id.Type() {
		case TypeNode:
			nodes++
		case TypeWay:
			ways++
		case TypeRelation:
			relations++
		}
	}

	return
}

type featureIDsSort FeatureIDs

// Sort will order the ids by type, node, way, relation, changeset,
// and then id.
func (ids FeatureIDs) Sort() {
	sort.Sort(featureIDsSort(ids))
}

func (ids featureIDsSort) Len() int      { return len(ids) }
func (ids featureIDsSort) Swap(i, j int) { ids[i], ids[j] = ids[j], ids[i] }
func (ids featureIDsSort) Less(i, j int) bool {
	return ids[i] < ids[j]
}
