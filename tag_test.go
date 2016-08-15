package osm

import (
	"reflect"
	"testing"
)

func TestTagsMarshal(t *testing.T) {
	data, err := Tags{}.Marshal()
	if err != nil {
		t.Errorf("marshal error: %v", err)
	}

	if data != nil {
		t.Errorf("empty tags should results in nil byte array, got %v", data)
	}

	var t1 Tags
	data, err = t1.Marshal()
	if err != nil {
		t.Errorf("marshal error: %v", err)
	}

	if data != nil {
		t.Errorf("empty tags should results in nil byte array, got %v", data)
	}

	t2 := Tags{
		Tag{Key: "highway", Value: "crossing"},
		Tag{Key: "source", Value: "Bind"},
	}

	data, err = t2.Marshal()
	if err != nil {
		t.Errorf("marshal error: %v", err)
	}

	t3, err := UnmarshalTags(data)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(t2, t3) {
		t.Errorf("unequal sets")
		t.Logf("%v", t2)
		t.Logf("%v", t3)
	}

	t4, err := UnmarshalTags(nil)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	if t4 != nil {
		t.Errorf("should get nil tags for nil data, got %v", t4)
	}
}

func TestTagsSortByKeyValue(t *testing.T) {
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
