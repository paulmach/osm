package osm

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	"github.com/paulmach/osm/internal/osmpb"

	"github.com/gogo/protobuf/proto"
)

// These values should be returned if the osm data is actual
// osm data to give some information about the source and license.
const (
	Copyright   = "OpenStreetMap and contributors"
	Attribution = "http://www.openstreetmap.org/copyright"
	License     = "http://opendatacommons.org/licenses/odbl/1-0/"
)

// OSM represents the core osm data
// designed to parse http://wiki.openstreetmap.org/wiki/OSM_XML
type OSM struct {
	Version   float64 `xml:"version,attr,omitempty"`
	Generator string  `xml:"generator,attr,omitempty"`

	// These three attributes are returned by the osm api.
	// The Copyright, Attribution and License constants contain
	// suggested values that match those returned by the official api.
	Copyright   string `xml:"copyright,attr,omitempty"`
	Attribution string `xml:"attribution,attr,omitempty"`
	License     string `xml:"license,attr,omitempty"`

	Bounds    *Bounds   `xml:"bounds,omitempty"`
	Nodes     Nodes     `xml:"node"`
	Ways      Ways      `xml:"way"`
	Relations Relations `xml:"relation"`

	// Changesets will typically not be included with actual data,
	// but all this stuff is technically all under the osm xml
	Changesets Changesets `xml:"changeset"`
	Notes      Notes      `xml:"note"`
	Users      Users      `xml:"user"`
}

// Marshal encodes the osm data using protocol buffers.
// Will only save the elements: nodes, ways and relations.
func (o *OSM) Marshal() ([]byte, error) {
	ss := &stringSet{}
	encoded := marshalOSM(o, ss, true)
	encoded.Strings = ss.Strings()

	return proto.Marshal(encoded)
}

// Append will add the given object to the OSM object.
func (o *OSM) Append(obj Object) {
	switch obj.ObjectID().Type() {
	case TypeNode:
		o.Nodes = append(o.Nodes, obj.(*Node))
	case TypeWay:
		o.Ways = append(o.Ways, obj.(*Way))
	case TypeRelation:
		o.Relations = append(o.Relations, obj.(*Relation))
	case TypeChangeset:
		o.Changesets = append(o.Changesets, obj.(*Changeset))
	case TypeNote:
		o.Notes = append(o.Notes, obj.(*Note))
	case TypeUser:
		o.Users = append(o.Users, obj.(*User))
	case TypeBounds:
		o.Bounds = obj.(*Bounds)
	default:
		panic(fmt.Sprintf("unsupported type: %[1]T: %[1]v", obj))
	}
}

// Elements returns all the nodes, ways and relations
// as a single slice of Elements.
func (o *OSM) Elements() Elements {
	if o == nil {
		return nil
	}

	result := make(Elements, 0, len(o.Nodes)+len(o.Ways)+len(o.Relations))
	for _, e := range o.Nodes {
		result = append(result, e)
	}

	for _, e := range o.Ways {
		result = append(result, e)
	}

	for _, e := range o.Relations {
		result = append(result, e)
	}

	return result
}

// Objects returns an array of objects containing any nodes, ways, relations,
// changesets, notes and users.
func (o *OSM) Objects() Objects {
	if o == nil {
		return nil
	}

	l := len(o.Nodes) + len(o.Ways) + len(o.Relations) + len(o.Changesets) + len(o.Notes) + len(o.Users)
	if o.Bounds != nil {
		l++
	}

	result := make(Objects, 0, l)
	if o.Bounds != nil {
		result = append(result, o.Bounds)
	}

	for _, o := range o.Nodes {
		result = append(result, o)
	}

	for _, o := range o.Ways {
		result = append(result, o)
	}

	for _, o := range o.Relations {
		result = append(result, o)
	}

	for _, o := range o.Changesets {
		result = append(result, o)
	}

	for _, o := range o.Users {
		result = append(result, o)
	}

	for _, o := range o.Notes {
		result = append(result, o)
	}

	return result
}

// FeatureIDs returns the slice of feature ids for all the
// nodes, ways and relations.
func (o *OSM) FeatureIDs() FeatureIDs {
	if o == nil {
		return nil
	}

	result := make(FeatureIDs, 0, len(o.Nodes)+len(o.Ways)+len(o.Relations))
	for _, e := range o.Nodes {
		result = append(result, e.FeatureID())
	}

	for _, e := range o.Ways {
		result = append(result, e.FeatureID())
	}

	for _, e := range o.Relations {
		result = append(result, e.FeatureID())
	}

	return result
}

// ElementIDs returns the slice of element ids for all the
// nodes, ways and relations.
func (o *OSM) ElementIDs() ElementIDs {
	if o == nil {
		return nil
	}

	result := make(ElementIDs, 0, len(o.Nodes)+len(o.Ways)+len(o.Relations))
	for _, e := range o.Nodes {
		result = append(result, e.ElementID())
	}

	for _, e := range o.Ways {
		result = append(result, e.ElementID())
	}

	for _, e := range o.Relations {
		result = append(result, e.ElementID())
	}

	return result
}

