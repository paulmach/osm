package osmgeojson

import (
	"fmt"

	osm "github.com/paulmach/go.osm"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/geo/geojson"
)

var uninterestingTags = map[string]bool{
	"source":            true,
	"source_ref":        true,
	"source:ref":        true,
	"history":           true,
	"attribution":       true,
	"created_by":        true,
	"tiger:county":      true,
	"tiger:tlid":        true,
	"tiger:upload_uuid": true,
}

type context struct {
	noID                 bool
	noMeta               bool
	noRelationMembership bool

	osm       *osm.OSM
	skippable map[osm.WayID]struct{}

	relationMember map[osm.FeatureID][]*relationSummary
	wayMember      map[osm.NodeID]struct{}
	nodeMap        map[osm.NodeID]*osm.Node
	wayMap         map[osm.WayID]*osm.Way
}

type relationSummary struct {
	ID   osm.RelationID    `json:"id"`
	Role string            `json:"role"`
	Tags map[string]string `json:"tags"`
}

// Convert takes a set of osm elements and converts them
// to a geojson feature collection.
func Convert(o *osm.OSM, opts ...Option) (*geojson.FeatureCollection, error) {
	ctx := &context{
		osm:       o,
		skippable: make(map[osm.WayID]struct{}),
	}

	for _, opt := range opts {
		if err := opt(ctx); err != nil {
			return nil, err
		}
	}

	ctx.wayMap = make(map[osm.WayID]*osm.Way, len(o.Ways))
	for _, w := range ctx.osm.Ways {
		ctx.wayMap[w.ID] = w
	}

	ctx.wayMember = make(map[osm.NodeID]struct{}, len(ctx.osm.Nodes))
	for _, w := range ctx.osm.Ways {
		for i := range w.Nodes {
			ctx.wayMember[w.Nodes[i].ID] = struct{}{}
		}
	}

	// figure out relation membership map
	ctx.relationMember = make(map[osm.FeatureID][]*relationSummary)
	for _, relation := range ctx.osm.Relations {
		var tags map[string]string
		for _, m := range relation.Members {
			if ctx.noRelationMembership && m.Type != osm.TypeNode {
				// If we don't need to do relation membership we only
				// need this for nodes to check if they're interesting.
				continue
			}

			if m.Type == osm.TypeWay {
				// We only need to store the way membership for ways that are
				// present. eg. relations could have thousands of members but only
				// a few in set of osm.
				if _, ok := ctx.wayMap[osm.WayID(m.Ref)]; !ok {
					continue
				}
			}

			if tags == nil {
				tags = relation.Tags.Map()
			}

			fid := m.FeatureID()
			ctx.relationMember[fid] = append(ctx.relationMember[fid], &relationSummary{
				ID:   relation.ID,
				Role: m.Role,
				Tags: tags,
			})
		}
	}

	features := make([]*geojson.Feature, 0, len(ctx.osm.Relations)+len(ctx.osm.Ways))

	// relations
	for _, relation := range ctx.osm.Relations {
		tt := relation.Tags.Find("type")
		if tt == "route" {
			feature := ctx.buildRouteLineString(relation)
			if feature != nil {
				features = append(features, feature)
			}
		} else if tt == "multipolygon" || tt == "boundary" {
			feature := ctx.buildPolygon(relation)
			if feature != nil {
				features = append(features, feature)
			}
		}

		// NOTE: we skip/ignore relation that aren't multipolygons, boundaries or routes
	}

	for _, way := range ctx.osm.Ways {
		// should skip only skippable relation members
		if _, skip := ctx.skippable[way.ID]; skip {
			continue
		}

		feature := ctx.wayToFeature(way)
		if feature != nil {
			features = append(features, feature)
		}
	}

	for _, node := range ctx.osm.Nodes {
		// should NOT skip if any are true:
		//   not a member of a way.
		//   a member of a relation member
		//   has any interesting tags
		// should skip if all are true:
		//   a member of a way.
		//   not a member of a relation member
		//   does not have any interesting tags
		if _, ok := ctx.wayMember[node.ID]; ok &&
			len(ctx.relationMember[node.FeatureID()]) == 0 &&
			!hasInterestingTags(node.Tags, nil) {
			continue
		}

		feature := ctx.nodeToFeature(node)
		if feature != nil {
			features = append(features, feature)
		}
	}

	fc := geojson.NewFeatureCollection()
	fc.Features = features

	return fc, nil
}

