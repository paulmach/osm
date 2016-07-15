package osm

import (
	"reflect"
	"testing"
)

func TestOSMMarshal(t *testing.T) {
	c := loadChange(t, "testdata/changeset_38162206.osc")
	o1 := flattenOSM(c)
	o1.Bound = &Bound{1, 2, 3, 4}

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
