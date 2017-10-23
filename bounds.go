package osm

import (
	"errors"

	"github.com/paulmach/orb/tile"
)

// Bounds are the bounds of osm data as defined in the xml file.
type Bounds struct {
	MinLat float64 `xml:"minlat,attr"`
	MaxLat float64 `xml:"maxlat,attr"`
	MinLon float64 `xml:"minlon,attr"`
	MaxLon float64 `xml:"maxlon,attr"`
}

// NewBoundsFromTile creates a bound given an online map tile index.
func NewBoundsFromTile(t tile.Tile) (*Bounds, error) {
	maxIndex := uint32(1 << t.Z)
	if t.X >= maxIndex {
		return nil, errors.New("osm: x index out of range for this zoom")
	}
	if t.Y >= maxIndex {
		return nil, errors.New("osm: y index out of range for this zoom")
	}

	b := t.Bound()
	return &Bounds{
		MinLat: b[0].Lat(),
		MaxLat: b[1].Lat(),
		MinLon: b[0].Lon(),
		MaxLon: b[1].Lon(),
	}, nil
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
