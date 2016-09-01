package osm

import (
	"encoding/xml"
	"sort"
	"time"
)

// NodeID corresponds the primary key of a node.
// The node id + version uniquely identify a node.
type NodeID int64

// Node is an osm point and allows for marshalling to/from osm xml.
type Node struct {
	XMLName     xml.Name    `xml:"node"`
	ID          NodeID      `xml:"id,attr"`
	Lat         float64     `xml:"lat,attr"`
	Lon         float64     `xml:"lon,attr"`
	User        string      `xml:"user,attr"`
	UserID      UserID      `xml:"uid,attr"`
	Visible     bool        `xml:"visible,attr"`
	Version     int         `xml:"version,attr"`
	ChangesetID ChangesetID `xml:"changeset,attr"`
	Timestamp   time.Time   `xml:"timestamp,attr"`
	Tags        Tags        `xml:"tag"`
}

// Nodes is a set of nodes with helper functions on top.
type Nodes []*Node

// Marshal encodes the nodes using protocol buffers.
func (ns Nodes) Marshal() ([]byte, error) {
	o := OSM{
		Nodes: ns,
	}

	return o.Marshal()
}

// UnmarshalNodes will unmarshal the data into a list of nodes.
func UnmarshalNodes(data []byte) (Nodes, error) {
	o, err := UnmarshalOSM(data)
	if err != nil {
		return nil, err
	}

	return o.Nodes, nil
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

type nodesSort Nodes

// SortByIDVersion will sort the set of nodes first by id and then version
// in ascending order.
func (ns Nodes) SortByIDVersion() {
	sort.Sort(nodesSort(ns))
}

func (ns nodesSort) Len() int      { return len(ns) }
func (ns nodesSort) Swap(i, j int) { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodesSort) Less(i, j int) bool {
	if ns[i].ID == ns[j].ID {
		return ns[i].Version < ns[j].Version
	}

	return ns[i].ID < ns[j].ID
}
