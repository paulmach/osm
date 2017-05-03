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
	NodeType      Type = "node"
	WayType            = "way"
	RelationType       = "relation"
	ChangesetType      = "changeset"
)

// A FeatureID is a identifier for a feature in OSM.
// It is meant to represent all the versions of a given element.
type FeatureID struct {
	Type Type
	Ref  int64
}

// ElementID is a helper to convert the id to an element id.
func (f FeatureID) ElementID(v int) ElementID {
	return ElementID{
		Type:    f.Type,
		Ref:     f.Ref,
		Version: v,
	}
}

// NodeID returns the id of this feature as a node id.
// The function will panic if this feature is not of NodeType.
func (f FeatureID) NodeID() NodeID {
	if f.Type != NodeType {
		panic(fmt.Sprintf("not a node: %v", f))
	}

	return NodeID(f.Ref)
}

// WayID returns the id of this feature as a way id.
// The function will panic if this feature is not of WayType.
func (f FeatureID) WayID() WayID {
	if f.Type != WayType {
		panic(fmt.Sprintf("not a way: %v", f))
	}

	return WayID(f.Ref)
}

// RelationID returns the id of this feature as a relation id.
// The function will panic if this feature is not of RelationType.
func (f FeatureID) RelationID() RelationID {
	if f.Type != RelationType {
		panic(fmt.Sprintf("not a relation: %v", f))
	}

	return RelationID(f.Ref)
}

// ChangesetID returns the id of this feature as a changeset id.
// The function will panic if this feature is not of ChangesetType.
func (f FeatureID) ChangesetID() ChangesetID {
	if f.Type != ChangesetType {
		panic(fmt.Sprintf("not a changeset: %v", f))
	}

	return ChangesetID(f.Ref)
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

type featureIDsSort FeatureIDs

// Sort will order the ids by type, node, way, relation, changeset,
// and then id.
func (ids FeatureIDs) Sort() {
	sort.Sort(featureIDsSort(ids))
}

func (ids featureIDsSort) Len() int      { return len(ids) }
func (ids featureIDsSort) Swap(i, j int) { ids[i], ids[j] = ids[j], ids[i] }
func (ids featureIDsSort) Less(i, j int) bool {
	a := ids[i]
	b := ids[j]

	if a.Type != b.Type {
		return typeToNumber[a.Type] < typeToNumber[b.Type]
	}

	return a.Ref < b.Ref
}

var typeToNumber = map[Type]int{
	NodeType:      1,
	WayType:       2,
	RelationType:  3,
	ChangesetType: 4,
}
