package osmgeojson

import (
	"encoding/xml"
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/osm"
)

func TestConvert_polygon(t *testing.T) {
	t.Run("building with inner ring", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="6989507" visible="true">
				<member type="way" ref="475373687" role="outer"/>
				<member type="way" ref="475373473" role="inner"/>
				<tag k="building" v="yes"/>
				<tag k="type" v="multipolygon"/>
			</relation>
			<way id="475373687" visible="true">
				<nd ref="4691265023"/>
				<nd ref="4691265647"/>
				<nd ref="4691264630"/>
				<nd ref="4691268540"/>
				<nd ref="4691265023"/>
			</way>
			<way id="475373473" visible="true">
				<nd ref="4691260535"/>
				<nd ref="4691260534"/>
				<nd ref="4691260533"/>
				<nd ref="4691260532"/>
				<nd ref="4691260535"/>
			</way>
			<node id="4691264630" lat="37.8259401" lon="-122.2549185"/>
			<node id="4691265023" lat="37.8264051" lon="-122.2551366"/>
			<node id="4691265647" lat="37.8260931" lon="-122.2547068"/>
			<node id="4691268540" lat="37.8262489" lon="-122.2552916"/>
			<node id="4691260532" lat="37.8262958" lon="-122.2551641"/>
			<node id="4691260533" lat="37.8260840" lon="-122.2548698"/>
			<node id="4691260534" lat="37.8260406" lon="-122.2549281"/>
			<node id="4691260535" lat="37.8262598" lon="-122.2551983"/>
		</osm>`

		polygon := orb.Polygon{
			{{-122.2551366, 37.8264051}, {-122.2552916, 37.8262489}, {-122.2549185, 37.8259401}, {-122.2547068, 37.8260931}, {-122.2551366, 37.8264051}},
			{{-122.2551983, 37.8262598}, {-122.2551641, 37.8262958}, {-122.2548698, 37.8260840}, {-122.2549281, 37.8260406}, {-122.2551983, 37.8262598}},
		}

		feature := geojson.NewFeature(polygon)
		feature.ID = "relation/6989507"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 6989507
		feature.Properties["tags"] = map[string]string{
			"building": "yes",
			"type":     "multipolygon",
		}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})
}

func TestConvert_multiPolygon(t *testing.T) {
	t.Run("invalid simple multipolygon, no outer way", func(t *testing.T) {
		xml := `
		<osm>
			<relation id='1'>
				<tag k='type' v='multipolygon' />
				<member type='way' ref='2' role='outer' />
				<member type='way' ref='3' role='inner' />
			</relation>
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc)
	})

	t.Run("one outer", func(t *testing.T) {
		xml := `
		<osm>
			<relation id='1'>
				<tag k='type' v='multipolygon' />
				<member type='way' ref='2' role='outer' />
				<member type='way' ref='3' role='inner' />
			</relation>
			<way id="2">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
			<way id="3">
				<nd ref="7" />
				<nd ref="8" />
				<nd ref="9" />
				<nd ref="7" />
			</way>
			<node id="4" lat="-2.0" lon="-2.0" />
			<node id="5" lat="-2.0" lon="2.0" />
			<node id="6" lat="2.0" lon="-2.0" />
			<node id="7" lat="-1.5" lon="-1.5" />
			<node id="8" lat="-1.5" lon="1.5" />
			<node id="9" lat="1.5" lon="-1.5" />
		</osm>`

		polygon := orb.Polygon{
			{{-2.0, -2.0}, {2.0, -2.0}, {-2.0, 2.0}, {-2.0, -2.0}},
			{{-1.5, -1.5}, {-1.5, 1.5}, {1.5, -1.5}, {-1.5, -1.5}},
		}

		feature := geojson.NewFeature(polygon)
		feature.ID = "way/2"
		feature.Properties["type"] = "way"
		feature.Properties["id"] = 2
		feature.Properties["relations"] = []*relationSummary{
			{
				Role: "outer",
				ID:   1,
				Tags: map[string]string{"type": "multipolygon"},
			},
		}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("merge rings", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="1" role="outer" />
				<member type="way" ref="3" role="outer" />
				<member type="way" ref="2" role="outer" />
			</relation>
			<way id="1">
				<nd ref="1" />
				<nd ref="2" />
			</way>
			<way id="2">
				<nd ref="2" />
				<nd ref="3" />
			</way>
			<way id="3">
				<nd ref="3" />
				<nd ref="1" />
			</way>
			<node id="1" lat="1.0" lon="1.0" />
			<node id="2" lat="1.0" lon="-1.0" />
			<node id="3" lat="-1.0" lon="1.0" />
		</osm>`

		polygon := orb.Polygon{{{1.0, 1.0}, {-1.0, 1.0}, {1.0, -1.0}, {1.0, 1.0}}}

		feature := geojson.NewFeature(polygon)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{"type": "multipolygon"}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("skip unclosed rings", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="1" role="outer" />
				<member type="way" ref="2" role="outer" />
			</relation>
			<way id="1">
				<nd ref="1" />
				<nd ref="2" />
				<nd ref="3" />
				<nd ref="1" />
			</way>
			<way id="2">
				<nd ref="2" />
				<nd ref="3" />
			</way>
			<node id="1" lat="1.0" lon="1.0" />
			<node id="2" lat="1.0" lon="-1.0" />
			<node id="3" lat="-1.0" lon="1.0" />
		</osm>`

		polygon := orb.Polygon{{{1.0, 1.0}, {-1.0, 1.0}, {1.0, -1.0}, {1.0, 1.0}}}

		feature := geojson.NewFeature(polygon)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{"type": "multipolygon"}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("a multipolygon amenity=xxx with outer line tagged amenity=yyy", func(t *testing.T) {
		// this should result in two features.
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<tag k="amenity" v="xxx" />
				<member type="way" ref="2" role="outer" />
				<member type="way" ref="3" role="inner" />
			</relation>
			<way id="2">
				<tag k="amenity" v="yyy" />
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
			<way id="3">
				<nd ref="7" />
				<nd ref="8" />
				<nd ref="9" />
				<nd ref="7" />
			</way>
			<node id="4" lat="-2.0" lon="-2.0" />
			<node id="5" lat="-2.0" lon="2.0" />
			<node id="6" lat="2.0" lon="-2.0" />
			<node id="7" lat="-1.5" lon="-1.5" />
			<node id="8" lat="-1.5" lon="1.5" />
			<node id="9" lat="1.5" lon="-1.5" />
		</osm>`

		polygon := orb.Polygon{
			{{-2.0, -2.0}, {2.0, -2.0}, {-2.0, 2.0}, {-2.0, -2.0}},
			{{-1.5, -1.5}, {-1.5, 1.5}, {1.5, -1.5}, {-1.5, -1.5}},
		}

		feature := geojson.NewFeature(polygon)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{"type": "multipolygon", "amenity": "xxx"}

		way := geojson.NewFeature(orb.Polygon{polygon[0]})
		way.ID = "way/2"
		way.Properties["type"] = "way"
		way.Properties["id"] = 2
		way.Properties["tags"] = map[string]string{"amenity": "yyy"}
		way.Properties["relations"] = []*relationSummary{
			{
				ID:   1,
				Role: "outer",
				Tags: map[string]string{"type": "multipolygon", "amenity": "xxx"},
			},
		}

		fc := geojson.NewFeatureCollection().Append(feature).Append(way)
		testConvert(t, xml, fc)
	})
}

func TestConvert_relationMembers(t *testing.T) {
	// complex example containing a generic relation, several ways as well as
	// tagged, uninteresting and untagged nodes
	// see https://github.com/openstreetmap/openstreetmap-website/pull/283
	raw := `
	<osm>
		<relation id="4294968148" visible="true" timestamp="2013-05-14T10:33:05Z" version="1" changeset="23123" user="tyrTester06" uid="1178">
		    <member type="way" ref="4295032195" role="line"/>
			<member type="node" ref="4295668179" role="point"/>
			<member type="node" ref="4295668178" role=""/>
			<member type="way" ref="4295032194" role=""/>
			<member type="way" ref="4295032193" role=""/>
			<member type="node" ref="4295668174" role="foo"/>
			<member type="node" ref="4295668175" role="bar"/>
			<tag k="type" v="fancy"/>
		</relation>
		<way id="4295032195" visible="true" timestamp="2013-05-14T10:33:05Z" version="1" changeset="23123" user="tyrTester06" uid="1178">
		  <nd ref="4295668174"/>
		  <nd ref="4295668172"/>
		  <nd ref="4295668171"/>
		  <nd ref="4295668170"/>
		  <nd ref="4295668173"/>
		  <nd ref="4295668175"/>
		  <tag k="highway" v="residential"/>
		</way>
		<way id="4295032194" visible="true" timestamp="2013-05-14T10:33:05Z" version="1" changeset="23123" user="tyrTester06" uid="1178">
		  <nd ref="4295668177"/>
		  <nd ref="4295668178"/>
		  <nd ref="4295668180"/>
		  <tag k="highway" v="service"/>
		</way>
		<way id="4295032193" visible="true" timestamp="2013-05-14T10:33:04Z" version="1" changeset="23123" user="tyrTester06" uid="1178">
		  <nd ref="4295668181"/>
		  <nd ref="4295668178"/>
		  <nd ref="4295668176"/>
		  <tag k="highway" v="service"/>
		</way>
		<node id="4295668172" version="1" changeset="23123" lat="46.4910906" lon="11.2735763" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z">
		  <tag k="highway" v="crossing"/>
		</node>+
		<node id="4295668173" version="1" changeset="23123" lat="46.4911004" lon="11.2759498" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z">
		  <tag k="created_by" v="foo"/>
		</node>+
		<node id="4295668170" version="1" changeset="23123" lat="46.4909732" lon="11.2753813" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
		<node id="4295668171" version="1" changeset="23123" lat="46.4909781" lon="11.2743295" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
		<node id="4295668174" version="1" changeset="23123" lat="46.4914820" lon="11.2731001" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
		<node id="4295668175" version="1" changeset="23123" lat="46.4915603" lon="11.2765254" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
		<node id="4295668176" version="1" changeset="23123" lat="46.4919468" lon="11.2756726" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
		<node id="4295668177" version="1" changeset="23123" lat="46.4919664" lon="11.2753031" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
		<node id="4295668178" version="1" changeset="23123" lat="46.4921083" lon="11.2755021" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
		<node id="4295668179" version="1" changeset="23123" lat="46.4921327" lon="11.2742229" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
		<node id="4295668180" version="1" changeset="23123" lat="46.4922893" lon="11.2757152" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
		<node id="4295668181" version="1" changeset="23123" lat="46.4923235" lon="11.2752747" user="tyrTester06" uid="1178" visible="true" timestamp="2013-05-14T10:33:04Z"/>
	  </osm>`

	o := &osm.OSM{}
	err := xml.Unmarshal([]byte(raw), &o)
	if err != nil {
		t.Fatalf("unable to unmarhsal xml: %v", err)
	}

	fc, err := Convert(o)
	if err != nil {
		t.Fatalf("convert error: %v", err)
	}

	if l := len(fc.Features); l != 8 {
		t.Errorf("incorrect number of features: %d != 8", l)
	}
}

func TestConvert_innerWays(t *testing.T) {
	t.Run("missing inner way", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="2" role="outer" />
				<member type="way" ref="3" role="inner" />
			</relation>
			<way id="2">
				<nd ref="3" />
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="3" />
			</way>
			<node id="3" lon="0.0" lat="0.0" />
			<node id="4" lon="1.0" lat="1.0" />
			<node id="5" lon="1.0" lat="0.0" />
		</osm>`

		polygon := orb.Polygon{{{0, 0}, {1, 0}, {1, 1}, {0, 0}}}

		feature := geojson.NewFeature(polygon)
		feature.ID = "way/2"
		feature.Properties["type"] = "way"
		feature.Properties["id"] = 2
		feature.Properties["relations"] = []*relationSummary{
			{
				Role: "outer",
				ID:   1,
				Tags: map[string]string{"type": "multipolygon"},
			},
		}
		feature.Properties["tainted"] = true

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("should skip the outer way if missing", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="2" role="outer" />
				<member type="way" ref="3" role="outer" />
			</relation>
			<way id="2">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
			<node id="4" lon="0.0" lat="0.0" />
			<node id="5" lon="1.0" lat="1.0" />
			<node id="6" lon="1.0" lat="0.0" />
		</osm>`

		polygon := orb.Polygon{{{0, 0}, {1, 0}, {1, 1}, {0, 0}}}

		feature := geojson.NewFeature(polygon)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{"type": "multipolygon"}
		feature.Properties["tainted"] = true

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("missing node", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="2" role="outer" />
				<member type="way" ref="3" role="outer" />
			</relation>
			<way id="2">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="7" />
				<nd ref="4" />
			</way>
			<way id="3">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
			<node id="4" lon="0.0" lat="0.0" />
			<node id="5" lon="1.0" lat="1.0" />
			<node id="6" lon="1.0" lat="0.0" />
		</osm>`

		multiPolygon := orb.MultiPolygon{
			{
				{{0, 0}, {1, 0}, {1, 1}, {0, 0}},
			},
			{
				{{0, 0}, {1, 0}, {1, 1}, {0, 0}},
			},
		}

		feature := geojson.NewFeature(multiPolygon)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{"type": "multipolygon"}
		feature.Properties["tainted"] = true

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("no coordinates", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="2" role="outer" />
			</relation>
			<way id="2">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc)
	})

	t.Run("no outer ring polygons should be skipped", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="2" role="inner" />
			</relation>
			<way id="2">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
			<node id="4" lon="0.0" lat="0.0" />
			<node id="5" lon="1.0" lat="1.0" />
			<node id="6" lon="1.0" lat="0.0" />
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc)
	})

	t.Run("should not return polygon if all outer ways are tainted", func(t *testing.T) {
		// should also not return it's uninteresting members
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="3" role="outer" />
				<member type="way" ref="2" role="outer" />
			</relation>
			<way id="1">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
			<way id="2">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
			<node id="4" lon="0.0" lat="0.0" />
			<node id="5" lon="1.0" lat="1.0" />
		</osm>`

		// should return one of the ways as a line string
		ls := orb.LineString{{0, 0}, {1, 1}, {0, 0}}

		feature := geojson.NewFeature(ls)
		feature.ID = "way/1"
		feature.Properties["type"] = "way"
		feature.Properties["id"] = 1
		feature.Properties["tainted"] = true

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})
}

func TestConvert_multiPolygonMultiOuter(t *testing.T) {
	raw := `
	<osm>
		<relation id="1">
			<tag k="type" v="multipolygon" />
			<tag k="building" v="yes" />
			<member type="way" ref="2" role="outer" />
			<member type="way" ref="3" role="inner" />
			<member type="way" ref="4" role="inner" />
			<member type="way" ref="5" role="outer" />
		</relation>
		<way id="2">
			<tag k="building" v="yes" />
			<nd ref="4" />
			<nd ref="5" />
			<nd ref="6" />
			<nd ref="7" />
			<nd ref="4" />
		</way>
		<way id="3">
			<tag k="area" v="yes" />
			<nd ref="8" />
			<nd ref="9" />
			<nd ref="10" />
			<nd ref="8" />
		</way>
		<way id="4">
			<tag k="barrier" v="fence" />
			<nd ref="11" />
			<nd ref="12" />
			<nd ref="13" />
			<nd ref="11" />
		</way>
		<way id="5">
			<tag k="building" v="yes" />
			<tag k="area" v="yes" />
			<nd ref="14" />
			<nd ref="15" />
			<nd ref="16" />
			<nd ref="14" />
		</way>
		<node id="4" lat="-1.0" lon="-1.0" />
		<node id="5" lat="-1.0" lon="1.0" />
		<node id="6" lat="1.0" lon="1.0" />
		<node id="7" lat="1.0" lon="-1.0" />
		<node id="8" lat="-0.5" lon="0.0" />
		<node id="9" lat="0.5" lon="0.0" />
		<node id="10" lat="0.5" lon="0.5" />
		<node id="11" lat="0.1" lon="-0.1" />
		<node id="12" lat="-0.1" lon="-0.1" />
		<node id="13" lat="0.0" lon="-0.2" />
		<node id="14" lat="0.1" lon="-1.1" />
		<node id="15" lat="-0.1" lon="-1.1" />
		<node id="16" lat="0.0" lon="-1.2" />
	</osm>`

	t.Run("multi polygon", func(t *testing.T) {
		mp := orb.MultiPolygon{
			{{{-1.1, 0.1}, {-1.2, 0.0}, {-1.1, -0.1}, {-1.1, 0.1}}},
			{
				{{-1.0, -1.0}, {1.0, -1.0}, {1.0, 1.0}, {-1.0, 1.0}, {-1.0, -1.0}},
				{{-0.1, 0.1}, {-0.1, -0.1}, {-0.2, 0.0}, {-0.1, 0.1}},
				{{0.0, -0.5}, {0.0, 0.5}, {0.5, 0.5}, {0.0, -0.5}},
			},
		}

		feature := geojson.NewFeature(mp)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{"type": "multipolygon", "building": "yes"}

		way3 := geojson.NewFeature(orb.Ring{{0, -0.5}, {0.5, 0.5}, {0, 0.5}, {0, -0.5}})
		way3.ID = "way/3"
		way3.Properties["type"] = "way"
		way3.Properties["id"] = 3
		way3.Properties["tags"] = map[string]string{"area": "yes"}
		way3.Properties["relations"] = []*relationSummary{
			{
				Role: "inner",
				ID:   1,
				Tags: map[string]string{"type": "multipolygon", "building": "yes"},
			},
		}

		way4 := geojson.NewFeature(orb.LineString{{-0.1, 0.1}, {-0.1, -0.1}, {-0.2, 0.0}, {-0.1, 0.1}})
		way4.ID = "way/4"
		way4.Properties["type"] = "way"
		way4.Properties["id"] = 4
		way4.Properties["tags"] = map[string]string{"barrier": "fence"}
		way4.Properties["relations"] = []*relationSummary{
			{
				Role: "inner",
				ID:   1,
				Tags: map[string]string{"type": "multipolygon", "building": "yes"},
			},
		}

		way5 := geojson.NewFeature(orb.Ring{{-1.1, 0.1}, {-1.2, 0.0}, {-1.1, -0.1}, {-1.1, 0.1}})
		way5.ID = "way/5"
		way5.Properties["type"] = "way"
		way5.Properties["id"] = 5
		way5.Properties["tags"] = map[string]string{"building": "yes", "area": "yes"}
		way5.Properties["relations"] = []*relationSummary{
			{
				Role: "outer",
				ID:   1,
				Tags: map[string]string{"type": "multipolygon", "building": "yes"},
			},
		}

		fc := geojson.NewFeatureCollection().
			Append(feature).Append(way3).Append(way4).Append(way5)
		testConvert(t, raw, fc)
	})

	t.Run("role-less members should be outer ways", func(t *testing.T) {
		o := &osm.OSM{}
		err := xml.Unmarshal([]byte(raw), &o)
		if err != nil {
			t.Fatalf("xml unmarshal error: %v", err)
		}

		// handle role-less members as outer ways
		o.Relations[0].Members[3].Role = ""

		fc, err := Convert(o)
		if err != nil {
			t.Fatalf("convert error: %v", err)
		}

		mp := fc.Features[0].Geometry.(orb.Polygon)
		if l := len(mp); l != 3 {
			t.Errorf("wrong number of outer rings: %d != 3", l)
		}
	})
}

func TestConvert_includeInvalidPolygons(t *testing.T) {
	t.Run("missing outer ring", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="2" role="inner" />
				<member type="way" ref="3" role="outer" />
			</relation>
			<way id="2">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
			<node id="4" lon="0.0" lat="0.0" />
			<node id="5" lon="1.0" lat="1.0" />
			<node id="6" lon="1.0" lat="0.0" />
		</osm>`

		ls := orb.Polygon{
			nil,
			orb.Ring{{0.0, 0.0}, {1.0, 1.0}, {1.0, 0.0}, {0.0, 0.0}},
		}

		feature := geojson.NewFeature(ls)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tainted"] = true
		feature.Properties["tags"] = map[string]string{"type": "multipolygon"}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc, IncludeInvalidPolygons(true))
	})

	t.Run("missing all rings", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="4" role="outer" />
				<member type="node" ref="2" role="admin_center" />
			</relation>
			<node id="2" lon="0.0" lat="0.0" />
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc, IncludeInvalidPolygons(true))
	})

	t.Run("missing everything", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="4" role="outer" />
				<member type="node" ref="2" role="admin_center" />
			</relation>
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc, IncludeInvalidPolygons(true))
	})

	t.Run("outer ring incomplete", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="2" role="inner" />
				<member type="way" ref="3" role="outer" />
				<member type="way" ref="4" role="outer" />
			</relation>
			<way id="2">
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="6" />
				<nd ref="4" />
			</way>
			<way id="3">
				<nd ref="4" />
				<nd ref="5" />
			</way>
			<node id="4" lon="0.0" lat="0.0" />
			<node id="5" lon="1.0" lat="1.0" />
			<node id="6" lon="1.0" lat="0.0" />
		</osm>`

		ls := orb.Polygon{
			{{1.0, 1.0}, {0.0, 0.0}},
			{{0.0, 0.0}, {1.0, 1.0}, {1.0, 0.0}, {0.0, 0.0}},
		}

		feature := geojson.NewFeature(ls)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tainted"] = true
		feature.Properties["tags"] = map[string]string{"type": "multipolygon"}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc, IncludeInvalidPolygons(true))
	})
}
