package osm

import (
	"errors"
	"fmt"
	"sort"
)

var (
	// ErrScannerClosed is returned by scanner.Err() if the scanner is closed
	// and there are no other io or xml errors to report.
	ErrScannerClosed = errors.New("osmxml: scanner closed by user")
)

// ElementID is a unique key for an osm element. It contains the
// type, id and version.
type ElementID struct {
	Type    Type
	Ref     int64
	Version int
}

// NodeID returns the id of this feature as a node id.
// The function will panic if this feature is not of NodeType.
func (e ElementID) NodeID() NodeID {
	if e.Type != TypeNode {
		panic(fmt.Sprintf("not a node: %v", e))
	}

	return NodeID(e.Ref)
}

// WayID returns the id of this feature as a way id.
// The function will panic if this feature is not of WayType.
func (e ElementID) WayID() WayID {
	if e.Type != TypeWay {
		panic(fmt.Sprintf("not a way: %v", e))
	}

	return WayID(e.Ref)
}

// RelationID returns the id of this feature as a relation id.
// The function will panic if this feature is not of RelationType.
func (e ElementID) RelationID() RelationID {
	if e.Type != TypeRelation {
		panic(fmt.Sprintf("not a relation: %v", e))
	}

	return RelationID(e.Ref)
}

// ChangesetID returns the id of this feature as a changeset id.
// The function will panic if this feature is not of ChangesetType.
func (e ElementID) ChangesetID() ChangesetID {
	if e.Type != TypeChangeset {
		panic(fmt.Sprintf("not a changeset: %v", e))
	}

	return ChangesetID(e.Ref)
}

// Scanner allows osm data from dump files to be read.
// It is based on the bufio.Scanner, common usage.
// Scanners are not safe for parallel use. One should feed the
// elements into their own channel and have workers read from that.
//
//	s := scanner.New(r)
//  defer s.Close()
//
//	for s.Next() {
//		e := s.Element()
//		// do something
//	}
//
//	if s.Err() != nil {
//		// scanner did no complete fully
//	}
type Scanner interface {
	Scan() bool
	Element() Element
	Err() error
	Close() error
}

// An Element represents a Node, Way, Relation or Changeset.
type Element interface {
	FeatureID() FeatureID
	ElementID() ElementID
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

// Sort will order the elements by type, node, way, relation, changeset,
// then id and lastly the version.
func (es Elements) Sort() {
	sort.Sort(elementsSort(es))
}

type elementsSort Elements

func (es elementsSort) Len() int      { return len(es) }
func (es elementsSort) Swap(i, j int) { es[i], es[j] = es[j], es[i] }
func (es elementsSort) Less(i, j int) bool {
	a := es[i].ElementID()
	b := es[j].ElementID()
	if a.Type != b.Type {
		return typeToNumber[a.Type] < typeToNumber[b.Type]
	}

	if a.Ref != b.Ref {
		return a.Ref < b.Ref
	}

	return a.Version < b.Version
}

// ElementIDs is a list of element ids with helper functions on top.
type ElementIDs []ElementID

type elementIDsSort ElementIDs

// Sort will order the ids by type, node, way, relation, changeset,
// and then id.
func (ids ElementIDs) Sort() {
	sort.Sort(elementIDsSort(ids))
}

func (ids elementIDsSort) Len() int      { return len(ids) }
func (ids elementIDsSort) Swap(i, j int) { ids[i], ids[j] = ids[j], ids[i] }
func (ids elementIDsSort) Less(i, j int) bool {
	a := ids[i]
	b := ids[j]

	if a.Type != b.Type {
		return typeToNumber[a.Type] < typeToNumber[b.Type]
	}

	if a.Ref != b.Ref {
		return a.Ref < b.Ref
	}

	return a.Version < b.Version
}
