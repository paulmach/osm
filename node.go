package osm

import (
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/paulmach/go.osm/internal/osmpb"
)

// NodeID corresponds the primary key of a node.
// The node id + version uniquely identify a node.
type NodeID int64

// ElementID is a helper returning the element id for this node id.
// Version is left at 0.
func (id NodeID) ElementID() ElementID {
	return ElementID{
		Type: NodeType,
		Ref:  int64(id),
	}
}

// Node is an osm point and allows for marshalling to/from osm xml.
type Node struct {
	XMLName     xmlNameJSONTypeNode `xml:"node" json:"type"`
	ID          NodeID              `xml:"id,attr" json:"id"`
	Lat         float64             `xml:"lat,attr" json:"lat"`
	Lon         float64             `xml:"lon,attr" json:"lon"`
	User        string              `xml:"user,attr" json:"user,omitempty"`
	UserID      UserID              `xml:"uid,attr" json:"uid,omitempty"`
	Visible     bool                `xml:"visible,attr" json:"visible"`
	Version     int                 `xml:"version,attr" json:"version,omitempty"`
	ChangesetID ChangesetID         `xml:"changeset,attr" json:"changeset,omitempty"`
	Timestamp   time.Time           `xml:"timestamp,attr" json:"timestamp"`
	Tags        Tags                `xml:"tag" json:"tags,omitempty"`

	// Committed, is the estimated time this object was committed
	// and made visible in the central OSM database.
	Committed *time.Time `xml:"committed,attr,omitempty" json:"committed,omitempty"`
}

// ElementID returns the element id of the node.
func (n *Node) ElementID() ElementID {
	return ElementID{
		Type:    NodeType,
		Ref:     int64(n.ID),
		Version: n.Version,
	}
}

// CommittedAt returns the best estimate on when this element
// became was written/committed into the database.
func (n *Node) CommittedAt() time.Time {
	if n.Committed != nil {
		return *n.Committed
	}

	return n.Timestamp
}

// Nodes is a list of nodes with helper functions on top.
type Nodes []*Node

// Marshal encodes the nodes using protocol buffers.
func (ns Nodes) Marshal() ([]byte, error) {
	if len(ns) == 0 {
		return nil, nil
	}

	ss := &stringSet{}
	encoded := marshalNodes(ns, ss, true)
	encoded.Strings = ss.Strings()

	return proto.Marshal(encoded)
}

// UnmarshalNodes will unmarshal the data into a list of nodes.
func UnmarshalNodes(data []byte) (Nodes, error) {
	if len(data) == 0 {
		return nil, nil
	}

	pbf := &osmpb.DenseNodes{}
	err := proto.Unmarshal(data, pbf)
	if err != nil {
		return nil, err
	}

	return unmarshalNodes(pbf, pbf.GetStrings(), nil)
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
