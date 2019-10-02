package osm

import (
	"encoding/xml"
	"strconv"

	"github.com/paulmach/osm/internal/osmpb"

	"github.com/gogo/protobuf/proto"
)

// Change is the structure of a changeset to be
// uploaded or downloaded from the osm api server.
// See: http://wiki.openstreetmap.org/wiki/OsmChange
type Change struct {
	Version   float64 `xml:"version,attr,omitempty"`
	Generator string  `xml:"generator,attr,omitempty"`

	// to indicate the origin of the data
	Copyright   string `xml:"copyright,attr,omitempty"`
	Attribution string `xml:"attribution,attr,omitempty"`
	License     string `xml:"license,attr,omitempty"`

	Create *OSM `xml:"create"`
	Modify *OSM `xml:"modify"`
	Delete *OSM `xml:"delete"`
}

// AppendCreate will append the object to the Create OSM object.
func (c *Change) AppendCreate(o Object) {
	if c.Create == nil {
		c.Create = &OSM{}
	}

	c.Create.Append(o)
}

// AppendModify will append the object to the Modify OSM object.
func (c *Change) AppendModify(o Object) {
	if c.Modify == nil {
		c.Modify = &OSM{}
	}

	c.Modify.Append(o)
}

// AppendDelete will append the object to the Delete OSM object.
func (c *Change) AppendDelete(o Object) {
	if c.Delete == nil {
		c.Delete = &OSM{}
	}

	c.Delete.Append(o)
}

// HistoryDatasource converts the change object to a datasource accessible
// by feature id. All the creates, modifies and deletes will be added
// in that order.
func (c *Change) HistoryDatasource() *HistoryDatasource {
	ds := &HistoryDatasource{}

	ds.add(c.Create, true)
	ds.add(c.Modify, true)
	ds.add(c.Delete, false)

	return ds
}

// Marshal encodes the osm change data using protocol buffers.
func (c *Change) Marshal() ([]byte, error) {
	ss := &stringSet{}
	encoded := marshalChange(c, ss, true)
	encoded.Strings = ss.Strings()

	return proto.Marshal(encoded)
}

// UnmarshalChange will unmarshal the data into a Change object.
func UnmarshalChange(data []byte) (*Change, error) {

	pbf := &osmpb.Change{}
	err := proto.Unmarshal(data, pbf)
	if err != nil {
		return nil, err
	}

	return unmarshalChange(pbf, pbf.GetStrings(), nil)
}

func marshalChange(c *Change, ss *stringSet, includeChangeset bool) *osmpb.Change {
	if c == nil {
		return nil
	}

	return &osmpb.Change{
		Create: marshalOSM(c.Create, ss, includeChangeset),
		Modify: marshalOSM(c.Modify, ss, includeChangeset),
		Delete: marshalOSM(c.Delete, ss, includeChangeset),
	}
}

func unmarshalChange(encoded *osmpb.Change, ss []string, cs *Changeset) (*Change, error) {
	var err error
	c := &Change{}

	c.Create, err = unmarshalOSM(encoded.Create, ss, cs)
	if err != nil {
		return nil, err
	}

	c.Modify, err = unmarshalOSM(encoded.Modify, ss, cs)
	if err != nil {
		return nil, err
	}

	c.Delete, err = unmarshalOSM(encoded.Delete, ss, cs)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// MarshalXML implements the xml.Marshaller method to allow for the
// correct wrapper/start element case and attr data.
func (c Change) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "osmChange"
	start.Attr = []xml.Attr{}

	if c.Version != 0 {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "version"},
			Value: strconv.FormatFloat(c.Version, 'g', -1, 64),
		})
	}

	if c.Generator != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "generator"}, Value: c.Generator})
	}

	if c.Copyright != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "copyright"}, Value: c.Copyright})
	}

	if c.Attribution != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "attribution"}, Value: c.Attribution})
	}

	if c.License != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "license"}, Value: c.License})
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if err := marshalInnerChange(e, "create", c.Create); err != nil {
		return err
	}

	if err := marshalInnerChange(e, "modify", c.Modify); err != nil {
		return err
	}

	if err := marshalInnerChange(e, "delete", c.Delete); err != nil {
		return err
	}

	return e.EncodeToken(start.End())
}

func marshalInnerChange(e *xml.Encoder, name string, o *OSM) error {
	if o == nil {
		return nil
	}

	t := xml.StartElement{Name: xml.Name{Local: name}}
	if err := e.EncodeToken(t); err != nil {
		return err
	}

	if err := o.marshalInnerXML(e); err != nil {
		return err
	}

	return e.EncodeToken(t.End())
}
