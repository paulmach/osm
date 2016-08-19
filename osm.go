package osm

import (
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/paulmach/go.osm/internal/osmpb"
)

// OSM represents the core osm data.
// I designed to parse http://wiki.openstreetmap.org/wiki/OSM_XML
type OSM struct {
	Bound     *Bound    `xml:"bounds"`
	Nodes     Nodes     `xml:"node"`
	Ways      Ways      `xml:"way"`
	Relations Relations `xml:"relation"`

	// Changesets will typically not be included with actually data,
	// but all this stuff is technically all under the osm xml
	Changesets Changesets `xml:"changeset"`
}

// Bound are the bounds of osm data as defined in the xml file.
type Bound struct {
	MinLat float64 `xml:"minlat,attr"`
	MaxLat float64 `xml:"maxlat,attr"`
	MinLon float64 `xml:"minlon,attr"`
	MaxLon float64 `xml:"maxlon,attr"`
}

func (o *OSM) changesetInfo() (ChangesetID, UserID, string) {
	if len(o.Nodes) != 0 {
		n := o.Nodes[0]
		return n.ChangesetID, n.UserID, n.User
	}

	if len(o.Ways) != 0 {
		w := o.Ways[0]
		return w.ChangesetID, w.UserID, w.User
	}

	if len(o.Relations) != 0 {
		r := o.Relations[0]
		return r.ChangesetID, r.UserID, r.User
	}

	return 0, 0, ""
}

// Marshal encodes the osm data using protocol buffers.
func (o *OSM) Marshal() ([]byte, error) {
	ss := &stringSet{}
	encoded := marshalOSM(o, ss, true)
	encoded.Strings = ss.Strings()

	return proto.Marshal(encoded)
}

// UnmarshalOSM will unmarshal the data into a OSM object.
func UnmarshalOSM(data []byte) (*OSM, error) {

	pbf := &osmpb.OSM{}
	err := proto.Unmarshal(data, pbf)
	if err != nil {
		return nil, err
	}

	return unmarshalOSM(pbf, pbf.GetStrings(), nil)
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
			MinLat: geoToInt64(o.Bound.MinLat),
			MaxLat: geoToInt64(o.Bound.MaxLat),
			MinLon: geoToInt64(o.Bound.MinLon),
			MaxLon: geoToInt64(o.Bound.MaxLon),
		}
	}

	return encoded
}

func unmarshalOSM(encoded *osmpb.OSM, ss []string, cs *Changeset) (*OSM, error) {
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
			n, err := unmarshalNode(en, ss, cs)
			if err != nil {
				return nil, err
			}

			o.Nodes[i] = n
		}
	}

	if encoded.DenseNodes != nil {
		var err error
		o.Nodes, err = unmarshalNodes(encoded.DenseNodes, ss, cs)
		if err != nil {
			return nil, err
		}
	}

	if len(encoded.Ways) != 0 {
		o.Ways = make([]*Way, len(encoded.Ways), len(encoded.Ways))
		for i, ew := range encoded.Ways {
			w, err := unmarshalWay(ew, ss, cs)
			if err != nil {
				return nil, err
			}

			o.Ways[i] = w
		}
	}

	if len(encoded.Relations) != 0 {
		o.Relations = make([]*Relation, len(encoded.Relations), len(encoded.Relations))
		for i, er := range encoded.Relations {
			r, err := unmarshalRelation(er, ss, cs)
			if err != nil {
				return nil, err
			}

			o.Relations[i] = r
		}
	}

	if encoded.Bounds != nil {
		o.Bound = &Bound{
			MinLat: float64(encoded.Bounds.GetMinLat()) / locMultiple,
			MaxLat: float64(encoded.Bounds.GetMaxLat()) / locMultiple,
			MinLon: float64(encoded.Bounds.GetMinLon()) / locMultiple,
			MaxLon: float64(encoded.Bounds.GetMaxLon()) / locMultiple,
		}
	}

	return o, nil
}
