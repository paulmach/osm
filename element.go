package osm

// Scanner allows osm data from dump files to be read.
// It is based on the bufio.Scanner, common usage.
//
//	s := scanner.New(r)
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
}

// An Element is a container for an osm thing that
// could be returned by a scanner.
type Element struct {
	Changeset *Changeset
	Node      *Node
	Way       *Way
	Relation  *Relation
}
