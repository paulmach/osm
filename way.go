package osm

import (
	"sort"
	"time"
)

// WayID is the primary key of a way.
// A way is uniquely identifiable by the id + version.
type WayID int

// Way is an osm way, ie collection of nodes.
type Way struct {
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

type waysSort Ways

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
