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
	Create  []OSM    `xml:"create"`
	Modify  []OSM    `xml:"modify"`
	Delete  []OSM    `xml:"delete"`
}

// Marshal encodes the osm data using protocol buffers.
func (c Change) Marshal() ([]byte, error) {
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

func marshalChange(c Change, ss *stringSet, includeChangeset bool) *osmpb.Change {
	encoded := &osmpb.Change{}

	// TODO: make so only need to scan creates once.

	encoded.Create = marshalOSM(
		OSM{
			Nodes:     c.CreatedNodes(),
			Ways:      c.CreatedWays(),
			Relations: c.CreatedRelations(),
		},
		ss,
		includeChangeset,
	)

	encoded.Modify = marshalOSM(
		OSM{
			Nodes:     c.ModifiedNodes(),
			Ways:      c.ModifiedWays(),
			Relations: c.ModifiedRelations(),
		},
		ss,
		includeChangeset,
	)

	encoded.Delete = marshalOSM(
		OSM{
			Nodes:     c.DeletedNodes(),
			Ways:      c.DeletedWays(),
			Relations: c.DeletedRelations(),
		},
		ss,
		includeChangeset,
	)

	// TODO: bound?

	return encoded
}

func unmarshalChange(encoded *osmpb.Change, ss []string) (*Change, error) {
	c := &Change{}

	creates, err := unmarshalOSM(encoded.Create, ss)
	if err != nil {
		return nil, err
	}
	c.Create = inflateOSM(creates)

	modifies, err := unmarshalOSM(encoded.Modify, ss)
	if err != nil {
		return nil, err
	}
	c.Modify = inflateOSM(modifies)

	deletes, err := unmarshalOSM(encoded.Delete, ss)
	if err != nil {
		return nil, err
	}
	c.Delete = inflateOSM(deletes)

	return c, nil
}

func inflateOSM(o OSM) []OSM {
	l := len(o.Nodes) + len(o.Ways) + len(o.Relations)
	result := make([]OSM, 0, l)

	for _, n := range o.Nodes {
		result = append(result, OSM{Nodes: Nodes{n}})
	}

	for _, w := range o.Ways {
		result = append(result, OSM{Ways: Ways{w}})
	}

	for _, r := range o.Relations {
		result = append(result, OSM{Relations: Relations{r}})
	}

	return result
}

// CreatedNodes returns the list of created nodes in this change.
func (c Change) CreatedNodes() Nodes {
	var ns Nodes
	for _, o := range c.Create {
		ns = append(ns, o.Nodes...)
	}

	return ns
}

// CreatedWays returns the list of created ways in this change.
func (c Change) CreatedWays() Ways {
	var ws Ways
	for _, o := range c.Create {
		ws = append(ws, o.Ways...)
	}

	return ws
}

// CreatedRelations returns the list of created relations in this change.
func (c Change) CreatedRelations() Relations {
	var rs Relations
	for _, o := range c.Create {
		rs = append(rs, o.Relations...)
	}

	return rs
}

// ModifiedNodes returns the list of modified nodes in this change.
func (c Change) ModifiedNodes() Nodes {
	var ns Nodes
	for _, o := range c.Modify {
		ns = append(ns, o.Nodes...)
	}

	return ns
}

// ModifiedWays returns the list of modified ways in this change.
func (c Change) ModifiedWays() Ways {
	var ws Ways
	for _, o := range c.Modify {
		ws = append(ws, o.Ways...)
	}

	return ws
}

// ModifiedRelations returns the list of modified relations in this change.
func (c Change) ModifiedRelations() Relations {
	var rs Relations
	for _, o := range c.Modify {
		rs = append(rs, o.Relations...)
	}

	return rs
}

// DeletedNodes returns the list of deleted nodes in this change.
func (c Change) DeletedNodes() Nodes {
	var ns Nodes
	for _, o := range c.Delete {
		ns = append(ns, o.Nodes...)
	}

	return ns
}

// DeletedWays returns the list of deleted ways in this change.
func (c Change) DeletedWays() Ways {
	var ws Ways
	for _, o := range c.Delete {
		ws = append(ws, o.Ways...)
	}

	return ws
}

// DeletedRelations returns the list of deleted relations in this change.
func (c Change) DeletedRelations() Relations {
	var rs Relations
	for _, o := range c.Delete {
		rs = append(rs, o.Relations...)
	}

	return rs
}
