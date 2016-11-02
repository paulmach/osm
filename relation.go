package osm

import (
	"encoding/xml"
	"sort"
	"time"
)

// RelationID is the primary key of a relation.
// A relation is uniquely identifiable by the id + version.
type RelationID int64

// Relation is an collection of nodes, ways and other relations
// with some defining attributes.
type Relation struct {
	XMLName     xml.Name    `xml:"relation"`
	ID          RelationID  `xml:"id,attr"`
	User        string      `xml:"user,attr"`
	UserID      UserID      `xml:"uid,attr"`
	Visible     bool        `xml:"visible,attr"`
	Version     int         `xml:"version,attr"`
	ChangesetID ChangesetID `xml:"changeset,attr"`
	Timestamp   time.Time   `xml:"timestamp,attr"`

	Tags    Tags     `xml:"tag"`
	Members []Member `xml:"member"`

	// Committed, is the estimated time this object was committed
	// and made visible in the central OSM database.
	Committed *time.Time `xml:"commited,attr,omitempty"`

	// Updates are changes the members of this relation independent
	// of an update to the relation itself. The OSM api allows a child
	// to be updatedwithout any changes to the parent.
	Updates Updates `xml:"update,omitempty"`
}

// Member is a member of a relation.
type Member struct {
	Type ElementType `xml:"type,attr"`
	Ref  int64       `xml:"ref,attr"`
	Role string      `xml:"role,attr"`

	Version     int         `xml:"version,attr,omitempty"`
	ChangesetID ChangesetID `xml:"changeset,attr,omitempty"`
}

// Relations is a collection with some helper functions attached.
type Relations []*Relation

// Marshal encodes the relations using protocol buffers.
func (rs Relations) Marshal() ([]byte, error) {
	o := OSM{
		Relations: rs,
	}

	return o.Marshal()
}

// UnmarshalRelations will unmarshal the data into a list of relations.
func UnmarshalRelations(data []byte) (Relations, error) {
	o, err := UnmarshalOSM(data)
	if err != nil {
		return nil, err
	}

	return o.Relations, nil
}

type relationsSort Relations

// SortIDVersion will sort the set of relations first by id and then version
// in ascending order.
func (rs Relations) SortIDVersion() {
	sort.Sort(relationsSort(rs))
}
func (rs relationsSort) Len() int      { return len(rs) }
func (rs relationsSort) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs relationsSort) Less(i, j int) bool {
	if rs[i].ID == rs[j].ID {
		return rs[i].Version < rs[j].Version
	}

	return rs[i].ID < rs[j].ID
}
