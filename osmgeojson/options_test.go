package osmgeojson

import (
	"encoding/xml"
	"testing"

	"github.com/onMaps/osm"
	"github.com/paulmach/orb/geojson"
)

var nodeXML = `
<osm>
	<relation id="1">
		<member type="node" ref="1" role="roo" />
	</relation>
	<node
		id="1"
		lat="1.234" lon="4.321"
		timestamp="2013-01-13T22:56:07Z"
		version="7"
		changeset="1234"
		user="johndoe"
		uid="123" />
</osm>`

var wayXML = `
<osm>
	<relation id="1">
		<member type="way" ref="1" role="roo" />
	</relation>
	<way
		id="1"
		lat="1.234" lon="4.321"
		timestamp="2013-01-13T22:56:07Z"
		version="7"
		changeset="1234"
		user="johndoe"
		uid="123">
		<nd ref="1" />
		<nd ref="2" />
	</way>
	<node id="1" lat="1" lon="1" />
	<node id="2" lat="2" lon="2" />
</osm>`

var relationXML = `
<osm>
	<relation id="1">
		<member type="relation" ref="1" role="roo" />
	</relation>
	<relation
		id="1"
		lat="1.234" lon="4.321"
		timestamp="2013-01-13T22:56:07Z"
		version="7"
		changeset="1234"
		user="johndoe"
		uid="123">
		<tag k="type" v="multipolygon" />
		<member type="way" ref="1" />
	</relation>
	<way id="1">
		<nd ref="1" />
		<nd ref="2" />
		<nd ref="3" />
		<nd ref="1" />
	</way>
	<node id="1" lat="1" lon="1" />
	<node id="2" lat="2" lon="2" />
	<node id="3" lat="3" lon="3" />
</osm>`

func TestOptionNoID(t *testing.T) {
	test := func(t *testing.T, xml string) {
		feature := convertXML(t, xml).Features[0]
		if v := feature.ID; v == nil {
			t.Errorf("id should be set: %v", v)
		}

		feature = convertXML(t, xml, NoID(true)).Features[0]
		if v := feature.ID; v != nil {
			t.Errorf("id should be nil: %v", v)
		}
	}

	t.Run("node", func(t *testing.T) {
		test(t, nodeXML)
	})

	t.Run("way", func(t *testing.T) {
		test(t, wayXML)
	})

	t.Run("relation", func(t *testing.T) {
		test(t, relationXML)
	})
}

func TestOptionNoMeta(t *testing.T) {
	test := func(t *testing.T, xml string) {
		feature := convertXML(t, xml).Features[0]
		if v := feature.Properties["meta"]; v == nil {
			t.Errorf("meta should be set: %v", v)
		}

		feature = convertXML(t, xml, NoMeta(true)).Features[0]
		if v := feature.Properties["meta"]; v != nil {
			t.Errorf("meta should be nil: %v", v)
		}
	}

	t.Run("node", func(t *testing.T) {
		test(t, nodeXML)
	})

	t.Run("way", func(t *testing.T) {
		test(t, wayXML)
	})

	t.Run("relation", func(t *testing.T) {
		test(t, relationXML)
	})
}

func TestOptionNoRelationMembership(t *testing.T) {
	test := func(t *testing.T, xml string) {
		feature := convertXML(t, xml).Features[0]
		if v := feature.Properties["relations"]; v == nil {
			t.Errorf("relations should be set: %v", v)
		}

		feature = convertXML(t, xml, NoRelationMembership(true)).Features[0]
		if v := feature.Properties["relations"]; v != nil {
			t.Errorf("relations should be nil: %v", v)
		}
	}

	t.Run("node", func(t *testing.T) {
		test(t, nodeXML)
	})

	t.Run("way", func(t *testing.T) {
		test(t, wayXML)
	})

	t.Run("relation", func(t *testing.T) {
		test(t, relationXML)
	})
}

func convertXML(t *testing.T, data string, opts ...Option) *geojson.FeatureCollection {
	o := &osm.OSM{}
	err := xml.Unmarshal([]byte(data), &o)
	if err != nil {
		t.Fatalf("failed to unmarshal xml: %v", err)
	}

	fc, err := Convert(o, opts...)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	return fc
}
