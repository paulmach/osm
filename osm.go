package osm

import (
	"encoding/xml"
	"time"
)

type OSM struct {
	Bounds    *Bounds    `xml:"bounds"`
	Nodes     []Node     `xml:"node"`
	Ways      []Way      `xml:"way"`
	Relations []Relation `xml:"relation"`
}

type Bounds struct {
	XMLName xml.Name `xml:"bounds"`
	MinLat  float64  `xml:"minlat,attr"`
	MaxLat  float64  `xml:"maxlat,attr"`
	MinLng  float64  `xml:"minlon,attr"`
	MaxLng  float64  `xml:"maxlon,attr"`
}

type Node struct {
	XMLName    xml.Name  `xml:"node"`
	ID         int       `xml:"id,attr"`
	Lat        float64   `xml:"lat,attr"`
	Lng        float64   `xml:"lon,attr"`
	User       string    `xml:"user,attr"`
	UserID     int       `xml:"uid,attr"`
	Visible    bool      `xml:"visible,attr"`
	Version    int       `xml:"version,attr"`
	ChangsetID int       `xml:"changeset,attr"`
	Timestamp  time.Time `xml:"timestamp,attr"`
	Tags       Tags      `xml:"tag"`
}

type NodeRef struct {
	XMLName xml.Name `xml:"nd"`
	Ref     int      `xml:"ref,attr"`
}

type Way struct {
	ID         int       `xml:"id,attr"`
	User       string    `xml:"user,attr"`
	UserID     int       `xml:"uid,attr"`
	Visible    bool      `xml:"visible,attr"`
	Version    int       `xml:"version,attr"`
	ChangsetID int       `xml:"changeset,attr"`
	Timestamp  time.Time `xml:"timestamp,attr"`
	NodeRefs   []NodeRef `xml:"nd"`
	Tags       Tags      `xml:"tag"`
}

type Relation struct {
	ID         int       `xml:"id,attr"`
	User       string    `xml:"user,attr"`
	UserID     int       `xml:"uid,attr"`
	Visible    bool      `xml:"visible,attr"`
	Version    int       `xml:"version,attr"`
	ChangsetID int       `xml:"changeset,attr"`
	Timestamp  time.Time `xml:"timestamp,attr"`

	Members []Member `xml:"member"`
}

type Member struct {
	XMLName xml.Name `xml:"member"`
	Type    string   `xml:"type,attr"`
	Ref     int      `xml:"ref,attr"`
	Role    int      `xml:"role,attr"`
}
