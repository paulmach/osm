package annotate

import (
	"math"

	"github.com/paulmach/orb/geo"
	"github.com/paulmach/osm"
)

func wayPointOnSurface(w *osm.Way) geo.Point {
	centroid := wayCentroid(w)

	// find closest node to centroid.
	// This is how ST_PointOnSurface is implemented.
	min := math.MaxFloat64
	index := 0
	for i, n := range w.Nodes {
		d := centroid.DistanceFrom(n.Point())
		if d < min {
			index = i
			min = d
		}
	}

	return w.Nodes[index].Point()
}

func wayCentroid(w *osm.Way) geo.Point {
	dist := 0.0
	point := geo.Point{}

	seg := [2]geo.Point{}

	for i := 0; i < len(w.Nodes)-1; i++ {
		seg[0] = w.Nodes[i].Point()
		seg[1] = w.Nodes[i+1].Point()

		d := seg[0].DistanceFrom(seg[1])

		point[0] += (seg[0][0] + seg[1][0]) / 2.0 * d
		point[1] += (seg[0][1] + seg[1][1]) / 2.0 * d

		dist += d
	}

	point[0] /= dist
	point[1] /= dist

	return point
}
