package osm

import (
	"errors"
	"math"
)

// Bounds are the bounds of osm data as defined in the xml file.
type Bounds struct {
	MinLat float64 `xml:"minlat,attr"`
	MaxLat float64 `xml:"maxlat,attr"`
	MinLon float64 `xml:"minlon,attr"`
	MaxLon float64 `xml:"maxlon,attr"`
}

// NewBoundFromTile creates a bound given an online map tile index.
func NewBoundFromTile(x, y, z uint64) (*Bounds, error) {
	maxIndex := uint64(1) << z
	if x >= maxIndex {
		return nil, errors.New("osm: x index out of range for this zoom")
	}
	if y >= maxIndex {
		return nil, errors.New("osm: y index out of range for this zoom")
	}

	shift := 31 - z
	if z > 31 {
		shift = 0
	}

	lon1, lat1 := scalarInverse(x<<shift, y<<shift, 31)
	lon2, lat2 := scalarInverse((x+1)<<shift, (y+1)<<shift, 31)

	return &Bounds{
		MinLat: lat2,
		MaxLat: lat1,
		MinLon: lon1,
		MaxLon: lon2,
	}, nil
}

func scalarInverse(x, y, level uint64) (lng, lat float64) {
	var factor uint64

	factor = 1 << level
	maxtiles := float64(factor)

	lng = 360.0 * (float64(x)/maxtiles - 0.5)
	lat = (2.0*math.Atan(math.Exp(math.Pi-(2*math.Pi)*(float64(y))/maxtiles)))*(180.0/math.Pi) - 90.0

	return lng, lat
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
