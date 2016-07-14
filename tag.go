package osm

import (
	"errors"
	"sync"
)

// Tag is a key+value item attached to osm nodes, ways and relations.
type Tag struct {
	Key   string `xml:"k,attr"`
	Value string `xml:"v,attr"`
}

// Tags is a collection of Tag objects with some helper functions.
type Tags []Tag

// Find will return the value for the key.
// Will return an empty string if not found.
func (ts Tags) Find(k string) string {
	for _, t := range ts {
		if t.Key == k {
			return t.Value
		}
	}

	return ""
}

func (ts Tags) KeyValues(ss *stringSet) (keys, values []uint32) {
	for _, t := range ts {
		keys = append(keys, ss.Add(t.Key))
		values = append(values, ss.Add(t.Value))
	}

	return keys, values
}

func tagsFromStrings(ss []string, keys, values []uint32) (Tags, error) {
	if len(keys) != len(values) {
		return nil, errors.New("keys not same length as values")
	}

	if len(keys) == 0 {
		return nil, nil
	}

	l := uint32(len(ss))
	result := make([]Tag, 0, len(keys))
	for i := range keys {
		if keys[i] >= l {
			return nil, errors.New("key index out of range")
		}

		if values[i] >= l {
			return nil, errors.New("values index out of range")
		}

		result = append(result, Tag{Key: ss[keys[i]], Value: ss[values[i]]})
	}

	return result, nil
}

type stringSet struct {
	lk     sync.Mutex
	values []string
	Set    map[string]uint32
}

func (ss *stringSet) Add(s string) uint32 {
	ss.lk.Lock()
	defer ss.lk.Unlock()

	if ss.Set == nil {
		ss.Set = make(map[string]uint32)

		// zero id is reserved as null for dense nodes packing
		ss.values = make([]string, 1, 100)
	}

	if i, ok := ss.Set[s]; ok {
		return i
	}

	ss.values = append(ss.values, s)
	ss.Set[s] = uint32(len(ss.values)) - 1

	return ss.Set[s]
}

func (ss *stringSet) Strings() []string {
	ss.lk.Lock()
	defer ss.lk.Unlock()

	return ss.values
}
