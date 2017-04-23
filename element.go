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

// Type is the type of different osm elements.
// ie. node, way, relation
type Type string

// Enums for the different element types.
const (
	NodeType      Type = "node"
	WayType            = "way"
	RelationType       = "relation"
	ChangesetType      = "changeset"
)

// An ElementID is a identifier that maps a thing in osm
// to a unique id. Note, does not include the element version.
type ElementID struct {
	Type Type
	Ref  int64
}

// NodeID returns the id of this element as a node id.
// The function will panic if this element is not of NodeType.
func (e ElementID) NodeID() NodeID {
	if e.Type != NodeType {
		panic(fmt.Sprintf("element %v is not a node", e))
	}

	return NodeID(e.Ref)
}

// WayID returns the id of this element as a way id.
// The function will panic if this element is not of WayType.
func (e ElementID) WayID() WayID {
	if e.Type != WayType {
		panic(fmt.Sprintf("element %v is not a way", e))
	}

	return WayID(e.Ref)
}

// RelationID returns the id of this element as a relation id.
// The function will panic if this element is not of RelationType.
func (e ElementID) RelationID() RelationID {
	if e.Type != RelationType {
		panic(fmt.Sprintf("element %v is not a relation", e))
	}

	return RelationID(e.Ref)
}

// ChangesetID returns the id of this element as a changeset id.
// The function will panic if this element is not of ChangesetType.
func (e ElementID) ChangesetID() ChangesetID {
	if e.Type != ChangesetType {
		panic(fmt.Sprintf("element %v is not a changeset", e))
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
	ElementID() ElementID

	// WayNode and Member also have the above functions but are
	// not elements. This is to help keep the meanings.
	private()
}

func (n *Node) private()      {}
func (w *Way) private()       {}
func (r *Relation) private()  {}
func (c *Changeset) private() {}

// Elements is a collection of the Element type.
type Elements []Element

type elementsSort Elements

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
// and then id.
func (es Elements) Sort() {
	sort.Sort(elementsSort(es))
}

func (es elementsSort) Len() int      { return len(es) }
func (es elementsSort) Swap(i, j int) { es[i], es[j] = es[j], es[i] }
func (es elementsSort) Less(i, j int) bool {
	a := es[i].ElementID()
	b := es[j].ElementID()
	if a.Type != b.Type {
		return typeToNumber[a.Type] < typeToNumber[b.Type]
	}

	return a.Ref < b.Ref
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

	return a.Ref < b.Ref
}

var typeToNumber = map[Type]int{
	NodeType:      1,
	WayType:       2,
	RelationType:  3,
	ChangesetType: 4,
}
