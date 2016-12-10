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

	// Committed, is the estimated time this object was committed
	// and made visible in the central OSM database.
	Committed *time.Time `xml:"commited,attr,omitempty"`

	// Updates are changes the nodes of this way independent
	// of an update to the way itself. The OSM api allows a child
	// to be updated without any changes to the parent.
	Updates Updates `xml:"update,omitempty"`
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

// ElementID returns the element id of the way.
func (w *Way) ElementID() ElementID {
	return ElementID{
		Type:    WayType,
		ID:      int64(w.ID),
		Version: w.Version,
	}
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