// getNode will find the node in the set.
// This allows to lazily create the node map only if
// the nodes+ways aren't augmented (ie. include the lat/lon on them).
func (ctx *context) getNode(id osm.NodeID) *osm.Node {
	if ctx.nodeMap == nil {
		ctx.nodeMap = make(map[osm.NodeID]*osm.Node, len(ctx.osm.Nodes))
		for _, n := range ctx.osm.Nodes {
			ctx.nodeMap[n.ID] = n
		}
	}

	return ctx.nodeMap[id]
}

func (ctx *context) nodeToFeature(n *osm.Node) *geojson.Feature {
	// our definition of empty, ill defined
	if n.Lon == 0 && n.Lat == 0 && n.Version == 0 {
		return nil
	}

	f := geojson.NewFeature(geo.NewPoint(n.Lon, n.Lat))

	if !ctx.noID {
		f.ID = fmt.Sprintf("node/%d", n.ID)
	}
	f.Properties["id"] = int(n.ID)
	f.Properties["type"] = "node"
	f.Properties["tags"] = n.Tags.Map()

	ctx.addMetaProperties(f.Properties, n)

	return f
}

func (ctx *context) wayToLineString(w *osm.Way) (geo.LineString, bool) {
	ls := make(geo.LineString, 0, len(w.Nodes))
	tainted := false
	for _, wn := range w.Nodes {
		if wn.Lon != 0 || wn.Lat != 0 {
			ls = append(ls, geo.NewPoint(wn.Lon, wn.Lat))
		} else if n := ctx.getNode(wn.ID); n != nil {
			ls = append(ls, geo.NewPoint(n.Lon, n.Lat))
		} else {
			tainted = true
		}
	}

	return ls, tainted
}

func (ctx *context) wayToFeature(w *osm.Way) *geojson.Feature {
	ls, tainted := ctx.wayToLineString(w)
	if len(ls) <= 1 {
		// one node ways are ignored.
		return nil
	}

	var f *geojson.Feature
	if w.Polygon() {
		p := geo.Polygon{toRing(ls)}
		reorient(p)
		f = geojson.NewFeature(p)
	} else {
		f = geojson.NewFeature(ls)
	}

	if !ctx.noID {
		f.ID = fmt.Sprintf("way/%d", w.ID)
	}
	f.Properties["id"] = int(w.ID)
	f.Properties["type"] = "way"
	f.Properties["tags"] = w.Tags.Map()

	if tainted {
		f.Properties["tainted"] = true
	}

	ctx.addMetaProperties(f.Properties, w)

	return f
}

func (ctx *context) buildRouteLineString(relation *osm.Relation) *geojson.Feature {
	lines := make([]geo.LineString, 0, 10)
	tainted := false
	for _, m := range relation.Members {
		if m.Type != osm.TypeWay {
			continue
		}

		way := ctx.wayMap[osm.WayID(m.Ref)]
		if way == nil {
			tainted = true
			continue
		}

		if !hasInterestingTags(way.Tags, nil) {
			ctx.skippable[way.ID] = struct{}{}
		}

		ls, t := ctx.wayToLineString(way)
		if t {
			tainted = true
		}

		if len(ls) == 0 {
			continue
		}

		lines = append(lines, ls)
	}

	if len(lines) == 0 {
		// route relation is here, but we don't have any of the way members?
		// TODO: what to do about this?
		return nil
	}

	lines = joinLineStrings(lines)

	var geometry geo.Geometry
	if len(lines) == 1 {
		geometry = lines[0]
	} else {
		geometry = geo.MultiLineString(lines)
	}

	f := geojson.NewFeature(geometry)
	if !ctx.noID {
		f.ID = fmt.Sprintf("relation/%d", relation.ID)
	}

	f.Properties["id"] = int(relation.ID)
	f.Properties["type"] = "relation"

	if tainted {
		f.Properties["tainted"] = true
	}

	f.Properties["tags"] = relation.Tags.Map()
	ctx.addMetaProperties(f.Properties, relation)

	return f
}

