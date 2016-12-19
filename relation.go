package osm

import (
	"encoding/json"
	"sort"
	"time"
)

// RelationID is the primary key of a relation.
// A relation is uniquely identifiable by the id + version.
type RelationID int64

// Relation is an collection of nodes, ways and other relations
// with some defining attributes.
type Relation struct {
	XMLName     xmlNameJSONTypeRel `xml:"relation" json:"type"`
	ID          RelationID         `xml:"id,attr" json:"id"`
	User        string             `xml:"user,attr" json:"user,omitempty"`
	UserID      UserID             `xml:"uid,attr" json:"uid,omitempty"`
	Visible     bool               `xml:"visible,attr" json:"visible"`
	Version     int                `xml:"version,attr" json:"version,omitempty"`
	ChangesetID ChangesetID        `xml:"changeset,attr" json:"changeset,omitempty"`
	Timestamp   time.Time          `xml:"timestamp,attr" json:"timestamp,omitempty"`

	Tags    Tags    `xml:"tag" json:"tags,omitempty"`
	Members Members `xml:"member" json:"members"`

	// Committed, is the estimated time this object was committed
	// and made visible in the central OSM database.
	Committed *time.Time `xml:"commited,attr,omitempty" json:"committed,omitempty"`

	// Updates are changes the members of this relation independent
	// of an update to the relation itself. The OSM api allows a child
	// to be updatedwithout any changes to the parent.
	Updates Updates `xml:"update,omitempty" json:"updates,omitempty"`
}

// Members represents an ordered list of relation members.
type Members []Member

// Member is a member of a relation.
type Member struct {
	Type ElementType `xml:"type,attr" json:"type"`
	Ref  int64       `xml:"ref,attr" json:"ref"`
	Role string      `xml:"role,attr" json:"role"`

	Version     int         `xml:"version,attr,omitempty" json:"version,omitempty"`
	ChangesetID ChangesetID `xml:"changeset,attr,omitempty" json:"changeset,omitempty"`

	// invalid unless the Type == NodeType
	Lat float64 `xml:"lat,attr,omitempty" json:"lat,omitempty"`
	Lon float64 `xml:"lon,attr,omitempty" json:"lon,omitempty"`
}

// ElementID returns the element id of the relation.
func (r *Relation) ElementID() ElementID {
	return ElementID{
		Type:    RelationType,
		ID:      int64(r.ID),
		Version: r.Version,
	}
}

// CommittedAt returns the best estimate on when this element
// became was written/committed into the database.
func (r *Relation) CommittedAt() time.Time {
	if r.Committed != nil {
		return *r.Committed
	}

	return r.Timestamp
}

// ApplyUpdatesUpTo will apply the updates to this object upto and including
// the given time.
func (r *Relation) ApplyUpdatesUpTo(t time.Time) error {
	for _, u := range r.Updates {
		if u.Timestamp.After(t) {
			continue
		}

		if err := r.ApplyUpdate(u); err != nil {
			return err
		}
	}

	return nil
}

// ApplyUpdate will modify the current relation and dictated by the given update.
// Will return UpdateIndexOutOfRangeError if the update index is too large.
func (r *Relation) ApplyUpdate(u Update) error {
	if u.Index >= len(r.Members) {
		return &UpdateIndexOutOfRangeError{Index: u.Index}
	}

	r.Members[u.Index].Version = u.Version
	r.Members[u.Index].ChangesetID = u.ChangesetID
	r.Members[u.Index].Lat = u.Lat
	r.Members[u.Index].Lon = u.Lon

	return nil
}

// MarshalJSON allows the members to be marshalled as defined by the
// overpass osmjson. This function is a wrapper to marshal null as [].
func (ms Members) MarshalJSON() ([]byte, error) {
	if len(ms) == 0 {
		return []byte(`[]`), nil
	}

	return json.Marshal([]Member(ms))
}

// Relations is a list of relations with some helper functions attached.
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

// SortByIDVersion will sort the set of relations first by id and then version
// in ascending order.
func (rs Relations) SortByIDVersion() {
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
