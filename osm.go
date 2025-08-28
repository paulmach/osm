package osm

import (
	"encoding/xml"
	"fmt"
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
	// JSON APIs can return version as a string or number, converted to string
	// for consistency.
	Version   string `xml:"version,attr,omitempty"`
	Generator string `xml:"generator,attr,omitempty"`

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

// MarshalJSON allows the tags to be marshalled as an object
// as defined by the overpass osmjson.
// http://overpass-api.de/output_formats.html#json
func (o OSM) MarshalJSON() ([]byte, error) {
	s := struct {
		Version     string  `json:"version,omitempty"`
		Generator   string  `json:"generator,omitempty"`
		Copyright   string  `json:"copyright,omitempty"`
		Attribution string  `json:"attribution,omitempty"`
		License     string  `json:"license,omitempty"`
		Elements    Objects `json:"elements"`
	}{o.Version, o.Generator, o.Copyright, o.Attribution, o.License, o.Objects()}

	return marshalJSON(s)
}

// MarshalXML implements the xml.Marshaller method to allow for the
// correct wrapper/start element case and attr data.
func (o OSM) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "osm"
	start.Attr = make([]xml.Attr, 0, 5)

	if o.Version != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "version"}, Value: o.Version})
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

// UnmarshalJSON will decode osm json representation
// as defined by the overpass osmjson. This format can
// also by returned by the official OSM API.
// http://overpass-api.de/output_formats.html#json
func (o *OSM) UnmarshalJSON(data []byte) error {
	s := struct {
		// Version can be string or number,
		// openstreetmap.org returns string
		// overpass returns number
		Version     interface{}        `json:"version"`
		Generator   string             `json:"generator"`
		Copyright   string             `json:"copyright"`
		Attribution string             `json:"attribution"`
		License     string             `json:"license"`
		Elements    []nocopyRawMessage `json:"elements"`
	}{}

	err := unmarshalJSON(data, &s)
	if err != nil {
		return err
	}

	o.Version = fmt.Sprintf("%v", s.Version)
	o.Generator = s.Generator
	o.Copyright = s.Copyright
	o.Attribution = s.Attribution
	o.License = s.License

	for index, data := range s.Elements {
		t, err := findType(index, data)
		if err != nil {
			return err
		}

		switch t {
		case "node":
			n := &Node{}
			err = unmarshalJSON(data, n)
			if err != nil {
				return err
			}
			o.Nodes = append(o.Nodes, n)
		case "way":
			w := &Way{}
			err = unmarshalJSON(data, w)
			if err != nil {
				return err
			}
			o.Ways = append(o.Ways, w)
		case "relation":
			r := &Relation{}
			err = unmarshalJSON(data, r)
			if err != nil {
				return err
			}
			o.Relations = append(o.Relations, r)
		case "changeset":
			cs := &Changeset{}
			err = unmarshalJSON(data, cs)
			if err != nil {
				return err
			}
			o.Changesets = append(o.Changesets, cs)
		case "note":
			n := &Note{}
			err = unmarshalJSON(data, n)
			if err != nil {
				return err
			}
			o.Notes = append(o.Notes, n)
		case "user":
			u := &User{}
			err = unmarshalJSON(data, u)
			if err != nil {
				return err
			}
			o.Users = append(o.Users, u)
		default:
			return fmt.Errorf("unknown type of '%s' for element index %d", t, index)
		}
	}

	return nil
}

type typeStruct struct {
	Type string `json:"type"`
}

func findType(index int, data []byte) (string, error) {
	ts := typeStruct{}
	err := unmarshalJSON(data, &ts)
	if err != nil {
		// should not happened due to previous decoding succeeded
		return "", err
	}

	if ts.Type == "" {
		return "", fmt.Errorf("could not find type in element index %d", index)
	}

	return ts.Type, nil
}
