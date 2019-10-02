package osm

import (
	"time"
)

// UserID is the primary key for a user.
// This is unique as the display name may change.
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
	ID          UserID              `xml:"id,attr" json:"id"`
	Name        string              `xml:"display_name,attr" json:"name"`
	Description string              `xml:"description" json:"description,omitempty"`
	Img         struct {
		Href string `xml:"href,attr" json:"href"`
	} `xml:"img" json:"img"`
	Changesets struct {
		Count int `xml:"count,attr" json:"count"`
	} `xml:"changesets" json:"changesets"`
	Traces struct {
		Count int `xml:"count,attr" json:"count"`
	} `xml:"traces" json:"traces"`
	Home struct {
		Lat  float64 `xml:"lat,attr" json:"lat"`
		Lon  float64 `xml:"lon,attr" json:"lon"`
		Zoom int     `xml:"zoom,attr" json:"zoom"`
	} `xml:"home" json:"home"`
	Languages []string `xml:"languages>lang" json:"languages"`
	Blocks    struct {
		Received struct {
			Count  int `xml:"count,attr" json:"count"`
			Active int `xml:"active,attr" json:"active"`
		} `xml:"received" json:"received"`
	} `xml:"blocks" json:"blocks"`
	Messages struct {
		Received struct {
			Count  int `xml:"count,attr" json:"count"`
			Unread int `xml:"unread,attr" json:"unread"`
		} `xml:"received" json:"received"`
		Sent struct {
			Count int `xml:"count,attr" json:"count"`
		} `xml:"sent" json:"sent"`
	} `xml:"messages" json:"messages"`
	CreatedAt time.Time `xml:"account_created,attr" json:"created_at"`
}

// ObjectID returns the object id of the user.
func (u *User) ObjectID() ObjectID {
	return u.ID.ObjectID()
}
