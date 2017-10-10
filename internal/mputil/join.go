package mputil

// Join will join a set of segments into a set of connected MultiSegments.
func Join(segments []Segment) []MultiSegment {
	lists := []MultiSegment{}
	segments = compact(segments)

	// matches are removed from `segments` and put into the current
	// group, so when `segments` is empty we're done.
	for len(segments) != 0 {
		current := MultiSegment{segments[len(segments)-1]}
		segments = segments[:len(segments)-1]

		// if the current group is a ring, we're done.
		// else add in all the lines.
		for len(segments) != 0 && !current.First().Equal(current.Last()) {
			first := current.First()
			last := current.Last()

			foundAt := -1
			for i, segment := range segments {
				if last.Equal(segment.First()) {
					// nice fit at the end of current

					segment.Line = segment.Line[1:]
					current = append(current, segment)
					foundAt = i
					break
				} else if last.Equal(segment.Last()) {
					// reverse it and it'll fit at the end
					segment.Reverse()

					segment.Line = segment.Line[1:]
					current = append(current, segment)
					foundAt = i
					break
				} else if first.Equal(segment.Last()) {
					// nice fit at the start of current
					segment.Line = segment.Line[:len(segment.Line)-1]
					current = append(MultiSegment{segment}, current...)

					foundAt = i
					break
				} else if first.Equal(segment.First()) {
					// reverse it and it'll fit at the start
					segment.Reverse()

					segment.Line = segment.Line[:len(segment.Line)-1]
					current = append(MultiSegment{segment}, current...)

					foundAt = i
					break
				}
			}

			if foundAt == -1 {
				break // Invalid geometry (dangling way, unclosed ring)
			}

			// remove the found/matched segment from the list.
			if foundAt < len(segments)/2 {
				// first half, shift up
				for i := foundAt; i > 0; i-- {
					segments[i] = segments[i-1]
				}
				segments = segments[1:]
			} else {
				// second half, shift down
				for i := foundAt + 1; i < len(segments); i++ {
					segments[i-1] = segments[i]
				}
				segments = segments[:len(segments)-1]
			}
		}

		lists = append(lists, current)
	}

	return lists
}

func compact(ms MultiSegment) MultiSegment {
	at := 0
	for _, s := range ms {
		if len(s.Line) <= 1 {
			continue
		}

		ms[at] = s
		at++
	}

	return ms[:at]
}
