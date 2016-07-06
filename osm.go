package osm

import (
	"encoding/xml"

	"github.com/paulmach/orb/geo"
)

// OSM represents the core osm data.
// I designed to parse http://wiki.openstreetmap.org/wiki/OSM_XML
type OSM struct {
	Bound     *Bounds   `xml:"bounds"`
	Nodes     Nodes     `xml:"node"`
	Ways      Ways      `xml:"way"`
	Relations Relations `xml:"relation"`

	// Changesets will normally only be populated when returning
	// a list of changesets.
	Changesets Changesets `xml:"changeset"`
}

// Bounds are the bounds of osm data as defined in the xml file.
type Bounds struct {
	XMLName xml.Name `xml:"bounds"`
	MinLat  float64  `xml:"minlat,attr"`
	MaxLat  float64  `xml:"maxlat,attr"`
	MinLng  float64  `xml:"minlon,attr"`
	MaxLng  float64  `xml:"maxlon,attr"`
}

// Bound returns a geo bound object from the struct/xml definition.
func (b *Bounds) Bound() geo.Bound {
	return geo.NewBound(b.MinLng, b.MaxLng, b.MinLat, b.MaxLat)
}
