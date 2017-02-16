package osm

// Bounds are the bounds of osm data as defined in the xml file.
type Bounds struct {
	MinLat float64 `xml:"minlat,attr"`
	MaxLat float64 `xml:"maxlat,attr"`
	MinLon float64 `xml:"minlon,attr"`
	MaxLon float64 `xml:"maxlon,attr"`
}

// ContainsNode returns true if the node is within the bound.
// Uses inclusive intervals, ie. returns true if on the boundary.
func (b *Bounds) ContainsNode(n *Node) bool {
	if n.Lat < b.MinLat || n.Lat > b.MaxLat {
		return false
	}

	if n.Lon < b.MinLon || n.Lon > b.MaxLon {
		return false
	}

	return true
}
