package osm

// ObjectID encodes the type and ref of an osm object, e.g. nodes, ways, relations, changesets, notes and users.
type ObjectID int64

// Type returns the Type of the object.
func (id ObjectID) Type() Type {
	switch id & typeMask {
	case nodeMask:
		return TypeNode
	case wayMask:
		return TypeWay
	case relationMask:
		return TypeRelation
	case changesetMask:
		return TypeChangeset
	case noteMask:
		return TypeNote
	case userMask:
		return TypeUser
	}

	panic("unknown type")
}

// Ref return the ID reference for the object. Not unique without the type.
func (id ObjectID) Ref() int64 {
	return int64((id & refMask) >> versionBits)
}

// An Object represents a Node, Way, Relation, Changeset, Note or User only.
type Object interface {
	ObjectID() ObjectID

	// private is so that **ID types don't implement this interface.
	private()
}

func (n *Node) private()      {}
func (w *Way) private()       {}
func (r *Relation) private()  {}
func (c *Changeset) private() {}
func (n *Note) private()      {}
func (u *User) private()      {}

// Objects is a set of objects with some helpers
type Objects []Object

// ObjectIDs is a slice of ObjectIDs with some helpers on top.
type ObjectIDs []ObjectID

// Scanner allows osm data from dump files to be read.
// It is based on the bufio.Scanner, common usage.
// Scanners are not safe for parallel use. One should feed the
// elements into their own channel and have workers read from that.
//
//	s := scanner.New(r)
//	defer s.Close()
//
//	for s.Next() {
//		o := s.Object()
//		// do something
//	}
//
//	if s.Err() != nil {
//		// scanner did no complete fully
//	}
type Scanner interface {
	Scan() bool
	Object() Object
	Err() error
	Close() error
}
