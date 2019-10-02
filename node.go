package osm

import (
	"sort"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/osm/internal/osmpb"

	"github.com/gogo/protobuf/proto"
)

// NodeID corresponds the primary key of a node.
// The node id + version uniquely identify a node.
type NodeID int64

// ObjectID is a helper returning the object id for this node id.
func (id NodeID) ObjectID(v int) ObjectID {
	return ObjectID(id.ElementID(v))
}

// FeatureID is a helper returning the feature id for this node id.
func (id NodeID) FeatureID() FeatureID {
	return FeatureID(nodeMask | (id << versionBits))
}

// ElementID is a helper to convert the id to an element id.
func (id NodeID) ElementID(v int) ElementID {
	return id.FeatureID().ElementID(v)
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

// ObjectID returns the object id of the node.
func (n *Node) ObjectID() ObjectID {
	return n.ID.ObjectID(n.Version)
}

// FeatureID returns the feature id of the node.
func (n *Node) FeatureID() FeatureID {
	return n.ID.FeatureID()
}

// ElementID returns the element id of the node.
func (n *Node) ElementID() ElementID {
	return n.ID.ElementID(n.Version)
}

// CommittedAt returns the best estimate on when this element
// became was written/committed into the database.
func (n *Node) CommittedAt() time.Time {
	if n.Committed != nil {
		return *n.Committed
	}

	return n.Timestamp
}

// TagMap returns the element tags as a key/value map.
func (n *Node) TagMap() map[string]string {
	return n.Tags.Map()
}

// Point returns the orb.Point location for the node.
// Will be (0, 0) for "deleted" nodes.
func (n *Node) Point() orb.Point {
	return orb.Point{n.Lon, n.Lat}
}

// Nodes is a list of nodes with helper functions on top.
type Nodes []*Node

// IDs returns the ids for all the ways.
func (ns Nodes) IDs() []NodeID {
	result := make([]NodeID, len(ns))
	for i, n := range ns {
		result[i] = n.ID
	}

	return result
}

// FeatureIDs returns the feature ids for all the nodes.
func (ns Nodes) FeatureIDs() FeatureIDs {
	r := make(FeatureIDs, len(ns))
	for i, n := range ns {
		r[i] = n.FeatureID()
	}

	return r
}

// ElementIDs returns the element ids for all the nodes.
func (ns Nodes) ElementIDs() ElementIDs {
	r := make(ElementIDs, len(ns))
	for i, n := range ns {
		r[i] = n.ElementID()
	}

	return r
}

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