// HistoryDatasource converts the osm object to a datasource accessible
// by the feature id.
func (o *OSM) HistoryDatasource() *HistoryDatasource {
	ds := &HistoryDatasource{}

	ds.add(o)
	return ds
}

// UnmarshalOSM will unmarshal the data into an OSM object.
func UnmarshalOSM(data []byte) (*OSM, error) {

	pbf := &osmpb.OSM{}
	err := proto.Unmarshal(data, pbf)
	if err != nil {
		return nil, err
	}

	return unmarshalOSM(pbf, pbf.GetStrings(), nil)
}

// includeChangeset can be set to false to not repeat the changeset
// info every item, if this comes from osm change data.
func marshalOSM(o *OSM, ss *stringSet, includeChangeset bool) *osmpb.OSM {
	encoded := &osmpb.OSM{}
	if o == nil {
		return nil
	}

	if len(o.Nodes) > 0 {
		encoded.DenseNodes = marshalNodes(o.Nodes, ss, includeChangeset)
	}

	if len(o.Ways) > 0 {
		encoded.Ways = make([]*osmpb.Way, len(o.Ways))
		for i, w := range o.Ways {
			encoded.Ways[i] = marshalWay(w, ss, includeChangeset)
		}
	}

	if len(o.Relations) > 0 {
		encoded.Relations = make([]*osmpb.Relation, len(o.Relations))
		for i, r := range o.Relations {
			encoded.Relations[i] = marshalRelation(r, ss, includeChangeset)
		}
	}

	if o.Bounds != nil {
		encoded.Bounds = &osmpb.Bounds{
			MinLat: geoToInt64(o.Bounds.MinLat),
			MaxLat: geoToInt64(o.Bounds.MaxLat),
			MinLon: geoToInt64(o.Bounds.MinLon),
			MaxLon: geoToInt64(o.Bounds.MaxLon),
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
		o.Nodes = make([]*Node, len(encoded.Nodes))
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
		o.Ways = make([]*Way, len(encoded.Ways))
		for i, ew := range encoded.Ways {
			w, err := unmarshalWay(ew, ss, cs)
			if err != nil {
				return nil, err
			}

			o.Ways[i] = w
		}
	}

	if len(encoded.Relations) != 0 {
		o.Relations = make([]*Relation, len(encoded.Relations))
		for i, er := range encoded.Relations {
			r, err := unmarshalRelation(er, ss, cs)
			if err != nil {
				return nil, err
			}

			o.Relations[i] = r
		}
	}

	if encoded.Bounds != nil {
		o.Bounds = &Bounds{
			MinLat: float64(encoded.Bounds.GetMinLat()) / locMultiple,
			MaxLat: float64(encoded.Bounds.GetMaxLat()) / locMultiple,
			MinLon: float64(encoded.Bounds.GetMinLon()) / locMultiple,
			MaxLon: float64(encoded.Bounds.GetMaxLon()) / locMultiple,
		}
	}

	return o, nil
}

// MarshalJSON allows the tags to be marshalled as an object
// as defined by the overpass osmjson.
// http://overpass-api.de/output_formats.html#json
func (o OSM) MarshalJSON() ([]byte, error) {
	s := struct {
		Version     float64 `json:"version,omitempty"`
		Generator   string  `json:"generator,omitempty"`
		Copyright   string  `json:"copyright,omitempty"`
		Attribution string  `json:"attribution,omitempty"`
		License     string  `json:"license,omitempty"`
		Elements    Objects `json:"elements"`
	}{o.Version, o.Generator, o.Copyright,
		o.Attribution, o.License, o.Objects()}

	return json.Marshal(s)
}

// MarshalXML implements the xml.Marshaller method to allow for the
// correct wrapper/start element case and attr data.
func (o OSM) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "osm"
	start.Attr = make([]xml.Attr, 0, 5)

	if o.Version != 0 {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "version"},
			Value: strconv.FormatFloat(o.Version, 'g', -1, 64),
		})
	}

	if o.Generator != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "generator"}, Value: o.Generator})
	}

	if o.Copyright != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "copyright"}, Value: o.Copyright})
	}

	if o.Attribution != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "attribution"}, Value: o.Attribution})
	}

	if o.License != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "license"}, Value: o.License})
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if err := o.marshalInnerXML(e); err != nil {
		return err
	}

	return e.EncodeToken(start.End())
}

func (o *OSM) marshalInnerXML(e *xml.Encoder) error {
	if o == nil {
		return nil
	}

	if err := e.Encode(o.Bounds); err != nil {
		return err
	}

	if err := e.Encode(o.Nodes); err != nil {
		return err
	}

	if err := e.Encode(o.Ways); err != nil {
		return err
	}

	if err := e.Encode(o.Relations); err != nil {
		return err
	}

	if err := e.Encode(o.Changesets); err != nil {
		return err
	}

	if err := e.Encode(o.Notes); err != nil {
		return err
	}

	return e.Encode(o.Users)
}

func (o *OSM) marshalInnerElementsXML(e *xml.Encoder) error {
	if err := e.Encode(o.Nodes); err != nil {
		return err
	}

	if err := e.Encode(o.Ways); err != nil {
		return err
	}

	return e.Encode(o.Relations)
}
