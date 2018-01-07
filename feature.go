package osm

import (
	"fmt"
	"sort"
)

// Type is the type of different osm elements.
// ie. node, way, relation
type Type string

// Constants for the different element types.
const (
	TypeNode      Type = "node"
	TypeWay       Type = "way"
	TypeRelation  Type = "relation"
	TypeChangeset Type = "changeset"
)

// FeatureID returns a feature id from the given type.
func (t Type) FeatureID(ref int64) FeatureID {
	switch t {
	case TypeNode:
		return NodeID(ref).FeatureID()
	case TypeWay:
		return WayID(ref).FeatureID()
	case TypeRelation:
		return RelationID(ref).FeatureID()
	case TypeChangeset:
		return ChangesetID(ref).FeatureID()
	}

	panic(fmt.Sprintf("unknown type: %v", t))
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
	case changesetMask:
		return TypeChangeset
	}

	panic("unknown type")
}

// Ref return the ID reference for the feature. Not unique without the type.
func (f FeatureID) Ref() int64 {
	return int64((f & refMask) >> versionBits)
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

// ChangesetID returns the id of this feature as a changeset id.
// The function will panic if this feature is not of ChangesetType.
func (f FeatureID) ChangesetID() ChangesetID {
	if f&changesetMask == 0 {
		panic(fmt.Sprintf("not a changeset: %v", f))
	}

	return ChangesetID(f.Ref())
}

// String returns "type/ref" for the feature.
func (f FeatureID) String() string {
	return fmt.Sprintf("%s/%d", f.Type(), f.Ref())
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
func (ids FeatureIDs) Counts() (nodes, ways, relations, changesets int) {
	for _, id := range ids {
		switch id.Type() {
		case TypeNode:
			nodes++
		case TypeWay:
			ways++
		case TypeRelation:
			relations++
		case TypeChangeset:
			changesets++
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
