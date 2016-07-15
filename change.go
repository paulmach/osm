package osm

import (
	"encoding/xml"

	"github.com/golang/protobuf/proto"
	"github.com/paulmach/go.osm/osmpb"
)

// Change is the structure of a changeset to be
// uploaded or downloaded from the server.
// See: http://wiki.openstreetmap.org/wiki/OsmChange
type Change struct {
	XMLName xml.Name `xml:"osmChange"`
	Create  *OSM     `xml:"create"`
	Modify  *OSM     `xml:"modify"`
	Delete  *OSM     `xml:"delete"`
}

// Marshal encodes the osm data using protocol buffers.
func (c *Change) Marshal() ([]byte, error) {
	ss := &stringSet{}
	encoded := marshalChange(c, ss, true)
	encoded.Strings = ss.Strings()

	return proto.Marshal(encoded)
}

func (c *Change) Unmarshal(data []byte) error {

	pbf := &osmpb.Change{}
	err := proto.Unmarshal(data, pbf)
	if err != nil {
		return err
	}

	change, err := unmarshalChange(pbf, pbf.GetStrings())
	if err != nil {
		return err
	}

	*c = *change
	return nil
}

func marshalChange(c *Change, ss *stringSet, includeChangeset bool) *osmpb.Change {
	encoded := &osmpb.Change{}

	// TODO: make so only need to scan creates once.

	encoded.Create = marshalOSM(c.Create, ss, includeChangeset)
	encoded.Modify = marshalOSM(c.Modify, ss, includeChangeset)
	encoded.Delete = marshalOSM(c.Delete, ss, includeChangeset)

	// TODO: bound?

	return encoded
}

func unmarshalChange(encoded *osmpb.Change, ss []string) (*Change, error) {
	var err error
	c := &Change{}

	c.Create, err = unmarshalOSM(encoded.Create, ss)
	if err != nil {
		return nil, err
	}

	c.Modify, err = unmarshalOSM(encoded.Modify, ss)
	if err != nil {
		return nil, err
	}

	c.Delete, err = unmarshalOSM(encoded.Delete, ss)
	if err != nil {
		return nil, err
	}

	return c, nil
}
