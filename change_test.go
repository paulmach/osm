package osm

import (
	"encoding/xml"
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

	if l := len(c.Creates); l != 3 {
		t.Errorf("incorrect number of creates, got %v", l)
	}

	if v := c.Creates[0].Nodes[0].ID; v != 2780675158 {
		t.Errorf("incorrect node id, got %v", v)
	}

	if v := c.Creates[1].Nodes[0].ID; v != 2780675159 {
		t.Errorf("incorrect node id, got %v", v)
	}

	if v := c.Creates[2].Ways[0].ID; v != 273193870 {
		t.Errorf("incorrect way id, got %v", v)
	}

	if l := len(c.Modifies); l != 1 {
		t.Errorf("incorrect number of modifies, got %v", l)
	}

	if v := c.Modifies[0].Ways[0].ID; v != 24830559 {
		t.Errorf("incorrect way id, got %v", v)
	}

	if l := len(c.Deletes); l != 3 {
		t.Errorf("incorrect number of deletes, got %v", l)
	}

	if v := c.Deletes[0].Ways[0].ID; v != 252107750 {
		t.Errorf("incorrect way id, got %v", v)
	}

	if v := c.Deletes[1].Ways[0].ID; v != 252107748 {
		t.Errorf("incorrect way id, got %v", v)
	}

	if v := c.Deletes[2].Nodes[0].ID; v != 301847601 {
		t.Errorf("incorrect node id, got %v", v)
	}
}
