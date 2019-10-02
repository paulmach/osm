package osm

import (
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/paulmach/orb"
)

// WayID is the primary key of a way.
// A way is uniquely identifiable by the id + version.
type WayID int64

// ObjectID is a helper returning the object id for this way id.
func (id WayID) ObjectID(v int) ObjectID {
	return ObjectID(id.ElementID(v))
}

// FeatureID is a helper returning the feature id for this way id.
func (id WayID) FeatureID() FeatureID {
	return FeatureID(wayMask | (id << versionBits))
}

// ElementID is a helper to convert the id to an element id.
func (id WayID) ElementID(v int) ElementID {
	return id.FeatureID().ElementID(v)
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
	Committed *time.Time `xml:"committed,attr,omitempty" json:"committed,omitempty"`

	// Updates are changes to the nodes of this way independent
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

// ObjectID returns the object id of the way.
func (w *Way) ObjectID() ObjectID {
	return w.ID.ObjectID(w.Version)
}

// FeatureID returns the feature id of the way.
func (w *Way) FeatureID() FeatureID {
	return w.ID.FeatureID()
}

// ElementID returns the element id of the way.
func (w *Way) ElementID() ElementID {
	return w.ID.ElementID(w.Version)
}

// FeatureID returns the feature id of the way node.
func (wn WayNode) FeatureID() FeatureID {
	return wn.ID.FeatureID()
}

// ElementID returns the element id of the way node.
func (wn WayNode) ElementID() ElementID {
	return wn.ID.ElementID(wn.Version)
}

// Point returns the orb.Point location for the way node.
// Will be (0, 0) if the way is not annotated.
func (wn WayNode) Point() orb.Point {
	return orb.Point{wn.Lon, wn.Lat}
}

// CommittedAt returns the best estimate on when this element
// became was written/committed into the database.
func (w *Way) CommittedAt() time.Time {
	if w.Committed != nil {
		return *w.Committed
	}

	return w.Timestamp
}

// TagMap returns the element tags as a key/value map.
func (w *Way) TagMap() map[string]string {
	return w.Tags.Map()
}

// ApplyUpdatesUpTo will apply the updates to this object upto and including
// the given time.
func (w *Way) ApplyUpdatesUpTo(t time.Time) error {
	var notApplied []Update
	for _, u := range w.Updates {
		if u.Timestamp.After(t) {
			notApplied = append(notApplied, u)
			continue
		}

		if err := w.applyUpdate(u); err != nil {
			return err
		}
	}

	w.Updates = notApplied
	return nil
}

// applyUpdate will modify the current way and dictated by the given update.
// Will return UpdateIndexOutOfRangeError if the update index is too large.
func (w *Way) applyUpdate(u Update) error {
	if u.Index >= len(w.Nodes) {
		return &UpdateIndexOutOfRangeError{Index: u.Index}
	}

	w.Nodes[u.Index].Version = u.Version
	w.Nodes[u.Index].ChangesetID = u.ChangesetID
	w.Nodes[u.Index].Lat = u.Lat
	w.Nodes[u.Index].Lon = u.Lon

	return nil
}

// LineString will convert the annotated nodes into a LineString datatype.
func (w *Way) LineString() orb.LineString {
	ls := make(orb.LineString, 0, len(w.Nodes))
	for _, n := range w.Nodes {
		if n.Version != 0 || n.Lon != 0 || n.Lat != 0 {
			// If version is there we assume it's annotated.
			// We use this assumption in a lot of places.
			ls = append(ls, n.Point())
		}
	}

	return ls
}

// LineStringAt will return the LineString from the annotated points at
// the given time. It will apply to the updates upto and including the give time.
func (w *Way) LineStringAt(t time.Time) orb.LineString {
	// linestring with all the zeros
	ls := make(orb.LineString, 0, len(w.Nodes))
	for _, n := range w.Nodes {
		ls = append(ls, n.Point())
	}

	for _, u := range w.Updates {
		if u.Timestamp.After(t) {
			break
		}

		if u.Index >= len(ls) {
			continue
		}

		ls[u.Index][0] = u.Lon
		ls[u.Index][1] = u.Lat
	}

	// remove all the zeros
	count := 0
	for i := range ls {
		if n := w.Nodes[i]; n.Version == 0 && n.Lon == 0 && n.Lat == 0 {
			continue
		}

		ls[count] = ls[i]
		count++
	}

	return ls[:count]
}

// Bounds computes the bounds for the given way nodes.
func (wn WayNodes) Bounds() *Bounds {
	b := &Bounds{
		MinLat: math.MaxFloat64,
		MaxLat: -math.MaxFloat64,
		MinLon: math.MaxFloat64,
		MaxLon: -math.MaxFloat64,
	}

	for _, n := range wn {
		b.MinLat = math.Min(b.MinLat, n.Lat)
		b.MaxLat = math.Max(b.MaxLat, n.Lat)

		b.MinLon = math.Min(b.MinLon, n.Lon)
		b.MaxLon = math.Max(b.MaxLon, n.Lon)
	}

	return b
}

// Bound computes the orb.Bound for the given way nodes.
func (wn WayNodes) Bound() orb.Bound {
	b := orb.Bound{
		Min: orb.Point{math.MaxFloat64, math.MaxFloat64},
		Max: orb.Point{-math.MaxFloat64, -math.MaxFloat64},
	}

	for _, n := range wn {
		b.Min[0] = math.Min(b.Min[0], n.Lon)
		b.Max[0] = math.Max(b.Max[0], n.Lon)

		b.Min[1] = math.Min(b.Min[1], n.Lat)
		b.Max[1] = math.Max(b.Max[1], n.Lat)
	}

	return b
}

// ElementIDs returns a list of element ids for the way nodes.
func (wn WayNodes) ElementIDs() ElementIDs {
	// add 1 to the memory length because a common use case
	// is to append the way.
	ids := make(ElementIDs, len(wn), len(wn)+1)
	for i, n := range wn {
		ids[i] = n.ElementID()
	}

	return ids
}

// FeatureIDs returns a list of feature ids for the way nodes.
func (wn WayNodes) FeatureIDs() FeatureIDs {
	// add 1 to the memory length because a common use case
	// is to append the way.
	ids := make(FeatureIDs, len(wn), len(wn)+1)
	for i, n := range wn {
		ids[i] = n.FeatureID()
	}

	return ids
}

// NodeIDs returns a list of node ids for the way nodes.
func (wn WayNodes) NodeIDs() []NodeID {
	ids := make([]NodeID, len(wn))
	for i, n := range wn {
		ids[i] = n.ID
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

// IDs returns the ids for all the ways.
func (ws Ways) IDs() []WayID {
	result := make([]WayID, len(ws))
	for i, w := range ws {
		result[i] = w.ID
	}

	return result
}

// FeatureIDs returns the feature ids for all the ways.
func (ws Ways) FeatureIDs() FeatureIDs {
	r := make(FeatureIDs, len(ws))
	for i, w := range ws {
		r[i] = w.FeatureID()
	}

	return r
}

// ElementIDs returns the element ids for all the ways.
func (ws Ways) ElementIDs() ElementIDs {
	r := make(ElementIDs, len(ws))
	for i, w := range ws {
		r[i] = w.ElementID()
	}

	return r
}

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
