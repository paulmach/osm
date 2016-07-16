package osm

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/paulmach/go.osm/osmpb"
)

// ChangesetID is the primary key for a osm changeset.
type ChangesetID int

// Changesets is a collection with some helper functions attached.
type Changesets []*Changeset

// A Changeset is a set of metadata around a set of osm changes.
type Changeset struct {
	ID            ChangesetID         `xml:"id,attr"`
	User          string              `xml:"user,attr"`
	UserID        UserID              `xml:"uid,attr"`
	CreatedAt     time.Time           `xml:"created_at,attr"`
	ClosedAt      time.Time           `xml:"closed_at,attr"`
	Open          bool                `xml:"open,attr"`
	ChangesCount  int                 `xml:"num_changes,attr"`
	MinLat        float64             `xml:"min_lat,attr"`
	MaxLat        float64             `xml:"max_lat,attr"`
	MinLng        float64             `xml:"min_lon,attr"`
	MaxLng        float64             `xml:"max_lon,attr"`
	CommentsCount int                 `xml:"comments_count,attr"`
	Tags          Tags                `xml:"tag"`
	Discussion    ChangesetDiscussion `xml:"discussion"`

	Change *Change
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

	userSid := ss.Add(c.User)
	keys, vals := c.Tags.KeyValues(ss)

	encoded := &osmpb.Changeset{
		Id:        proto.Int64(int64(c.ID)),
		Keys:      keys,
		Vals:      vals,
		Uid:       proto.Int32(int32(c.UserID)),
		UserSid:   &userSid,
		CreatedAt: timeToUnix(c.CreatedAt),
		ClosedAt:  timeToUnix(c.ClosedAt),
		Open:      proto.Bool(c.Open),
	}

	if c.MinLat != 0 || c.MaxLat != 0 || c.MinLng != 0 || c.MaxLng != 0 {
		encoded.Bounds = &osmpb.Bounds{
			MinLat: proto.Int64(geoToInt64(c.MinLat)),
			MaxLat: proto.Int64(geoToInt64(c.MaxLat)),
			MinLng: proto.Int64(geoToInt64(c.MinLng)),
			MaxLng: proto.Int64(geoToInt64(c.MaxLng)),
		}
	}

	if c.Change != nil {
		encoded.Change = marshalChange(c.Change, ss, false)
	}

	encoded.Strings = ss.Strings()
	return proto.Marshal(encoded)
}

func UnmarshalChangeset(data []byte) (*Changeset, error) {
	return nil, nil

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
