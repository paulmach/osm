package osm

import (
	"encoding/xml"
	"time"
)

type Nodes []*Node

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

// ActiveAt returns the active node at the give time for a history of nodes.
// These nodes should normally be returned from a /nodes/:id/history api call.
func (ns Nodes) ActiveAt(t time.Time) *Node {
	if len(ns) == 0 {
		return nil
	}

	var active *Node
	for _, n := range ns {
		if n.Timestamp.After(t) {
			return active
		}

		active = n
	}

	return active
}
