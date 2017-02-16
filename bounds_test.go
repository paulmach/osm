package osm

import "testing"

func TestBoundsContainsNode(t *testing.T) {
	b := &Bounds{}

	if v := b.ContainsNode(&Node{}); !v {
		t.Errorf("should contain node on boundary")
	}

	if v := b.ContainsNode(&Node{Lat: -1}); v {
		t.Errorf("should not contain node outside bounds")
	}
	if v := b.ContainsNode(&Node{Lat: 1}); v {
		t.Errorf("should not contain node outside bounds")
	}
	if v := b.ContainsNode(&Node{Lon: -1}); v {
		t.Errorf("should not contain node outside bounds")
	}

	if v := b.ContainsNode(&Node{Lon: 1}); v {
		t.Errorf("should not contain node outside bounds")
	}
}
