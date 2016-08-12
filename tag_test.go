package osm

import "testing"

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
