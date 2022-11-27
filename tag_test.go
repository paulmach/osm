package osm

import (
	"bytes"
	"reflect"
	"testing"
)

func TestTags_FindTag(t *testing.T) {
	cases := []struct {
		name string
		tags Tags
		key  string
		tag  *Tag
	}{
		{
			name: "find tag",
			tags: Tags{
				{Key: "area", Value: "true"},
				{Key: "building", Value: "yes"},
			},
			key: "building",
			tag: &Tag{Key: "building", Value: "yes"},
		},
		{
			name: "not found",
			tags: Tags{
				{Key: "building", Value: "yes"},
			},
			key: "not found",
			tag: nil,
		},
		{
			name: "empty value",
			tags: Tags{
				{Key: "present", Value: ""},
			},
			key: "present",
			tag: &Tag{Key: "present", Value: ""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.tags.FindTag(tc.key)
			if !reflect.DeepEqual(v, tc.tag) {
				t.Errorf("incorrect find tag: %v != %v", v, tc.tag)
			}
		})
	}
}

func TestTags_HasTag(t *testing.T) {
	cases := []struct {
		name string
		tags Tags
		key  string
		has  bool
	}{
		{
			name: "has tag",
			tags: Tags{
				{Key: "area", Value: "true"},
				{Key: "building", Value: "yes"},
			},
			key: "building",
			has: true,
		},
		{
			name: "not found",
			tags: Tags{
				{Key: "building", Value: "yes"},
			},
			key: "not found",
			has: false,
		},
		{
			name: "empty value",
			tags: Tags{
				{Key: "present", Value: ""},
			},
			key: "present",
			has: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.tags.HasTag(tc.key)
			if v != tc.has {
				t.Errorf("incorrect has tag: %v != %v", v, tc.has)
			}
		})
	}
}

func TestTags_AnyInteresting(t *testing.T) {
	cases := []struct {
		name        string
		tags        Tags
		interesting bool
	}{
		{
			name: "has interesting",
			tags: Tags{
				{Key: "building", Value: "yes"},
			},
			interesting: true,
		},
		{
			name:        "no tags",
			tags:        Tags{},
			interesting: false,
		},
		{
			name: "no interesting tags",
			tags: Tags{
				{Key: "source", Value: "whatever"},
				{Key: "history", Value: "lots"},
			},
			interesting: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.tags.AnyInteresting()
			if v != tc.interesting {
				t.Errorf("incorrect interesting: %v != %v", v, tc.interesting)
			}
		})
	}
}

func TestTags_MarshalJSON(t *testing.T) {
	data, err := Tags{}.MarshalJSON()
	if err != nil {
		t.Errorf("marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`{}`)) {
		t.Errorf("incorrect data, got: %v", string(data))
	}

	t2 := Tags{
		Tag{Key: "highway 🏤 ", Value: "crossing"},
		Tag{Key: "source", Value: "Bind 🏤 "},
	}

	data, err = t2.MarshalJSON()
	if err != nil {
		t.Errorf("marshal error: %v", err)
	}
	if !bytes.Equal(data, []byte(`{"highway 🏤 ":"crossing","source":"Bind 🏤 "}`)) {
		t.Errorf("incorrect data, got: %v", string(data))
	}
}

func TestTags_UnmarshalJSON(t *testing.T) {
	tags := Tags{}
	data := []byte(`{"highway 🏤 ":"crossing","source":"Bind 🏤 "}`)

	err := tags.UnmarshalJSON(data)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	tags.SortByKeyValue()
	t2 := Tags{
		Tag{Key: "highway 🏤 ", Value: "crossing"},
		Tag{Key: "source", Value: "Bind 🏤 "},
	}

	if !reflect.DeepEqual(tags, t2) {
		t.Errorf("incorrect tags: %v", tags)
	}
}

func TestTags_SortByKeyValue(t *testing.T) {
	tags := Tags{
		Tag{Key: "highway", Value: "crossing"},
		Tag{Key: "source", Value: "Bind"},
	}

	tags.SortByKeyValue()
	if v := tags[0].Key; v != "highway" {
		t.Errorf("incorrect sort got %v", v)
	}

	if v := tags[1].Key; v != "source" {
		t.Errorf("incorrect sort got %v", v)
	}

	tags = Tags{
		Tag{Key: "source", Value: "Bind"},
		Tag{Key: "highway", Value: "crossing"},
	}

	tags.SortByKeyValue()
	if v := tags[0].Key; v != "highway" {
		t.Errorf("incorrect sort got %v", v)
	}

	if v := tags[1].Key; v != "source" {
		t.Errorf("incorrect sort got %v", v)
	}
}
