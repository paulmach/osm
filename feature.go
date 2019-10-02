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
	TypeBounds    Type = "bounds"
)

// objectID returns an object id from the given type.
func (t Type) objectID(ref int64, v int) (ObjectID, error) {
	switch t {
	case TypeNode:
		return NodeID(ref).ObjectID(v), nil
	case TypeWay:
		return WayID(ref).ObjectID(v), nil
	case TypeRelation:
		return RelationID(ref).ObjectID(v), nil
	case TypeChangeset:
		return ChangesetID(ref).ObjectID(), nil
	case TypeNote:
		return NoteID(ref).ObjectID(), nil
	case TypeUser:
		return UserID(ref).ObjectID(), nil
	case TypeBounds:
		var b *Bounds
		return b.ObjectID(), nil
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

	refMask     = 0x00FFFFFFFFFF0000
	featureMask = 0x7FFFFFFFFFFF0000
	typeMask    = 0x7F00000000000000

	boundsMask    = 0x0800000000000000
	nodeMask      = 0x1000000000000000
	wayMask       = 0x2000000000000000
	relationMask  = 0x3000000000000000
	changesetMask = 0x4000000000000000
	noteMask      = 0x5000000000000000
	userMask      = 0x6000000000000000
)

// A FeatureID is an identifier for a feature in OSM.
// It is meant to represent all the versions of a given element.
type FeatureID int64

// Type returns the Type of the feature.
// Returns empty string for invalid type.
func (id FeatureID) Type() Type {
	switch id & typeMask {
	case nodeMask:
		return TypeNode
	case wayMask:
		return TypeWay
	case relationMask:
		return TypeRelation
	}

	return ""
}

// Ref return the ID reference for the feature. Not unique without the type.
func (id FeatureID) Ref() int64 {
	return int64((id & refMask) >> versionBits)
}

// ObjectID is a helper to convert the id to an object id.
func (id FeatureID) ObjectID(v int) ObjectID {
	return ObjectID(id.ElementID(v))
}

// ElementID is a helper to convert the id to an element id.
func (id FeatureID) ElementID(v int) ElementID {
	return ElementID(id | (versionMask & FeatureID(v)))
}

// NodeID returns the id of this feature as a node id.
// The function will panic if this feature is not of TypeNode..
func (id FeatureID) NodeID() NodeID {
	if id&nodeMask != nodeMask {
		panic(fmt.Sprintf("not a node: %v", id))
	}

	return NodeID(id.Ref())
}

// WayID returns the id of this feature as a way id.
// The function will panic if this feature is not of TypeWay.
func (id FeatureID) WayID() WayID {
	if id&wayMask != wayMask {
		panic(fmt.Sprintf("not a way: %v", id))
	}

	return WayID(id.Ref())
}

// RelationID returns the id of this feature as a relation id.
// The function will panic if this feature is not of TypeRelation.
func (id FeatureID) RelationID() RelationID {
	if id&relationMask != relationMask {
		panic(fmt.Sprintf("not a relation: %v", id))
	}

	return RelationID(id.Ref())
}

// String returns "type/ref" for the feature.
func (id FeatureID) String() string {
	t := Type("unknown")
	switch id & typeMask {
	case nodeMask:
		t = TypeNode
	case wayMask:
		t = TypeWay
	case relationMask:
		t = TypeRelation
	}
	return fmt.Sprintf("%s/%d", t, id.Ref())
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