func (ctx *context) buildPolygon(relation *osm.Relation) *geojson.Feature {
	tags := relation.Tags.Map()

	var outer []geo.LineString
	var inner []geo.LineString

	tainted := false
	outerCount := 0

	var outerWay *osm.Way // used to get featureID if only one outer way
	for _, m := range relation.Members {
		if m.Type != osm.TypeWay {
			continue
		}

		if m.Role != "inner" && m.Role != "outer" {
			continue
		}

		if m.Role == "outer" {
			outerCount++
		}

		way := ctx.wayMap[osm.WayID(m.Ref)]
		if way == nil {
			tainted = true
			continue
		}

		if m.Role == "outer" {
			if !hasInterestingTags(way.Tags, tags) {
				ctx.skippable[way.ID] = struct{}{}
			}
		} else {
			if !hasInterestingTags(way.Tags, nil) {
				ctx.skippable[way.ID] = struct{}{}
			}
		}

		ls, t := ctx.wayToLineString(way)
		if t {
			tainted = true
		}

		if len(ls) == 0 {
			// we have the way but none the the node members
			continue
		}

		if m.Role == "outer" {
			outerWay = way
			outer = append(outer, ls)
		} else {
			inner = append(inner, ls)
		}
	}

	var geometry geo.Geometry

	// If there is only one outer way, and the relation doesn't have any interesting tags
	// use the way to define this polygon. ie. use the way's type, id and tags.
	tagObject := osm.Element(relation)

	if len(outer) == 0 {
		// no outer polygon, skip this relation
		return nil
	} else if len(outer) == 1 && outerCount == 1 {
		// This section handles "old style" multipolygons that don't/shouldn't
		// exist anymore. In the past tags were set on the outer ring way and
		// the relation was used to add holes to the way.
		outerRing := toRing(outer[0])
		if !outerRing.Valid() {
			// at least 4 points and first and last are the same.
			return nil
		}

		inner = joinLineStrings(inner)
		polygon := make(geo.Polygon, 0, len(inner)+1)

		polygon = append(polygon, outerRing)
		for _, ip := range inner {
			polygon = append(polygon, toRing(ip))
		}

		reorient(polygon) // so it follows the right hand rule.
		geometry = polygon

		if !hasInterestingTags(relation.Tags, map[string]string{"type": "true"}) {
			ctx.skippable[outerWay.ID] = struct{}{}

			tags = outerWay.Tags.Map()
			tagObject = outerWay
		}
	} else {
		// more than one outer, need to map inner polygons to
		// the outer that contians them.
		outer = joinLineStrings(outer)

		mp := make(geo.MultiPolygon, 0, len(outer))
		for _, ls := range outer {
			ring := toRing(ls)
			if !ring.Valid() {
				// needs at least 4 points.
				continue
			}

			mp = append(mp, geo.Polygon{ring})
		}

		if len(mp) == 0 {
			// no valid outer ways.
			return nil
		}

		inner = joinLineStrings(inner)
		for _, ls := range inner {
			addToMultiPolygon(mp, ls)
		}

		// reorient to follow the right hand rule.
		for _, p := range mp {
			reorient(p)
		}

		geometry = mp
		if len(mp) == 1 {
			geometry = mp[0]
		}
	}

	featureID := tagObject.FeatureID()
	f := geojson.NewFeature(geometry)

	if !ctx.noID {
		f.ID = fmt.Sprintf("%s/%d", featureID.Type, featureID.Ref)
	}
	f.Properties["id"] = int(featureID.Ref)
	f.Properties["type"] = string(featureID.Type)

	if tainted {
		f.Properties["tainted"] = true
	}

	f.Properties["tags"] = tags
	ctx.addMetaProperties(f.Properties, tagObject)

	return f
}

