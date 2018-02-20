package osm

import (
	"time"
)

// UserID is the primary key for a user.
// This is unique the display name may change.
type UserID int64

// ObjectID is a helper returning the object id for this user id.
func (id UserID) ObjectID() ObjectID {
	return ObjectID(userMask | (id << versionBits))
}

// Users is a collection of users with some helpers attached.
type Users []*User

// A User is a registered OSM user.
type User struct {
	XMLName     xmlNameJSONTypeUser `xml:"user" json:"type"`
	ID          UserID              `xml:"id,attr"`
	Name        string              `xml:"display_name,attr"`
	Description string              `xml:"description"`
	Img         struct {
		Href string `xml:"href,attr"`
	} `xml:"img"`
	Changesets struct {
		Count int `xml:"count,attr"`
	} `xml:"changesets"`
	Traces struct {
		Count int `xml:"count,attr"`
	} `xml:"traces"`
	Home struct {
		Lat  float64 `xml:"lat,attr"`
		Lon  float64 `xml:"lon,attr"`
		Zoom int     `xml:"zoom,attr"`
	} `xml:"home"`
	Languages []string `xml:"languages>lang"`
	Blocks    struct {
		Received struct {
			Count  int `xml:"count,attr"`
			Active int `xml:"active,attr"`
		} `xml:"received"`
	} `xml:"blocks"`
	Messages struct {
		Received struct {
			Count  int `xml:"count,attr"`
			Unread int `xml:"unread,attr"`
		} `xml:"received"`
		Sent struct {
			Count int `xml:"count,attr"`
		} `xml:"sent"`
	} `xml:"messages"`
	CreatedAt time.Time `xml:"account_created,attr"`
}

// ObjectID returns the object id of the user.
func (u *User) ObjectID() ObjectID {
	return u.ID.ObjectID()
}
