package osm

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestChange(t *testing.T) {
	data := []byte(`
<osmChange version="0.6" generator="OpenStreetMap server" copyright="OpenStreetMap and contributors" attribution="http://www.openstreetmap.org/copyright" license="http://opendatacommons.org/licenses/odbl/1-0/">
<create>
<node id="2780675158" changeset="21598503" timestamp="2014-04-10T00:43:05Z" version="1" visible="true" user="jeroenrozema74" uid="183483" lat="50.7107023" lon="6.0043943"/>
</create>
<create>
<node id="2780675159" changeset="21598503" timestamp="2014-04-10T00:43:05Z" version="1" visible="true" user="jeroenrozema74" uid="183483" lat="50.710755" lon="5.9998612"/>
</create>
<create>
<way id="273193870" changeset="21598503" timestamp="2014-04-10T00:43:07Z" version="1" visible="true" user="jeroenrozema74" uid="183483">
<nd ref="2780675158"/>
<nd ref="2780675160"/>
<nd ref="2780675161"/>
<nd ref="2780675162"/>
<nd ref="2780675164"/>
<tag k="highway" v="unclassified"/>
</way>
</create>
<modify>
<way id="24830559" changeset="21598503" timestamp="2014-04-10T00:43:07Z" version="9" visible="true" user="jeroenrozema74" uid="183483">
<nd ref="269419649"/>
<nd ref="269419627"/>
<nd ref="166810716"/>
<nd ref="1149226084"/>
<nd ref="269704932"/>
<nd ref="269419651"/>
<nd ref="2751084538"/>
<nd ref="269419653"/>
<nd ref="269419654"/>
<nd ref="269419655"/>
<nd ref="2780675158"/>
<nd ref="269658287"/>
<nd ref="2351330343"/>
<nd ref="269419658"/>
<tag k="highway" v="tertiary"/>
<tag k="name" v="Krikelsteinstraße"/>
</way>
</modify>
<delete>
<way id="252107750" changeset="21598503" timestamp="2014-04-10T00:43:11Z" version="3" visible="false" user="jeroenrozema74" uid="183483"/>
</delete>
<delete>
<way id="252107748" changeset="21598503" timestamp="2014-04-10T00:43:11Z" version="4" visible="false" user="jeroenrozema74" uid="183483"/>
</delete>
<delete>
<node id="301847601" changeset="21598503" timestamp="2014-04-10T00:43:15Z" version="2" visible="false" user="jeroenrozema74" uid="183483"/>
</delete>
</osmChange>
	`)

	c := Change{}
	err := xml.Unmarshal(data, &c)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if l := len(c.Create.Nodes); l != 2 {
		t.Errorf("incorrect number of created nodes, got %v", l)
	}

	if l := len(c.Create.Ways); l != 1 {
		t.Errorf("incorrect number of created ways, got %v", l)
	}

	if v := c.Create.Nodes[0].ID; v != 2780675158 {
		t.Errorf("incorrect node id, got %v", v)
	}

	if v := c.Create.Nodes[1].ID; v != 2780675159 {
		t.Errorf("incorrect node id, got %v", v)
	}

	if v := c.Create.Ways[0].ID; v != 273193870 {
		t.Errorf("incorrect way id, got %v", v)
	}

	if l := len(c.Modify.Ways); l != 1 {
		t.Errorf("incorrect number of modified ways, got %v", l)
	}

	if v := c.Modify.Ways[0].ID; v != 24830559 {
		t.Errorf("incorrect way id, got %v", v)
	}

	if l := len(c.Delete.Nodes); l != 1 {
		t.Errorf("incorrect number of deleted nodes, got %v", l)
	}

	if l := len(c.Delete.Ways); l != 2 {
		t.Errorf("incorrect number of deleted ways, got %v", l)
	}

	if v := c.Delete.Ways[0].ID; v != 252107750 {
		t.Errorf("incorrect way id, got %v", v)
	}

	if v := c.Delete.Ways[1].ID; v != 252107748 {
		t.Errorf("incorrect way id, got %v", v)
	}

	if v := c.Delete.Nodes[0].ID; v != 301847601 {
		t.Errorf("incorrect node id, got %v", v)
	}
}