func (ctx *context) addMetaProperties(props geojson.Properties, e osm.Element) {
	if !ctx.noRelationMembership {
		relations := ctx.relationMember[e.FeatureID()]
		if len(relations) != 0 {
			props["relations"] = relations
		} else {
			props["relations"] = []*relationSummary{}
		}
	}

	if ctx.noMeta {
		return
	}

	meta := make(map[string]interface{}, 5)
	switch e := e.(type) {
	case *osm.Node:
		if !e.Timestamp.IsZero() {
			meta["timestamp"] = e.Timestamp
		}

		if e.Version != 0 {
			meta["version"] = e.Version
		}

		if e.ChangesetID != 0 {
			meta["changeset"] = e.ChangesetID
		}

		if e.User != "" {
			meta["user"] = e.User
		}

		if e.UserID != 0 {
			meta["uid"] = e.UserID
		}

	case *osm.Way:
		if !e.Timestamp.IsZero() {
			meta["timestamp"] = e.Timestamp
		}

		if e.Version != 0 {
			meta["version"] = e.Version
		}

		if e.ChangesetID != 0 {
			meta["changeset"] = e.ChangesetID
		}

		if e.User != "" {
			meta["user"] = e.User
		}

		if e.UserID != 0 {
			meta["uid"] = e.UserID
		}

	case *osm.Relation:
		if !e.Timestamp.IsZero() {
			meta["timestamp"] = e.Timestamp
		}

		if e.Version != 0 {
			meta["version"] = e.Version
		}

		if e.ChangesetID != 0 {
			meta["changeset"] = e.ChangesetID
		}

		if e.User != "" {
			meta["user"] = e.User
		}

		if e.UserID != 0 {
			meta["uid"] = e.UserID
		}

	default:
		panic("unsupported type")
	}

	props["meta"] = meta
}

func hasInterestingTags(tags osm.Tags, ignore map[string]string) bool {
	if len(tags) == 0 {
		return false
	}

	for _, tag := range tags {
		k, v := tag.Key, tag.Value
		if !uninterestingTags[k] &&
			(ignore == nil || !(ignore[k] == "true" || ignore[k] == v)) {
			return true
		}
	}

	return false
}

func addToMultiPolygon(mp geo.MultiPolygon, ls geo.LineString) {
	for i := range mp {
		if polygonContains(mp[i][0], ls) {
			mp[i] = append(mp[i], toRing(ls))
			return
		}
	}

	// TODO: inner without an outer??
}

func polygonContains(outer geo.Ring, ls geo.LineString) bool {
	for _, p := range ls {
		inside := false

		x, y := p[0], p[1]
		i, j := 0, len(outer)-1
		for i < len(outer) {
			xi, yi := outer[i][0], outer[i][1]
			xj, yj := outer[j][0], outer[j][1]

			if ((yi > y) != (yj > y)) &&
				(x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
				inside = !inside
			}

			j = i
			i++
		}

		if inside {
			return true
		}
	}

	return false
}

func reorient(p geo.Polygon) {
	if p[0].Orientation() != orb.CCW {
		p[0].Reverse()
	}

	for i := 1; i < len(p); i++ {
		if p[i].Orientation() != orb.CW {
			p[i].Reverse()
		}
	}
}

func toRing(ls geo.LineString) geo.Ring {
	if len(ls) < 2 {
		return geo.Ring(ls)
	}

	// duplicate last point
	if ls[0] != ls[len(ls)-1] {
		return geo.Ring(append(ls, ls[0]))
	}

	return geo.Ring(ls)
}
