package osm

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"
)

func TestOSMMarshal(t *testing.T) {
	c := loadChange(t, "testdata/changeset_38162206.osc")
	o1 := flattenOSM(c)
	o1.Bounds = &Bounds{1.1, 2.2, 3.3, 4.4}

	data, err := o1.Marshal()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	o2, err := UnmarshalOSM(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(o1, o2) {
		t.Errorf("osm are not equal")
		t.Logf("%+v", o1)
		t.Logf("%+v", o2)
	}

	// second changeset
	c = loadChange(t, "testdata/changeset_38162210.osc")
	o1 = flattenOSM(c)

	data, err = o1.Marshal()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	o2, err = UnmarshalOSM(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(o1, o2) {
		t.Errorf("osm are not equal")
		t.Logf("%+v", o1)
		t.Logf("%+v", o2)
	}
}

func TestOSMMarshalJSON(t *testing.T) {
	o := &OSM{
		Version:   0.6,
		Generator: "go.osm",
		Nodes: Nodes{
			&Node{ID: 123},
		},
		Ways: Ways{
			&Way{ID: 456},
		},
		Relations: Relations{
			&Relation{ID: 789},
		},
		Changesets: Changesets{
			&Changeset{ID: 10},
		},
	}

	data, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`{"version":0.6,"generator":"go.osm","elements":[{"type":"node","id":123,"lat":0,"lon":0,"visible":false,"timestamp":"0001-01-01T00:00:00Z"},{"type":"way","id":456,"visible":false,"timestamp":"0001-01-01T00:00:00Z","nodes":[]},{"type":"relation","id":789,"visible":false,"timestamp":"0001-01-01T00:00:00Z","members":[]},{"type":"changeset","id":10,"created_at":"0001-01-01T00:00:00Z","closed_at":"0001-01-01T00:00:00Z","open":false}]}`)) {
		t.Errorf("incorrect json: %v", string(data))
	}
}

func TestOSMMarshalXML(t *testing.T) {
	o := &OSM{
		Version:     0.7,
		Generator:   "go.osm-test",
		Copyright:   "copyright1",
		Attribution: "attribution1",
		License:     "license1",
		Nodes: Nodes{
			&Node{ID: 123},
		},
	}

	data, err := xml.Marshal(o)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	expected := `<osm version="0.7" generator="go.osm-test" copyright="copyright1" attribution="attribution1" license="license1"><node id="123" lat="0" lon="0" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"></node></osm>`

	if !bytes.Equal(data, []byte(expected)) {
		t.Errorf("incorrect marshal, got: %s", string(data))
	}

	// omit attributes if not defined
	o = &OSM{}
	data, err = xml.Marshal(o)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	expected = `<osm></osm>`
	if !bytes.Equal(data, []byte(expected)) {
		t.Errorf("incorrect marshal, got: %s", string(data))
	}
}

func flattenOSM(c *Change) *OSM {
	o := c.Create
	if o == nil {
		o = &OSM{}
	}

	if c.Modify != nil {
		o.Nodes = append(o.Nodes, c.Modify.Nodes...)
		o.Ways = append(o.Ways, c.Modify.Ways...)
		o.Relations = append(o.Relations, c.Modify.Relations...)
	}

	if c.Delete != nil {
		o.Nodes = append(o.Nodes, c.Delete.Nodes...)
		o.Ways = append(o.Ways, c.Delete.Ways...)
		o.Relations = append(o.Relations, c.Delete.Relations...)
	}

	return o
}

func cleanXMLNameFromOSM(o *OSM) {
	for _, n := range o.Nodes {
		n.XMLName = xmlNameJSONTypeNode{}
	}

	for _, w := range o.Ways {
		w.XMLName = xmlNameJSONTypeWay{}
	}

	for _, r := range o.Relations {
		// r.XMLName = xml.Name{}
		r.XMLName = xmlNameJSONTypeRel{}
	}
}
