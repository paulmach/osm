package osm

import (
	"encoding/json"
	"errors"
	"sort"
	"sync"
)

// UninterestingTags are boring tags. If an element only has
// these tags it does not usually need to be displayed.
// For example, if a node with just these tags is part of a way, it
// probably does not need its own icon along the way.
var UninterestingTags = map[string]bool{
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

// Map returns the tags as a key/value map.
func (ts Tags) Map() map[string]string {
	result := make(map[string]string, len(ts))
	for _, t := range ts {
		result[t.Key] = t.Value
	}

	return result
}

// AnyInteresting will return true if there is at last one interesting tag.
func (ts Tags) AnyInteresting() bool {
	for _, t := range ts {
		if !UninterestingTags[t.Key] {
			return true
		}
	}

	return false
}

// MarshalJSON allows the tags to be marshalled as a key/value object,
// as defined by the overpass osmjson.
func (ts Tags) MarshalJSON() ([]byte, error) {
	return json.Marshal(ts.Map())
}

// UnmarshalJSON allows the tags to be unmarshalled from a key/value object,
// as defined by the overpass osmjson.
func (ts *Tags) UnmarshalJSON(data []byte) error {
	o := make(map[string]string)
	err := json.Unmarshal(data, &o)
	if err != nil {
		return err
	}

	tags := make(Tags, 0, len(o))

	for k, v := range o {
		tags = append(tags, Tag{Key: k, Value: v})
	}

	*ts = tags
	return nil
}

type tagsSort Tags

// SortByKeyValue will do an inplace sort of the tags.
func (ts Tags) SortByKeyValue() {
	sort.Sort(tagsSort(ts))
}
func (ts tagsSort) Len() int      { return len(ts) }
func (ts tagsSort) Swap(i, j int) { ts[i], ts[j] = ts[j], ts[i] }
func (ts tagsSort) Less(i, j int) bool {
	if ts[i].Key == ts[j].Key {
		return ts[i].Value < ts[j].Value
	}

	return ts[i].Key < ts[j].Key
}

func (ts Tags) keyValues(ss *stringSet) (keys, values []uint32) {
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
