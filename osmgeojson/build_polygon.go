package osmgeojson

import (
	"fmt"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/internal/mputil"
)

func (ctx *context) buildPolygon(relation *osm.Relation) *geojson.Feature {
	tags := relation.Tags.Map()

	var outer []mputil.Segment
	var inner []mputil.Segment

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
			if len(m.Nodes) != 0 {
				way = &osm.Way{
					ID:    osm.WayID(m.Ref),
					Nodes: m.Nodes,
				}
			} else {
				tainted = true
				continue
			}
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
			// we have the way but none of the node members
			continue
		}

		segment := mputil.Segment{
			Orientation: m.Orientation,
			Line:        ls,
		}

		if m.Role == "outer" {
			outerWay = way

			if segment.Orientation == orb.CW {
				segment.Reverse()
			}

			outer = append(outer, segment)
		} else {
			if segment.Orientation == orb.CCW {
				segment.Reverse()
			}

			inner = append(inner, segment)
		}
	}

	var geometry orb.Geometry

	// If there is only one outer way, and the relation doesn't have any interesting tags
	// use the way to define this polygon. ie. use the way's type, id and tags.
	tagObject := osm.Element(relation)

	if len(outer) == 0 && !ctx.includeInvalidPolygons {
		// no outer polygon, skip this relation
		return nil
	} else if len(outer) == 1 && outerCount == 1 {
		// This section handles "old style" multipolygons that don't/shouldn't
		// exist anymore. In the past tags were set on the outer ring way and
		// the relation was used to add holes to the way.
		outerRing := mputil.MultiSegment(outer).Ring(orb.CCW)
		if len(outerRing) < 4 || !outerRing.Closed() {
			// not a valid outer ring
			return nil
		}

		innerSections := mputil.Join(inner)
		polygon := make(orb.Polygon, 0, len(inner)+1)

		polygon = append(polygon, outerRing)
		for _, is := range innerSections {
			polygon = append(polygon, is.Ring(orb.CW))
		}

		geometry = polygon

		if !hasInterestingTags(relation.Tags, map[string]string{"type": "true"}) {
			ctx.skippable[outerWay.ID] = struct{}{}

			tags = outerWay.Tags.Map()
			tagObject = outerWay
		}
	} else {
		// more than one outer, need to map inner polygons to
		// the outer that contains them.
		outerSections := mputil.Join(outer)

		mp := make(orb.MultiPolygon, 0, len(outer))
		for _, os := range outerSections {
			ring := os.Ring(orb.CCW)
			if !ctx.includeInvalidPolygons && (len(ring) < 4 || !ring.Closed()) {
				// needs at least 4 points and matching endpoints
				continue
			}

			mp = append(mp, orb.Polygon{ring})
		}

		if len(mp) == 0 && !ctx.includeInvalidPolygons {
			// no valid outer ways.
			return nil
		}

		innerSections := mputil.Join(inner)
		for _, is := range innerSections {
			ring := is.Ring(orb.CW)
			mp = addToMultiPolygon(mp, ring, ctx.includeInvalidPolygons)
		}

		if len(mp) == 0 {
			return nil
		}

		geometry = mp
		if len(mp) == 1 {
			geometry = mp[0]
		}
	}

	featureID := tagObject.FeatureID()
	f := geojson.NewFeature(geometry)

	if !ctx.noID {
		f.ID = fmt.Sprintf("%s/%d", featureID.Type(), featureID.Ref())
	}
	f.Properties["id"] = int(featureID.Ref())
	f.Properties["type"] = string(featureID.Type())

	if tainted {
		f.Properties["tainted"] = true
	}

	f.Properties["tags"] = tags
	ctx.addMetaProperties(f.Properties, tagObject)

	return f
}

func addToMultiPolygon(mp orb.MultiPolygon, ring orb.Ring, includeInvalidPolygons bool) orb.MultiPolygon {
	for i := range mp {
		if polygonContains(mp[i][0], ring) {
			mp[i] = append(mp[i], ring)
			return mp
		}
	}

	if !includeInvalidPolygons {
		// inner without its outer
		return mp
	}

	if len(mp) > 0 {
		// if the outer ring of the first polygon is not closed,
		// we don't really know if this inner should be part of it.
		// But... we assume yes.
		fr := mp[0][0]
		if len(fr) != 0 && fr[0] != fr[len(fr)-1] {
			mp[0] = append(mp[0], ring)
			return mp
		}

		// trying to find an existing "without outer" polygon to add this to.
		for i := range mp {
			if len(mp[i][0]) == 0 {
				mp[i] = append(mp[i], ring)
				return mp
			}
		}
	}

	// no polygons with empty outer, so create one.
	// create another polygon with empty outer.
	return append(mp, orb.Polygon{nil, ring})
}

func polygonContains(outer orb.Ring, r orb.Ring) bool {
	for _, p := range r {
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

func reorient(p orb.Polygon) {
	if p[0].Orientation() != orb.CCW {
		p[0].Reverse()
	}

	for i := 1; i < len(p); i++ {
		if p[i].Orientation() != orb.CW {
			p[i].Reverse()
		}
	}
}
