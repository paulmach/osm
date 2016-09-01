package osm

import (
	"encoding/xml"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/paulmach/go.osm/internal/osmpb"
)

// ChangesetID is the primary key for a osm changeset.
type ChangesetID int

// Changesets is a collection with some helper functions attached.
type Changesets []*Changeset

// A Changeset is a set of metadata around a set of osm changes.
type Changeset struct {
	XMLName       xml.Name            `xml:"changeset"`
	ID            ChangesetID         `xml:"id,attr"`
	User          string              `xml:"user,attr"`
	UserID        UserID              `xml:"uid,attr"`
	CreatedAt     time.Time           `xml:"created_at,attr"`
	ClosedAt      time.Time           `xml:"closed_at,attr"`
	Open          bool                `xml:"open,attr"`
	ChangesCount  int                 `xml:"num_changes,attr,omitempty"`
	MinLat        float64             `xml:"min_lat,attr"`
	MaxLat        float64             `xml:"max_lat,attr"`
	MinLon        float64             `xml:"min_lon,attr"`
	MaxLon        float64             `xml:"max_lon,attr"`
	CommentsCount int                 `xml:"comments_count,attr,omitempty"`
	Tags          Tags                `xml:"tag"`
	Discussion    ChangesetDiscussion `xml:"discussion,omitempty"`

	Change *Change `xml:"-"`
}

// Comment is a helper and returns the changeset comment from the tag.
func (c *Changeset) Comment() string {
	return c.Tags.Find("comment")
}

// CreatedBy is a helper and returns the changeset created by from the tag.
func (c *Changeset) CreatedBy() string {
	return c.Tags.Find("created_by")
}

// Locale is a helper and returns the changeset locale from the tag.
func (c *Changeset) Locale() string {
	return c.Tags.Find("locale")
}

// Host is a helper and returns the changeset host from the tag.
func (c *Changeset) Host() string {
	return c.Tags.Find("host")
}

// ImageryUsed is a helper and returns imagery used for the changeset from the tag.
func (c *Changeset) ImageryUsed() string {
	return c.Tags.Find("imagery_used")
}

// Source is a helper and returns source for the changeset from the tag.
func (c *Changeset) Source() string {
	return c.Tags.Find("source")
}

// Bot is a helper and returns true if the bot tag is a yes.
func (c *Changeset) Bot() bool {
	// As of July 5, 2015: 300k yes, 123 no, 8 other
	return c.Tags.Find("bot") == "yes"
}

// Marshal encodes the changeset data using protocol buffers.
// Does not encode the changeset discussion.
func (c *Changeset) Marshal() ([]byte, error) {
	ss := &stringSet{}

	var userSid *uint32
	if c.User != "" {
		v := ss.Add(c.User)
		userSid = &v
	}
	keys, vals := c.Tags.keyValues(ss)

	encoded := &osmpb.Changeset{
		Keys:      keys,
		Vals:      vals,
		UserSid:   userSid,
		CreatedAt: timeToUnixPointer(c.CreatedAt),
		ClosedAt:  timeToUnixPointer(c.ClosedAt),
	}

	// only set these values if they make any sense.
	if c.ID != 0 {
		encoded.Id = proto.Int64(int64(c.ID))
	}

	if c.Open {
		encoded.Open = proto.Bool(c.Open)
	}

	if c.UserID != 0 {
		encoded.UserId = proto.Int32(int32(c.UserID))
	}

	if c.MinLat != 0 || c.MaxLat != 0 || c.MinLon != 0 || c.MaxLon != 0 {
		encoded.Bounds = &osmpb.Bounds{
			MinLat: geoToInt64(c.MinLat),
			MaxLat: geoToInt64(c.MaxLat),
			MinLon: geoToInt64(c.MinLon),
			MaxLon: geoToInt64(c.MaxLon),
		}
	}

	if c.Change != nil &&
		(c.Change.Create != nil || c.Change.Modify != nil || c.Change.Delete != nil) {
		encoded.Change = marshalChange(c.Change, ss, false)
	}

	encoded.Strings = ss.Strings()
	return proto.Marshal(encoded)
}

// UnmarshalChangeset will unmarshal the data into a OSM object.
func UnmarshalChangeset(data []byte) (*Changeset, error) {
	encoded := &osmpb.Changeset{}
	err := proto.Unmarshal(data, encoded)
	if err != nil {
		return nil, err
	}

	ss := encoded.GetStrings()
	tags, err := tagsFromStrings(ss, encoded.GetKeys(), encoded.GetVals())
	if err != nil {
		return nil, err
	}

	cs := &Changeset{
		ID:        ChangesetID(encoded.GetId()),
		UserID:    UserID(encoded.GetUserId()),
		User:      ss[encoded.GetUserSid()],
		CreatedAt: unixToTime(encoded.GetCreatedAt()),
		ClosedAt:  unixToTime(encoded.GetClosedAt()),
		Open:      encoded.GetOpen(),
		Tags:      tags,
	}

	if encoded.Bounds != nil {
		cs.MinLat = float64(encoded.Bounds.GetMinLat()) / locMultiple
		cs.MaxLat = float64(encoded.Bounds.GetMaxLat()) / locMultiple
		cs.MinLon = float64(encoded.Bounds.GetMinLon()) / locMultiple
		cs.MaxLon = float64(encoded.Bounds.GetMaxLon()) / locMultiple
	}

	if encoded.Change != nil {
		cs.Change, err = unmarshalChange(encoded.Change, ss, cs)
		if err != nil {
			return nil, err
		}
	}

	return cs, nil
}

// IDs returns the ids of the changesets in the slice.
func (cs Changesets) IDs() []ChangesetID {
	if len(cs) == 0 {
		return nil
	}

	r := make([]ChangesetID, 0, len(cs))
	for _, c := range cs {
		r = append(r, c.ID)
	}

	return r
}

// ChangesetDiscussion is a conversation about a changeset.
type ChangesetDiscussion struct {
	Comments []*ChangesetComment `xml:"comment"`
}

// ChangesetComment is a specific comment in a changeset discussion.
type ChangesetComment struct {
	User      string    `xml:"user,attr"`
	UserID    UserID    `xml:"uid,attr"`
	CreatedAt time.Time `xml:"date,attr"`
	Text      string    `xml:"text"`
}

// MarshalXML implements the xml.Marshaller method to exclude this
// whole element if the comments are empty.
func (csd ChangesetDiscussion) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(csd.Comments) == 0 {
		return nil
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	t := xml.StartElement{Name: xml.Name{Local: "comment"}}
	if err := e.EncodeElement(csd.Comments, t); err != nil {
		return err
	}

	if err := e.EncodeToken(start.End()); err != nil {
		return err
	}

	return nil
}