func TestChangeMarshal(t *testing.T) {
	c1 := loadChange(t, "testdata/changeset_38162206.osc")
	data, err := c1.Marshal()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	c2, err := UnmarshalChange(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("changes are not equal")
		t.Logf("%+v", c1)
		t.Logf("%+v", c2)
	}

	// second changeset
	c1 = loadChange(t, "testdata/changeset_38162210.osc")
	data, err = c1.Marshal()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	c2, err = UnmarshalChange(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("changes are not equal")
		t.Logf("%+v", c1)
		t.Logf("%+v", c2)
	}
}

func compareChange(t *testing.T, c1, c2 *Change) {
	t.Logf("Created Nodes: %d %d", len(c1.Create.Nodes), len(c2.Create.Nodes))
	t.Logf("Modified Nodes: %d %d", len(c1.Modify.Nodes), len(c2.Modify.Nodes))
	t.Logf("Deleted Nodes: %d %d", len(c1.Delete.Nodes), len(c2.Delete.Nodes))

	// c1CreatedNodes := c1.CreatedNodes()
	// c1CreatedNodes.SortByIDVersion()

	// c2CreatedNodes := c2.CreatedNodes()
	// c2CreatedNodes.SortByIDVersion()

	// if !reflect.DeepEqual(c1CreatedNodes, c2CreatedNodes) {
	// 	t.Errorf("created nodes not equal")
	// 	t.Logf("%+v", c1CreatedNodes)
	// 	t.Logf("%+v", c2CreatedNodes)
	// }

	// c1ModifiedNodes := c1.ModifiedNodes()
	// c1ModifiedNodes.SortByIDVersion()

	// c2ModifiedNodes := c2.ModifiedNodes()
	// c2ModifiedNodes.SortByIDVersion()

	// if !reflect.DeepEqual(c1ModifiedNodes, c2ModifiedNodes) {
	// 	t.Errorf("modified nodes not equal")
	// 	t.Logf("%+v", c1ModifiedNodes)
	// 	t.Logf("%+v", c2ModifiedNodes)
	// }

	// c1DeletedNodes := c1.DeletedNodes()
	// c1DeletedNodes.SortByIDVersion()

	// c2DeletedNodes := c2.DeletedNodes()
	// c2DeletedNodes.SortByIDVersion()

	// if !reflect.DeepEqual(c1DeletedNodes, c2DeletedNodes) {
	// 	t.Errorf("deleted nodes not equal")
	// 	t.Logf("%+v", c1DeletedNodes)
	// 	t.Logf("%+v", c2DeletedNodes)
	// }

	// if len(c1.Create) != len(c2.Create) {
	// 	t.Errorf("creates not same length: %d != %d", len(c1.Create), len(c2.Create))
	// 	return
	// }

	// for i := range c1.Create {
	// 	t.Logf("create %d", i)
	// 	compareOSM(t, c1.Create[i], c2.Create[i])
	// }

	// if len(c1.Modify) != len(c2.Modify) {
	// 	t.Errorf("modifies not same length: %d != %d", len(c1.Modify), len(c2.Modify))
	// 	return
	// }

	// for i := range c1.Modify {
	// 	t.Logf("modify %d", i)
	// 	compareOSM(t, c1.Modify[i], c2.Modify[i])
	// }

	// if len(c1.Delete) != len(c2.Delete) {
	// 	t.Errorf("deletes not same length: %d != %d", len(c1.Delete), len(c2.Delete))
	// 	return
	// }

	// for i := range c1.Delete {
	// 	t.Logf("delete %d", i)
	// 	compareOSM(t, c1.Delete[i], c2.Delete[i])
	// }
}

func compareOSM(t *testing.T, o1, o2 OSM) {
	if len(o1.Nodes) != len(o2.Nodes) {
		t.Errorf("node length not equal: %d != %d", len(o1.Nodes), len(o2.Nodes))
		return
	}

	for i := range o1.Nodes {
		if !reflect.DeepEqual(o1.Nodes[i], o2.Nodes[i]) {
			t.Errorf("nodes not equal: %d", i)
			t.Logf("%+v", o1.Nodes[i])
			t.Logf("%+v", o2.Nodes[i])
		}
	}
}
