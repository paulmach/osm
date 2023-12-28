package osmgeojson

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/osm"
)

func TestConvert(t *testing.T) {
	t.Run("blank osm", func(t *testing.T) {
		xml := `<osm></osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc)
	})

	t.Run("node", func(t *testing.T) {
		xml := `<osm><node id='1' lat='1.234' lon='4.321' /></osm>`

		feature := geojson.NewFeature(orb.Point{4.321, 1.234})
		feature.ID = "node/1"
		feature.Properties["type"] = "node"
		feature.Properties["id"] = 1

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("way", func(t *testing.T) {
		xml := `
		<osm>
			<way id='1'><nd ref='2' /><nd ref='3' /><nd ref='4' /></way>
			<node id='2' lat='0.0' lon='1.0' />
			<node id='3' lat='0.0' lon='1.1' />
			<node id='4' lat='0.1' lon='1.2' />
		</osm>`

		feature := geojson.NewFeature(orb.LineString{{1, 0}, {1.1, 0}, {1.2, 0.1}})
		feature.ID = "way/1"
		feature.Properties["type"] = "way"
		feature.Properties["id"] = 1

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("simple relation polygon", func(t *testing.T) {
		xml := `
		<osm>
			<relation id='1'>
				<tag k='type' v='multipolygon' />
				<member type='way' ref='2' role='outer' />
				<member type='way' ref='3' role='inner' />
			</relation>
			<way id='2'>
				<tag k='area' v='yes' />
				<nd ref='4' /><nd ref='5' /><nd ref='6' /><nd ref='7' /><nd ref='4' />
			</way>
			<way id='3'>
				<nd ref='8' /><nd ref='9' /><nd ref='10' /><nd ref='8' />
			</way>
			<node id='4' lat='-1.0' lon='-1.0' />
			<node id='5' lat='-1.0' lon='1.0' />n
			<node id='6' lat='1.0' lon='1.0' />
			<node id='7' lat='1.0' lon='-1.0' />
			<node id='8' lat='-0.5' lon='0.0' />
			<node id='9' lat='0.5' lon='0.0' />
			<node id='10' lat='0.0' lon='0.5' />
		</osm>`

		polygon := orb.Polygon{
			{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}, {-1, -1}},
			{{0, -0.5}, {0, 0.5}, {0.5, 0}, {0, -0.5}},
		}

		feature := geojson.NewFeature(polygon)
		feature.ID = "way/2"
		feature.Properties["type"] = "way"
		feature.Properties["id"] = 2
		feature.Properties["tags"] = map[string]string{"area": "yes"}
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

	t.Run("relation with nodes in members", func(t *testing.T) {
		xml := `
		<osm>
			<relation id='1'>
				<tag k='type' v='multipolygon' />
				<tag k="admin_level" v="8"/>
				<tag k="boundary" v="administrative"/>
				<member type='way' ref='2' role='outer'>
					<nd lat='-1.0' lon='-1.0' />
					<nd lat='-1.0' lon='1.0' />
					<nd lat='1.0' lon='1.0' />
					<nd lat='1.0' lon='-1.0' />
					<nd lat='-1.0' lon='-1.0' />
				</member>
				<member type='way' ref='3' role='inner'>
					<nd lat='-0.5' lon='0.0' />
					<nd lat='0.5' lon='0.0' />
					<nd lat='0.0' lon='0.5' />
					<nd lat='-0.5' lon='0.0' />
				</member>
			</relation>
		</osm>`

		polygon := orb.Polygon{
			{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}, {-1, -1}},
			{{0, -0.5}, {0, 0.5}, {0.5, 0}, {0, -0.5}},
		}

		feature := geojson.NewFeature(polygon)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{
			"admin_level": "8",
			"boundary":    "administrative",
			"type":        "multipolygon",
		}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("relation", func(t *testing.T) {
		xml := `
		<osm>
			<relation id='1' version='1' timestamp='2018-01-01T00:00:00Z'
				changeset='123' user='user' uid='431'>
				<tag k='type' v='multipolygon' />
				<tag k='amenity' v='hospital' />
				<member type='way' ref='2' role='outer' />
				<member type='way' ref='3' role='inner' />
			</relation>
			<way id='2'>
				<nd ref='4' /><nd ref='5' /><nd ref='6' /><nd ref='7' /><nd ref='4' />
			</way>
			<way id='3'>
				<nd ref='8' /><nd ref='9' /><nd ref='10' /><nd ref='8' />
			</way>
			<node id='4' lat='-1.0' lon='-1.0' />
			<node id='5' lat='-1.0' lon='1.0' />n
			<node id='6' lat='1.0' lon='1.0' />
			<node id='7' lat='1.0' lon='-1.0' />
			<node id='8' lat='-0.5' lon='0.0' />
			<node id='9' lat='0.5' lon='0.0' />
			<node id='10' lat='0.0' lon='0.5' />
		</osm>`

		polygon := orb.Polygon{
			{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}, {-1, -1}},
			{{0, -0.5}, {0, 0.5}, {0.5, 0}, {0, -0.5}},
		}

		feature := geojson.NewFeature(polygon)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{
			"amenity": "hospital",
			"type":    "multipolygon",
		}
		feature.Properties["meta"] = map[string]interface{}{
			"changeset": osm.ChangesetID(123),
			"timestamp": "2018-01-01T00:00:00Z",
			"uid":       osm.UserID(431),
			"user":      "user",
			"version":   1,
		}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})
}

func TestConvert_interestingNodes(t *testing.T) {
	xml := `
	<osm>
		<way id="1">
			<tag k="foo" v="bar" />
			<nd ref="2" />
			<nd ref="3" />
			<nd ref="4" />
		</way>
		<node id="2" lat="0.0" lon="1.0" />
		<node id="3" lat="0.0" lon="1.1">
			<tag k="asd" v="fasd" />
		</node>
		<node id="4" lat="0.1" lon="1.2">
			<tag k="created_by" v="me" />
		</node>
		<node id="5" lat="0.0" lon="0.0" version="3" />
	</osm>`

	way := geojson.NewFeature(orb.LineString{{1, 0}, {1.1, 0}, {1.2, 0.1}})
	way.ID = "way/1"
	way.Properties["type"] = "way"
	way.Properties["id"] = 1
	way.Properties["tags"] = map[string]string{"foo": "bar"}

	node1 := geojson.NewFeature(orb.Point{1.1, 0})
	node1.ID = "node/3"
	node1.Properties["type"] = "node"
	node1.Properties["id"] = 3
	node1.Properties["tags"] = map[string]string{"asd": "fasd"}

	node2 := geojson.NewFeature(orb.Point{0, 0})
	node2.ID = "node/5"
	node2.Properties["type"] = "node"
	node2.Properties["id"] = 5
	node2.Properties["meta"] = map[string]interface{}{"version": 3}

	fc := geojson.NewFeatureCollection().Append(way).Append(node1).Append(node2)
	testConvert(t, xml, fc)
}

func TestConvert_polygonDetection(t *testing.T) {
	xml := `
	<osm>
		<way id="1">
			<tag k="area" v="yes" />
			<nd ref="1" />
			<nd ref="2" />
			<nd ref="3" />
			<nd ref="1" />
		</way>
		<node id="1" lat="2" lon="2" />
		<node id="2" lat="2" lon="3" />
		<node id="3" lat="3" lon="2" />
	</osm>`

	polygon := orb.Polygon{{{2, 2}, {3, 2}, {2, 3}, {2, 2}}}
	feature := geojson.NewFeature(polygon)
	feature.ID = "way/1"
	feature.Properties["type"] = "way"
	feature.Properties["id"] = 1
	feature.Properties["tags"] = map[string]string{"area": "yes"}

	fc := geojson.NewFeatureCollection().Append(feature)
	testConvert(t, xml, fc)
}

func TestConvert_routeRelation(t *testing.T) {
	t.Run("single section", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="route" />
				<member type="way" ref="2" role="forward" />
			</relation>
			<way id="2"><nd ref="4" /><nd ref="5" /></way>
			<node id="4" lon="-1.0" lat="-1.0" />
			<node id="5" lon="0.0" lat="0.0" />
		</osm>`

		feature := geojson.NewFeature(orb.LineString{{-1, -1}, {0, 0}})
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{"type": "route"}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("multiple sections", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="route" />
				<member type="way" ref="2" role="forward" />
				<member type="way" ref="3" role="backward" />
				<member type="way" ref="4" role="forward" />
			</relation>
			<way id="2"><nd ref="4" /><nd ref="5" /></way>
			<way id="3"><nd ref="5" /><nd ref="6" /></way>
			<way id="4"><nd ref="7" /><nd ref="8" /></way>
			<node id="4" lon="-1.0" lat="-1.0" />
			<node id="5" lon="0.0" lat="0.0" />
			<node id="6" lon="1.0" lat="1.0" />
			<node id="7" lon="10.0" lat="10.0" />
			<node id="8" lon="20.0" lat="20.0" />
		</osm>`

		mls := orb.MultiLineString{
			{{10, 10}, {20, 20}},
			{{-1, -1}, {0, 0}, {1, 1}},
		}

		feature := geojson.NewFeature(mls)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tags"] = map[string]string{"type": "route"}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})
}

func TestConvert_nonInterestingNodes(t *testing.T) {
	t.Run("include node not part of ways, even if boring", func(t *testing.T) {
		xml := `<osm>
			<node id="1" lat="3" lon="4"></node>
		</osm>`

		node := geojson.NewFeature(orb.Point{4, 3})
		node.ID = "node/1"
		node.Properties["type"] = "node"
		node.Properties["id"] = 1

		fc := geojson.NewFeatureCollection().Append(node)
		testConvert(t, xml, fc)
	})

	t.Run("include node that is part of relation, even if boring", func(t *testing.T) {
		xml := `<osm>
			<relation id="3">
				<member type="node" ref="1" />
			</relation>
			<node id="1" lat="3" lon="4"></node>
		</osm>`

		node := geojson.NewFeature(orb.Point{4, 3})
		node.ID = "node/1"
		node.Properties["type"] = "node"
		node.Properties["id"] = 1
		node.Properties["relations"] = []*relationSummary{
			{
				Role: "",
				ID:   3,
			},
		}

		fc := geojson.NewFeatureCollection().Append(node)
		testConvert(t, xml, fc)
	})

	t.Run("include interesting nodes in output", func(t *testing.T) {
		xml := `<osm>
			<way id="1">
				<nd ref="1" />"
				<nd lat="1.0" lon="2.0" />"
				<nd ref="2" />"
			</way>
			<node id="1" lat="3" lon="4">
				<tag k="foo" v="bar"/>
			</node>
		</osm>`

		way := geojson.NewFeature(orb.LineString{{4, 3}, {2, 1}})
		way.ID = "way/1"
		way.Properties["type"] = "way"
		way.Properties["id"] = 1
		way.Properties["tainted"] = true

		node := geojson.NewFeature(orb.Point{4, 3})
		node.ID = "node/1"
		node.Properties["type"] = "node"
		node.Properties["id"] = 1
		node.Properties["tags"] = map[string]string{"foo": "bar"}

		fc := geojson.NewFeatureCollection().Append(way).Append(node)
		testConvert(t, xml, fc)
	})

	t.Run("non referenced nodes should mark way as tainted", func(t *testing.T) {
		xml := `
		<osm>
			<way id="1">
				<nd ref="1" />"
				<nd lat="1.0" lon="2.0" />"
				<nd ref="2" />"
			</way>
			<node id="1" lat="3" lon="4" />
		</osm>`

		feature := geojson.NewFeature(orb.LineString{{4, 3}, {2, 1}})
		feature.ID = "way/1"
		feature.Properties["type"] = "way"
		feature.Properties["id"] = 1
		feature.Properties["tainted"] = true

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})
}

func TestConvert_emptyElements(t *testing.T) {
	t.Run("node", func(t *testing.T) {
		xml := `
		<osm>
			<node id="1" />
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc)
	})

	t.Run("way", func(t *testing.T) {
		xml := `
		<osm>
			<way id="1" />
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc)
	})

	t.Run("relation", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1" />
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc)
	})
}

func TestConvert_tainted(t *testing.T) {
	t.Run("way", func(t *testing.T) {
		xml := `
		<osm>
			<way id="2">
				<nd ref="3" />
				<nd ref="4" />
				<nd ref="5" />
			</way>
			<node id="3" lon="1.0" lat="1.0" />
			<node id="4" lon="1.0" lat="1.0" />
		</osm>`

		feature := geojson.NewFeature(orb.LineString{{1, 1}, {1, 1}})
		feature.ID = "way/2"
		feature.Properties["type"] = "way"
		feature.Properties["id"] = 2
		feature.Properties["tainted"] = true

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})

	t.Run("relation", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="1">
				<tag k="type" v="multipolygon" />
				<member type="way" ref="2" role="outer" />
				<member type="way" ref="3" role="outer" />
			</relation>
			<way id="2">
				<nd ref="3" />
				<nd ref="4" />
				<nd ref="5" />
				<nd ref="3" />
			</way>
			<node id="3" lon="1.0" lat="1.0" />
			<node id="4" lon="0.0" lat="1.0" />
			<node id="5" lon="1.0" lat="0.0" />
		</osm>`

		polygon := orb.Polygon{{{1, 1}, {0, 1}, {1, 0}, {1, 1}}}

		feature := geojson.NewFeature(polygon)
		feature.ID = "relation/1"
		feature.Properties["type"] = "relation"
		feature.Properties["id"] = 1
		feature.Properties["tainted"] = true
		feature.Properties["tags"] = map[string]string{"type": "multipolygon"}

		fc := geojson.NewFeatureCollection().Append(feature)
		testConvert(t, xml, fc)
	})
}

