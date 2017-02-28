package core

import osm "github.com/paulmach/go.osm"

// Histories is a container for element history data.
// It simplifies access by abstracting the clearing of
// the version number from the element id.
type Histories struct {
	data map[osm.ElementID]ChildList
}

// Get returns the history in ChildList form.
func (h *Histories) Get(id osm.ElementID) ChildList {
	if h.data == nil {
		return nil
	}

	return h.data[id.ClearVersion()]
}

// Set sets the element and history into the map.
// The element is deleted if list is nil.
func (h *Histories) Set(id osm.ElementID, list ChildList) {
	if h.data == nil {
		h.data = make(map[osm.ElementID]ChildList)
	}

	if list == nil {
		delete(h.data, id.ClearVersion())
	}

	h.data[id.ClearVersion()] = list
}
