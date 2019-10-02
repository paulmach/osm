package osm

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/paulmach/orb"
)

// RelationID is the primary key of a relation.
// A relation is uniquely identifiable by the id + version.
type RelationID int64

// ObjectID is a helper returning the object id for this relation id.
func (id RelationID) ObjectID(v int) ObjectID {
	return ObjectID(id.ElementID(v))
}

// FeatureID is a helper returning the feature id for this relation id.
func (id RelationID) FeatureID() FeatureID {
	return FeatureID((relationMask | id<<versionBits))
}

// ElementID is a helper to convert the id to an element id.
func (id RelationID) ElementID(v int) ElementID {
	return id.FeatureID().ElementID(v)
}

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
	Committed *time.Time `xml:"committed,attr,omitempty" json:"committed,omitempty"`

	// Updates are changes to the members of this relation independent
	// of an update to the relation itself. The OSM api allows a child
	// to be updated without any changes to the parent.
	Updates Updates `xml:"update,omitempty" json:"updates,omitempty"`

	// Bounds are included by overpass, and maybe others
	Bounds *Bounds `xml:"bounds,omitempty" json:"bounds,omitempty"`
}

// Members represents an ordered list of relation members.
type Members []Member

// Member is a member of a relation.
type Member struct {
	Type Type   `xml:"type,attr" json:"type"`
	Ref  int64  `xml:"ref,attr" json:"ref"`
	Role string `xml:"role,attr" json:"role"`

	Version     int         `xml:"version,attr,omitempty" json:"version,omitempty"`
	ChangesetID ChangesetID `xml:"changeset,attr,omitempty" json:"changeset,omitempty"`

	// Node location if Type == Node
	// Closest vertex to centroid if Type == Way
	// Empty/invalid if Type == Relation
	Lat float64 `xml:"lat,attr,omitempty" json:"lat,omitempty"`
	Lon float64 `xml:"lon,attr,omitempty" json:"lon,omitempty"`

	// Orientation is the direction of the way around a ring of a multipolygon.
	// Only valid for multipolygon or boundary relations.
	Orientation orb.Orientation `xml:"orienation,attr,omitempty" json:"orienation,omitempty"`
}

// ObjectID returns the object id of the relation.
func (r *Relation) ObjectID() ObjectID {
	return r.ID.ObjectID(r.Version)
}

// FeatureID returns the feature id of the relation.
func (r *Relation) FeatureID() FeatureID {
	return r.ID.FeatureID()
}

// ElementID returns the element id of the relation.
func (r *Relation) ElementID() ElementID {
	return r.ID.ElementID(r.Version)
}

// FeatureID returns the feature id of the member.
func (m Member) FeatureID() FeatureID {
	switch m.Type {
	case TypeNode:
		return NodeID(m.Ref).FeatureID()
	case TypeWay:
		return WayID(m.Ref).FeatureID()
	case TypeRelation:
		return RelationID(m.Ref).FeatureID()
	}

	panic("unknown type")
}

// ElementID returns the element id of the member.
func (m Member) ElementID() ElementID {
	return m.FeatureID().ElementID(m.Version)
}

// Point returns the orb.Point location for the member.
// Will be (0, 0) if the relation is not annotated.
// For way members this location is annotated as the "surface point".
func (m Member) Point() orb.Point {
	return orb.Point{m.Lon, m.Lat}
}

// CommittedAt returns the best estimate on when this element
// became was written/committed into the database.
func (r *Relation) CommittedAt() time.Time {
	if r.Committed != nil {
		return *r.Committed
	}

	return r.Timestamp
}

// TagMap returns the element tags as a key/value map.
func (r *Relation) TagMap() map[string]string {
	return r.Tags.Map()
}

// ApplyUpdatesUpTo will apply the updates to this object upto and including
// the given time.
func (r *Relation) ApplyUpdatesUpTo(t time.Time) error {
	var notApplied []Update
	for _, u := range r.Updates {
		if u.Timestamp.After(t) {
			notApplied = append(notApplied, u)
			continue
		}

		if err := r.applyUpdate(u); err != nil {
			return err
		}
	}

	r.Updates = notApplied
	return nil
}

// applyUpdate will modify the current relation and dictated by the given update.
// Will return UpdateIndexOutOfRangeError if the update index is too large.
func (r *Relation) applyUpdate(u Update) error {
	if u.Index >= len(r.Members) {
		return &UpdateIndexOutOfRangeError{Index: u.Index}
	}

	r.Members[u.Index].Version = u.Version
	r.Members[u.Index].ChangesetID = u.ChangesetID
	r.Members[u.Index].Lat = u.Lat
	r.Members[u.Index].Lon = u.Lon

	if u.Reverse {
		r.Members[u.Index].Orientation *= -1
	}

	return nil
}

// FeatureIDs returns the a list of feature ids for the members.
func (ms Members) FeatureIDs() FeatureIDs {
	ids := make(FeatureIDs, len(ms), len(ms)+1)
	for i, m := range ms {
		ids[i] = m.FeatureID()
	}

	return ids
}

// ElementIDs returns the a list of element ids for the members.
func (ms Members) ElementIDs() ElementIDs {
	ids := make(ElementIDs, len(ms), len(ms)+1)
	for i, m := range ms {
		ids[i] = m.ElementID()
	}

	return ids
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

// IDs returns the ids for all the relations.
func (rs Relations) IDs() []RelationID {
	result := make([]RelationID, len(rs))
	for i, r := range rs {
		result[i] = r.ID
	}

	return result
}

// FeatureIDs returns the feature ids for all the relations.
func (rs Relations) FeatureIDs() FeatureIDs {
	result := make(FeatureIDs, len(rs))
	for i, r := range rs {
		result[i] = r.FeatureID()
	}

	return result
}

// ElementIDs returns the element ids for all the relations.
func (rs Relations) ElementIDs() ElementIDs {
	result := make(ElementIDs, len(rs))
	for i, r := range rs {
		result[i] = r.ElementID()
	}

	return result
}

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
