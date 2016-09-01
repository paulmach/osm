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
	NodeRefs    []NodeRef   `xml:"nd"`
	Tags        Tags        `xml:"tag"`
}

// NodeRef is a short node used as part of ways and relations in the osm xml.
type NodeRef struct {
	Ref NodeID `xml:"ref,attr"`
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