func TestConvert_taintedEmptyElements(t *testing.T) {
	t.Run("one node way", func(t *testing.T) {
		xml := `
		<osm>
			<way id="2">
				<tag k="foo" v="bar" />
				<nd ref="4" />
			</way>
			<node id="4" lat="1" lon="2" />
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc)
	})

	t.Run("empty relations", func(t *testing.T) {
		xml := `
		<osm>
			<relation id="2">
				<tag k="type" v="multipolygon" />
			</relation>
		</osm>`

		fc := geojson.NewFeatureCollection()
		testConvert(t, xml, fc)
	})
}

func TestConvert_meta(t *testing.T) {
	xml := `
	<osm>
		<node
			id="1"
			lat="1.234" lon="4.321"
			timestamp="2013-01-13T22:56:07Z"
			version="7"
			changeset="1234"
			user="johndoe"
			uid="123" />
	</osm>`

	feature := geojson.NewFeature(orb.Point{4.321, 1.234})
	feature.ID = "node/1"
	feature.Properties["type"] = "node"
	feature.Properties["id"] = 1
	feature.Properties["meta"] = map[string]interface{}{
		"timestamp": "2013-01-13T22:56:07Z",
		"version":   7,
		"changeset": 1234,
		"user":      "johndoe",
		"uid":       123,
	}

	fc := geojson.NewFeatureCollection().Append(feature)
	testConvert(t, xml, fc)
}

func TestConvert_useAugmentedNodeValues(t *testing.T) {
	xml := `
	<osm>
		<way id="1">
			<tag k="foo" v="bar" />
			<nd ref="1" lat="1.0" lon="2.0" />
			<nd ref="2" lat="1.0" lon="2.0" />
			<nd ref="3" lat="1.0" lon="2.0" />
		</way>
	</osm>`

	feature := geojson.NewFeature(orb.LineString{{2, 1}, {2, 1}, {2, 1}})
	feature.ID = "way/1"
	feature.Properties["type"] = "way"
	feature.Properties["id"] = 1
	feature.Properties["tags"] = map[string]string{"foo": "bar"}

	fc := geojson.NewFeatureCollection().Append(feature)
	testConvert(t, xml, fc)
}

func TestBuildRouteLineString(t *testing.T) {
	ctx := &context{
		osm:       &osm.OSM{},
		skippable: map[osm.WayID]struct{}{},
		wayMap: map[osm.WayID]*osm.Way{
			2: {
				ID: 2,
				Nodes: osm.WayNodes{
					{ID: 1, Lat: 1, Lon: 2},
					{ID: 2},
					{ID: 3, Lat: 3, Lon: 4},
				},
			},
		},
	}

	relation := &osm.Relation{
		ID: 1,
		Members: osm.Members{
			{Type: osm.TypeNode, Ref: 1},
			{Type: osm.TypeWay, Ref: 2},
			{Type: osm.TypeWay, Ref: 3},
		},
	}

	feature := ctx.buildRouteLineString(relation)
	if !orb.Equal(feature.Geometry, orb.LineString{{2, 1}, {4, 3}}) {
		t.Errorf("incorrect geometry: %v", feature.Geometry)
	}

	relation = &osm.Relation{
		ID: 1,
		Members: osm.Members{
			{Type: osm.TypeWay, Ref: 20},
			{Type: osm.TypeWay, Ref: 30},
		},
	}

	feature = ctx.buildRouteLineString(relation)
	if feature != nil {
		t.Errorf("should not return feature if no ways present: %v", feature)
	}
}

func testConvert(t *testing.T, rawXML string, expected *geojson.FeatureCollection, opts ...Option) {
	t.Helper()

	o := &osm.OSM{}
	err := xml.Unmarshal([]byte(rawXML), &o)
	if err != nil {
		t.Fatalf("failed to unmarshal xml: %v", err)
	}

	// clean up expected a bit
	for _, f := range expected.Features {
		if f.Properties["tags"] == nil {
			f.Properties["tags"] = map[string]string{}
		}

		if f.Properties["meta"] == nil {
			f.Properties["meta"] = map[string]interface{}{}
		}

		if f.Properties["relations"] == nil {
			f.Properties["relations"] = []*relationSummary{}
		} else {
			for _, rs := range f.Properties["relations"].([]*relationSummary) {
				if rs.Tags == nil {
					rs.Tags = map[string]string{}
				}
			}
		}
	}

	fc, err := Convert(o, opts...)
	if err != nil {
		t.Fatalf("convert error: %v", err)
	}

	raw := jsonLoop(t, fc)
	expectedRaw := jsonLoop(t, expected)

	if !reflect.DeepEqual(raw, expectedRaw) {
		if len(raw.Features) != len(expectedRaw.Features) {
			t.Logf("%v", jsonMarshalIndent(t, raw))
			t.Logf("%v", jsonMarshalIndent(t, expectedRaw))
			t.Errorf("not equal")
		} else {
			for i := range expectedRaw.Features {
				if !reflect.DeepEqual(raw.Features[i], expectedRaw.Features[i]) {
					t.Logf("%v", jsonMarshalIndent(t, raw.Features[i]))
					t.Logf("%v", jsonMarshalIndent(t, expectedRaw.Features[i]))
					t.Errorf("Feature %d not equal", i)
				}
			}
		}
	}
}

type rawFC struct {
	Type     string        `json:"type"`
	Features []interface{} `json:"features"`
}

func jsonLoop(t *testing.T, fc *geojson.FeatureCollection) *rawFC {
	data, err := json.Marshal(fc)
	if err != nil {
		t.Fatalf("unable to marshal fc: %v", err)
	}

	result := &rawFC{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unable to unmarshal data: %v", err)
	}

	return result
}

func jsonMarshalIndent(t *testing.T, raw interface{}) string {
	data, err := json.MarshalIndent(raw, "", " ")
	if err != nil {
		t.Fatalf("unable to marshal json: %v", err)
	}

	return string(data)
}
