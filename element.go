package osm

import "errors"

var (
	// ErrScannerClosed is returned by scanner.Err() if the scanner is closed
	// and there are no other io or xml errors to report.
	ErrScannerClosed = errors.New("osmxml: scanner closed by user")
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

// An Element is a container for an osm thing that
// could be returned by a scanner.
type Element struct {
	Changeset *Changeset
	Node      *Node
	Way       *Way
	Relation  *Relation
}

// ElementType is the type of different osm elements.
// ie. node, way, relation
type ElementType string

// Enums for the different element types.
const (
	NodeType     ElementType = "node"
	WayType                  = "way"
	RelationType             = "relation"
)
