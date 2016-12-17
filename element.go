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

// ElementType is the type of different osm elements.
// ie. node, way, relation
type ElementType string

// Enums for the different element types.
const (
	NodeType      ElementType = "node"
	WayType                   = "way"
	RelationType              = "relation"
	ChangesetType             = "changeset"
)

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
	ElementID() ElementID
}

// Elements is a collection of the Element type.
type Elements []Element

type elementsSort Elements

// Sort will order the elements by type, node, way, relation, changeset,
// and then id and version.
func (es Elements) Sort() {
	sort.Sort(elementsSort(es))
}

func (es elementsSort) Len() int      { return len(es) }
func (es elementsSort) Swap(i, j int) { es[i], es[j] = es[j], es[i] }
func (es elementsSort) Less(i, j int) bool {
	return compIDs(es[i].ElementID(), es[j].ElementID())
}

// An ElementID is a identifier that maps a thing is osm
// to a unique id.
type ElementID struct {
	Type    ElementType
	ID      int64
	Version int
}

// NodeID returns the id of this element as a node id.
// The function will panic if this element is not of NodeType.
func (e ElementID) NodeID() NodeID {
	if e.Type != NodeType {
		panic(fmt.Sprintf("element %v is not a node", e))
	}

	return NodeID(e.ID)
}

// WayID returns the id of this element as a way id.
// The function will panic if this element is not of WayType.
func (e ElementID) WayID() WayID {
	if e.Type != WayType {
		panic(fmt.Sprintf("element %v is not a way", e))
	}

	return WayID(e.ID)
}

// RelationID returns the id of this element as a relation id.
// The function will panic if this element is not of RelationType.
func (e ElementID) RelationID() RelationID {
	if e.Type != RelationType {
		panic(fmt.Sprintf("element %v is not a relation", e))
	}

	return RelationID(e.ID)
}

// ChangesetID returns the id of this element as a changeset id.
// The function will panic if this element is not of ChangesetType.
func (e ElementID) ChangesetID() ChangesetID {
	if e.Type != ChangesetType {
		panic(fmt.Sprintf("element %v is not a changeset", e))
	}

	return ChangesetID(e.ID)
}

// ElementIDs is a list of element ids with helper functions on top.
type ElementIDs []ElementID

type elementIDsSort ElementIDs

// Sort will order the ids by type, node, way, relation, changeset,
// and then id and version.
func (ids ElementIDs) Sort() {
	sort.Sort(elementIDsSort(ids))
}

func (ids elementIDsSort) Len() int      { return len(ids) }
func (ids elementIDsSort) Swap(i, j int) { ids[i], ids[j] = ids[j], ids[i] }
func (ids elementIDsSort) Less(i, j int) bool {
	return compIDs(ids[i], ids[j])
}

func compIDs(a, b ElementID) bool {
	if a.Type != b.Type {
		return typeToNumber[a.Type] < typeToNumber[b.Type]
	}

	if a.ID != b.ID {
		return a.ID < b.ID
	}

	return a.Version < b.Version
}

var typeToNumber = map[ElementType]int{
	NodeType:      1,
	WayType:       2,
	RelationType:  3,
	ChangesetType: 4,
}
