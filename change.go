package osm

import "encoding/xml"

// Change is the structure of a changeset to be
// uploaded or downloaded from the server.
// See: http://wiki.openstreetmap.org/wiki/OsmChange
type Change struct {
	XMLName  xml.Name `xml:"osmChange"`
	Creates  []OSM    `xml:"create"`
	Modifies []OSM    `xml:"modify"`
	Deletes  []OSM    `xml:"delete"`
}
