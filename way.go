package osm

import (
	"encoding/json"
	"sort"
	"time"
)

// WayID is the primary key of a way.
// A way is uniquely identifiable by the id + version.
type WayID int64

// ElementID is a helper returning the element id for this node id.
// Version is left at 0.
func (id WayID) ElementID() ElementID {
	return ElementID{
		Type: WayType,
		Ref:  int64(id),
	}
}

// Way is an osm way, ie collection of nodes.
type Way struct {
	XMLName     xmlNameJSONTypeWay `xml:"way" json:"type"`
	ID          WayID              `xml:"id,attr" json:"id"`
	User        string             `xml:"user,attr" json:"user,omitempty"`
	UserID      UserID             `xml:"uid,attr" json:"uid,omitempty"`
	Visible     bool               `xml:"visible,attr" json:"visible"`
	Version     int                `xml:"version,attr" json:"version,omitempty"`
	ChangesetID ChangesetID        `xml:"changeset,attr" json:"changeset,omitempty"`
	Timestamp   time.Time          `xml:"timestamp,attr" json:"timestamp"`
	Nodes       WayNodes           `xml:"nd" json:"nodes"`
	Tags        Tags               `xml:"tag" json:"tags,omitempty"`

	// Committed, is the estimated time this object was committed
	// and made visible in the central OSM database.
	Committed *time.Time `xml:"commited,attr,omitempty" json:"committed,omitempty"`

	// Updates are changes the nodes of this way independent
	// of an update to the way itself. The OSM api allows a child
	// to be updated without any changes to the parent.
	Updates Updates `xml:"update,omitempty" json:"updates,omitempty"`

	// Bounds are included by overpass, and maybe others
	Bounds *Bounds `xml:"bounds,omitempty" json:"bounds,omitempty"`
}

// WayNodes represents a collection of way nodes.
type WayNodes []WayNode

// WayNode is a short node used as part of ways and relations in the osm xml.
type WayNode struct {
	ID NodeID `xml:"ref,attr"`

	// These attributes are populated for concrete versions of ways.
	Version     int         `xml:"version,attr,omitempty"`
	ChangesetID ChangesetID `xml:"changeset,attr,omitempty"`
	Lat         float64     `xml:"lat,attr,omitempty"`
	Lon         float64     `xml:"lon,attr,omitempty"`
}

// ElementID returns the element id of the way.
func (w *Way) ElementID() ElementID {
	return ElementID{
		Type:    WayType,
		Ref:     int64(w.ID),
		Version: w.Version,
	}
}

// CommittedAt returns the best estimate on when this element
// became was written/committed into the database.
func (w *Way) CommittedAt() time.Time {
	if w.Committed != nil {
		return *w.Committed
	}

	return w.Timestamp
}

// ApplyUpdatesUpTo will apply the updates to this object upto and including
// the given time.
func (w *Way) ApplyUpdatesUpTo(t time.Time) error {
	for _, u := range w.Updates {
		if u.Timestamp.After(t) {
			continue
		}

		if err := w.ApplyUpdate(u); err != nil {
			return err
		}
	}

	return nil
}

// ApplyUpdate will modify the current way and dictated by the given update.
// Will return UpdateIndexOutOfRangeError if the update index is too large.
func (w *Way) ApplyUpdate(u Update) error {
	if u.Index >= len(w.Nodes) {
		return &UpdateIndexOutOfRangeError{Index: u.Index}
	}

	w.Nodes[u.Index].Version = u.Version
	w.Nodes[u.Index].ChangesetID = u.ChangesetID
	w.Nodes[u.Index].Lat = u.Lat
	w.Nodes[u.Index].Lon = u.Lon

	return nil
}

// ElementIDs returns a list of element ids for the way nodes.
func (wn WayNodes) ElementIDs() ElementIDs {
	ids := make(ElementIDs, len(wn)+1)
	for i, n := range wn {
		ids[i] = ElementID{
			Type:    NodeType,
			Ref:     int64(n.ID),
			Version: n.Version,
		}
	}

	return ids
}

// MarshalJSON allows the waynodes to be marshalled as an array of ids,
// as defined by the overpass osmjson.
func (wn WayNodes) MarshalJSON() ([]byte, error) {
	a := make([]int64, 0, len(wn))
	for _, n := range wn {
		a = append(a, int64(n.ID))
	}

	return json.Marshal(a)
}

// UnmarshalJSON allows the tags to be unmarshalled from an array of ids,
// as defined by the overpass osmjson.
func (wn *WayNodes) UnmarshalJSON(data []byte) error {
	var a []int64
	err := json.Unmarshal(data, &a)
	if err != nil {
		return err
	}

	nodes := make(WayNodes, len(a))
	for i, id := range a {
		nodes[i].ID = NodeID(id)
	}

	*wn = nodes
	return nil
}

// Ways is a list of osm ways with some helper functions attached.
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
