package osm

import (
	"encoding/json"
	"sort"
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
	Key   string `xml:"k,attr" json:"k"`
	Value string `xml:"v,attr" json:"v,omitempty"`
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

// FindTag will return the Tag for the given key.
// Can be used to determine if a key exists, even with an empty value.
// Returns nil if not found.
func (ts Tags) FindTag(k string) *Tag {
	for _, t := range ts {
		if t.Key == k {
			return &t
		}
	}

	return nil
}

// HasTag will return the true if a tag exists for the given key.
func (ts Tags) HasTag(k string) bool {
	for _, t := range ts {
		if t.Key == k {
			return true
		}
	}

	return false
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
	return marshalJSON(ts.Map())
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
