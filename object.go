package osm

import (
	"fmt"
	"strconv"
	"strings"
)

// ObjectID encodes the type and ref of an osm object,
// e.g. nodes, ways, relations, changesets, notes and users.
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
	case boundsMask:
		return TypeBounds
	}

	panic("unknown type")
}

// Ref returns the ID reference for the object. Not unique without the type.
func (id ObjectID) Ref() int64 {
	return int64((id & refMask) >> versionBits)
}

// Version returns the version of the object.
// Will return 0 if the object doesn't have versions like users, notes and changesets.
func (id ObjectID) Version() int {
	return int(id & (versionMask))
}

// String returns "type/ref:version" for the object.
func (id ObjectID) String() string {
	if id.Version() == 0 {
		return fmt.Sprintf("%s/%d:-", id.Type(), id.Ref())
	}

	return fmt.Sprintf("%s/%d:%d", id.Type(), id.Ref(), id.Version())
}

// ParseObjectID takes a string and tries to determine the object id from it.
// The string must be formatted as "type/id:version", the same as the result of the String method.
func ParseObjectID(s string) (ObjectID, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid element id: %v", s)
	}

	parts2 := strings.Split(parts[1], ":")
	if l := len(parts2); l == 0 || l > 2 {
		return 0, fmt.Errorf("invalid element id: %v", s)
	}

	var version int
	ref, err := strconv.ParseInt(parts2[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid element id: %v: %v", s, err)
	}

	if len(parts2) == 2 && parts2[1] != "-" {
		v, e := strconv.ParseInt(parts2[1], 10, 64)
		if e != nil {
			return 0, fmt.Errorf("invalid element id: %v: %v", s, err)
		}
		version = int(v)
	}

	oid, err := Type(parts[0]).objectID(ref, version)
	if err != nil {
		return 0, fmt.Errorf("invalid element id: %v: %v", s, err)
	}

	return oid, nil
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
func (b *Bounds) private()    {}

// Objects is a set of objects with some helpers
type Objects []Object

// ObjectIDs returns a slice of the object ids of the osm objects.
func (os Objects) ObjectIDs() ObjectIDs {
	if len(os) == 0 {
		return nil
	}

	ids := make(ObjectIDs, 0, len(os))
	for _, o := range os {
		ids = append(ids, o.ObjectID())
	}

	return ids
}

// ObjectIDs is a slice of ObjectIDs with some helpers on top.
type ObjectIDs []ObjectID

// A Scanner reads osm data from planet dump files.
// It is based on the bufio.Scanner, common usage.
// Scanners are not safe for parallel use. One should feed the
// objects into their own channel and have workers read from that.
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
//		// scanner did not complete fully
//	}
type Scanner interface {
	Scan() bool
	Object() Object
	Err() error
	Close() error
}
