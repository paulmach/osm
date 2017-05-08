package osmgeojson

import "github.com/paulmach/orb/geo"

// joinLineStrings will merge line that share the same endpoints.
// It will reverse lines if necessary.
func joinLineStrings(lines []geo.LineString) []geo.LineString {
	groups := []geo.MultiLineString{}

	// matches are removed from `lines` and put into the current
	// group, so when `lines` is empty we're done.
	for len(lines) != 0 {
		current := geo.MultiLineString{lines[len(lines)-1]}
		lines = lines[:len(lines)-1]

		// if the current group is a ring, we're done.
		// else add in all the lines.
		for len(lines) != 0 && !first(current).Equal(last(current)) {
			first := first(current)
			last := last(current)

			foundAt := -1
			for i, line := range lines {
				if last.Equal(line[0]) {
					// nice fit at the end of current
					current = append(current, line[1:])
					foundAt = i
					break
				} else if last.Equal(line[len(line)-1]) {
					// reverse it and it'll fit at the end
					line.Reverse()
					current = append(current, line[1:])
					foundAt = i
					break
				} else if first.Equal(line[len(line)-1]) {
					// nice fit at the start of current
					line = line[:len(line)-1]
					current = append(geo.MultiLineString{line}, current...)

					foundAt = i
					break
				} else if first.Equal(line[0]) {
					// reverse it and it'll fit at the start
					line.Reverse()

					line = line[:len(line)-1]
					current = append(geo.MultiLineString{line}, current...)

					foundAt = i
					break
				}
			}

			if foundAt == -1 {
				break // Invalid geometry (dangling way, unclosed ring)
			}

			// remove the found/matched line from the list.
			if foundAt < len(lines)/2 {
				// first half, shift up
				for i := foundAt; i > 0; i-- {
					lines[i] = lines[i-1]
				}
				lines = lines[1:]
			} else {
				// second half, shift down
				for i := foundAt + 1; i < len(lines); i++ {
					lines[i-1] = lines[i]
				}
				lines = lines[:len(lines)-1]
			}
		}

		groups = append(groups, current)
	}

	// merge the groups (a multi line string) into a single line string.
	result := make([]geo.LineString, len(groups))
	for i, group := range groups {
		result[i] = merge(group)
	}

	return result
}

func merge(ml geo.MultiLineString) geo.LineString {
	length := 0
	for _, l := range ml {
		length += len(l)
	}

	full := make(geo.LineString, length)

	at := 0
	for _, l := range ml {
		copy(full[at:], l)
		at += len(l)
	}

	return full
}

func first(ml geo.MultiLineString) geo.Point {
	return ml[0][0]
}

func last(ml geo.MultiLineString) geo.Point {
	l := ml[len(ml)-1]
	return l[len(l)-1]
}
