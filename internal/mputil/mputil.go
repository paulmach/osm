package mputil

import (
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/osm"
)

// Segment is a section of a multipolygon with some extra information
// on the member it came from.
type Segment struct {
	Index       uint32
	Orientation orb.Orientation
	Reversed    bool
	Line        orb.LineString
}

// Reverse will reverse the line string of the segment.
func (s *Segment) Reverse() {
	s.Reversed = !s.Reversed
	s.Line.Reverse()
}

// First returns the first point in the segment linestring.
func (s Segment) First() orb.Point {
	return s.Line[0]
}

// Last returns the last point in the segment linestring.
func (s Segment) Last() orb.Point {
	return s.Line[len(s.Line)-1]
}

// MultiSegment is an ordered set of segments that form a continuous
// section of a multipolygon.
type MultiSegment []Segment

// First returns the first point in the list of linestrings.
func (ms MultiSegment) First() orb.Point {
	return ms[0].Line[0]
}

// Last returns the last point in the list of linestrings.
func (ms MultiSegment) Last() orb.Point {
	line := ms[len(ms)-1].Line
	return line[len(line)-1]
}

// ToLineString converts a multisegment into a geo linestring object.
func (ms MultiSegment) ToLineString() orb.LineString {
	length := 0
	for _, s := range ms {
		length += len(s.Line)
	}

	line := make(orb.LineString, 0, length)
	for _, s := range ms {
		line = append(line, s.Line...)
	}

	return line
}

// ToRing converts the multisegment to a ring of the given orientation.
// It uses the orientation on the members if possible.
func (ms MultiSegment) ToRing(o orb.Orientation) orb.Ring {
	length := 0
	for _, s := range ms {
		length += len(s.Line)
	}

	ring := make(orb.Ring, 0, length)

	haveOrient := false
	reversed := false
	for _, s := range ms {
		if s.Orientation != 0 {
			haveOrient = true

			// if s.Orientation == o && s.Reversed {
			// 	reversed = true
			// }
			// if s.Orientation != 0 && !s.Reversed {
			// 	reversed = true
			// }

			if (s.Orientation == o) == s.Reversed {
				reversed = true
			}
		}

		ring = append(ring, s.Line...)
	}

	if (haveOrient && reversed) || (!haveOrient && ring.Orientation() != o) {
		ring.Reverse()
	}

	return ring
}

// Orientation computes the orientation of a multisegment like if it was ring.
func (ms MultiSegment) Orientation() orb.Orientation {
	area := 0.0
	prev := ms.First()

	// implicitly move everything to near the origin to help with roundoff
	offset := prev
	for _, segment := range ms {
		for _, point := range segment.Line {
			area += (prev[0]-offset[0])*(point[1]-offset[1]) -
				(point[0]-offset[0])*(prev[1]-offset[1])

			prev = point
		}
	}

	if area > 0 {
		return orb.CCW
	}

	return orb.CW
}

// Group will take the members and group them by inner our outer parts
// of the relation. Will also build the way geometry.
func Group(
	members osm.Members,
	ways map[osm.WayID]*osm.Way,
	at time.Time,
) (outer, inner []Segment, tainted bool) {
	for i, m := range members {
		if m.Type != osm.TypeWay {
			continue
		}

		w := ways[osm.WayID(m.Ref)]
		if w == nil {
			tainted = true
			continue // could be not found error, or something else.
		}

		line := w.LineStringAt(at)
		if len(line) != len(w.Nodes) {
			tainted = true
		}

		l := Segment{
			Index:       uint32(i),
			Orientation: m.Orientation,
			Reversed:    false,
			Line:        line,
		}

		if m.Role == "outer" {
			if l.Orientation == orb.CW {
				l.Reverse()
			}
			outer = append(outer, l)
		} else if m.Role == "inner" {
			if l.Orientation == orb.CCW {
				l.Reverse()
			}
			inner = append(inner, l)
		}
	}

	return outer, inner, tainted
}
