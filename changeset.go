package osm

import (
	"encoding/xml"
	"time"

	"github.com/paulmach/go.geo"
)

type Changesets struct {
	XMLName    xml.Name     `xml:"osm"`
	Changesets []*Changeset `xml:"changeset"`
}

type Changeset struct {
	XMLName       xml.Name            `xml:"changeset"`
	ID            int                 `xml:"id,attr"`
	User          string              `xml:"user,attr"`
	UserID        int                 `xml:"uid,attr"`
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
}

func (c *Changeset) Bound() *geo.Bound {
	return geo.NewBound(c.MinLng, c.MaxLng, c.MinLat, c.MaxLat)
}

func (c *Changeset) Comment() string {
	return c.Tags.Find("comment")
}

func (c *Changeset) CreatedBy() string {
	return c.Tags.Find("created_by")
}

func (c *Changeset) Locale() string {
	return c.Tags.Find("locale")
}

func (c *Changeset) Host() string {
	return c.Tags.Find("host")
}

func (c *Changeset) ImageryUsed() string {
	return c.Tags.Find("imagery_used")
}

func (c *Changeset) Source() string {
	return c.Tags.Find("source")
}

func (c *Changeset) Bot() bool {
	return c.Tags.Find("bot") == "yes"
}

// ChangesetDiscussion is a conversation about a changeset.
type ChangesetDiscussion struct {
	xml.Name `xml:"discussion"`
	Comments []*ChangesetComment `xml:"comment"`
}

// ChangesetComment is a specific comment in a changeset discussion.
type ChangesetComment struct {
	xml.Name  `xml:"comment"`
	User      string    `xml:"user,attr"`
	UserID    int       `xml:"uid,attr"`
	CreatedAt time.Time `xml:"date,attr"`
	Text      string    `xml:"text"`
}
