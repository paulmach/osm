package osm

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/paulmach/go.osm/osmpb"
	"github.com/paulmach/orb/geo"
)

// OSM represents the core osm data.
// I designed to parse http://wiki.openstreetmap.org/wiki/OSM_XML
type OSM struct {
	Bound     *Bounds   `xml:"bounds"`
	Nodes     Nodes     `xml:"node"`
	Ways      Ways      `xml:"way"`
	Relations Relations `xml:"relation"`
}

// OSMChangesets decodes a list of changesets in osm xml.
type OSMChangesets struct {
	Changesets Changesets `xml:"changeset"`
}

// Bounds are the bounds of osm data as defined in the xml file.
type Bounds struct {
	MinLat float64 `xml:"minlat,attr"`
	MaxLat float64 `xml:"maxlat,attr"`
	MinLng float64 `xml:"minlon,attr"`
	MaxLng float64 `xml:"maxlon,attr"`
}

// Bound returns a geo bound object from the struct/xml definition.
func (b *Bounds) Bound() geo.Bound {
	return geo.NewBound(b.MinLng, b.MaxLng, b.MinLat, b.MaxLat)
}

// Marshal encodes the osm data using protocol buffers.
func (o *OSM) Marshal() ([]byte, error) {
	ss := &stringSet{}
	encoded := marshalOSM(o, ss, true)
	encoded.Strings = ss.Strings()

	return proto.Marshal(encoded)
}

// includeChangeset can be set to false to not repeat the changeset
// info for every item, if this comes from osm change data.
func marshalOSM(o *OSM, ss *stringSet, includeChangeset bool) *osmpb.OSM {
	encoded := &osmpb.OSM{}
	if o == nil {
		return nil
	}

	if len(o.Nodes) > 0 {
		encoded.DenseNodes = marshalNodes(o.Nodes, ss, includeChangeset)
	}

	if len(o.Ways) > 0 {
		encoded.Ways = make([]*osmpb.Way, len(o.Ways), len(o.Ways))
		for i, w := range o.Ways {
			encoded.Ways[i] = marshalWay(w, ss, includeChangeset)
		}
	}

	if len(o.Relations) > 0 {
		encoded.Relations = make([]*osmpb.Relation, len(o.Relations), len(o.Relations))
		for i, r := range o.Relations {
			encoded.Relations[i] = marshalRelation(r, ss, includeChangeset)
		}
	}

	if o.Bound != nil {
		encoded.Bounds = &osmpb.Bounds{
			MinLat: proto.Int64(geoToInt64(o.Bound.MinLat)),
			MaxLat: proto.Int64(geoToInt64(o.Bound.MaxLat)),
			MinLng: proto.Int64(geoToInt64(o.Bound.MinLng)),
			MaxLng: proto.Int64(geoToInt64(o.Bound.MaxLng)),
		}
	}

	return encoded
}

func unmarshalOSM(encoded *osmpb.OSM, ss []string) (*OSM, error) {
	if encoded == nil {
		return nil, nil
	}

	o := &OSM{}
	if len(encoded.Nodes) != 0 && encoded.DenseNodes != nil {
		return nil, errors.New("found both nodes and dense nodes")
	}

	if len(encoded.Nodes) != 0 {
		o.Nodes = make([]*Node, len(encoded.Nodes), len(encoded.Nodes))
		for i, en := range encoded.Nodes {
			n, err := unmarshalNode(en, ss)
			if err != nil {
				return nil, err
			}

			o.Nodes[i] = n
		}
	}

	if encoded.DenseNodes != nil {
		var err error
		o.Nodes, err = unmarshalNodes(encoded.DenseNodes, ss)
		if err != nil {
			return nil, err
		}
	}

	if len(encoded.Ways) != 0 {
		o.Ways = make([]*Way, len(encoded.Ways), len(encoded.Ways))
		for i, ew := range encoded.Ways {
			w, err := unmarshalWay(ew, ss)
			if err != nil {
				return nil, err
			}

			o.Ways[i] = w
		}
	}

	if len(encoded.Relations) != 0 {
		o.Relations = make([]*Relation, len(encoded.Relations), len(encoded.Relations))
		for i, er := range encoded.Relations {
			r, err := unmarshalRelation(er, ss)
			if err != nil {
				return nil, err
			}

			o.Relations[i] = r
		}
	}

	if encoded.Bounds != nil {
		o.Bound = &Bounds{
			MinLat: float64(o.Bound.MinLat / locMultiple),
			MaxLat: float64(o.Bound.MinLat / locMultiple),
			MinLng: float64(o.Bound.MaxLng / locMultiple),
			MaxLng: float64(o.Bound.MaxLng / locMultiple),
		}
	}

	return o, nil
}
