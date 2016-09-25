package osm

import (
	"encoding/xml"
	"sort"
	"time"
)

// WayID is the primary key of a way.
// A way is uniquely identifiable by the id + version.
type WayID int64

// Way is an osm way, ie collection of nodes.
type Way struct {
	XMLName     xml.Name    `xml:"way"`
	ID          WayID       `xml:"id,attr"`
	User        string      `xml:"user,attr"`
	UserID      UserID      `xml:"uid,attr"`
	Visible     bool        `xml:"visible,attr"`
	Version     int         `xml:"version,attr"`
	ChangesetID ChangesetID `xml:"changeset,attr"`
	Timestamp   time.Time   `xml:"timestamp,attr"`
	Nodes       []WayNode   `xml:"nd"`
	Tags        Tags        `xml:"tag"`

	// Minors are diffs from the original version representing
	// node updates independent of way version updates.
	Minors []MinorWay `xml:"minor-way,omitempty"`
}

// WayNode is a short node used as part of ways and relations in the osm xml.
type WayNode struct {
	ID NodeID `xml:"ref,attr"`

	// These attributes are populated for concrete versions of ways.
	Version     int         `xml:"version,attr,omitempty"`
	ChangesetID ChangesetID `xml:"changeset,attr,omitempty"`
	Lat         float64     `xml:"lat,attr,omitempty"`
	Lon         float64     `xml:"lon,attr,omitempty"`
}

// A MinorWay contains diff information for a minor version update of a
// way caused by nodes being updated independent of the way.
type MinorWay struct {
	Timestamp  time.Time      `xml:"timestamp,attr"`
	MinorNodes []MinorWayNode `xml:"minor-nd,omitempty"`
}

// A MinorWayNode is a reference to a updated node in a minor way version.
type MinorWayNode struct {
	Index       int         `xml:"index,attr"`
	Version     int         `xml:"version,attr,omitempty"`
	ChangesetID ChangesetID `xml:"changeset,attr,omitempty"`
	Lat         float64     `xml:"lat,attr,omitempty"`
	Lon         float64     `xml:"lon,attr,omitempty"`
}

// Ways is a set of osm ways with some helper functions attached.
type Ways []*Way

// Marshal encodes the ways using protocol buffers.
func (ws Ways) Marshal() ([]byte, error) {
	o := OSM{
		Ways: ws,
	}

	return o.Marshal()
}

// UnmarshalWays will unmarshal the data into a list of ways.
func UnmarshalWays(data []byte) (Ways, error) {
	o, err := UnmarshalOSM(data)
	if err != nil {
		return nil, err
	}

	return o.Ways, nil
}

type waysSort Ways

// SortByIDVersion will sort the set of ways first by id and then version
// in ascending order.
func (ws Ways) SortByIDVersion() {
	sort.Sort(waysSort(ws))
}
func (ws waysSort) Len() int      { return len(ws) }
func (ws waysSort) Swap(i, j int) { ws[i], ws[j] = ws[j], ws[i] }
func (ws waysSort) Less(i, j int) bool {
	if ws[i].ID == ws[j].ID {
		return ws[i].Version < ws[j].Version
	}

	return ws[i].ID < ws[j].ID
}
