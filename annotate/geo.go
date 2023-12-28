package annotate

import (
	"math"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/internal/mputil"
)

func wayPointOnSurface(w *osm.Way) orb.Point {
	centroid := wayCentroid(w)

	// find closest node to centroid.
	// This is how ST_PointOnSurface is implemented.
	min := math.MaxFloat64
	index := 0
	for i, n := range w.Nodes {
		d := geo.Distance(centroid, n.Point())
		if d < min {
			index = i
			min = d
		}
	}

	return w.Nodes[index].Point()
}

func wayCentroid(w *osm.Way) orb.Point {
	dist := 0.0
	point := orb.Point{}

	seg := [2]orb.Point{}

	for i := 0; i < len(w.Nodes)-1; i++ {
		seg[0] = w.Nodes[i].Point()
		seg[1] = w.Nodes[i+1].Point()

		d := geo.Distance(seg[0], seg[1])

		point[0] += (seg[0][0] + seg[1][0]) / 2.0 * d
		point[1] += (seg[0][1] + seg[1][1]) / 2.0 * d

		dist += d
	}

	point[0] /= dist
	point[1] /= dist

	return point
}

// orientation will annotate the orientation of multipolygon relation members.
// This makes it possible to reconstruct relations with partial data in the right direction.
// Return value indicates if the result is 'tainted', e.g. not all way members were present.
func orientation(members osm.Members, ways map[osm.WayID]*osm.Way, at time.Time) bool {
	outer, inner, tainted := mputil.Group(members, ways, at)

	outers := mputil.Join(outer)
	inners := mputil.Join(inner)

	for _, outer := range outers {
		annotateOrientation(members, outer, orb.CCW)
	}

	for _, inner := range inners {
		annotateOrientation(members, inner, orb.CW)
	}

	return tainted
}

func annotateOrientation(members osm.Members, ms mputil.MultiSegment, o orb.Orientation) {
	factor := orb.Orientation(1)
	if ms.Orientation() != o {
		factor = -1
	}

	for _, segment := range ms {
		if segment.Reversed {
			members[segment.Index].Orientation = -1 * factor * o
		} else {
			members[segment.Index].Orientation = factor * o
		}
	}
}
