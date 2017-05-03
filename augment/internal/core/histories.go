package core

import osm "github.com/paulmach/go.osm"

// Histories is a container for element history data.
// It simplifies access by abstracting the clearing of
// the version number from the element id.
type Histories struct {
	data map[osm.FeatureID]ChildList
}

// Get returns the history in ChildList form.
func (h *Histories) Get(id osm.FeatureID) ChildList {
	if h.data == nil {
		return nil
	}

	return h.data[id]
}

// Set sets the element history into the map.
// The element is deleted if list is nil.
func (h *Histories) Set(id osm.FeatureID, list ChildList) {
	if h.data == nil {
		h.data = make(map[osm.FeatureID]ChildList)
	}

	if list == nil {
		delete(h.data, id)
	}

	h.data[id] = list
}
