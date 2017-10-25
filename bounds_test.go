package osm

import (
	"fmt"
	"testing"

	"github.com/paulmach/orb/maptile"
)

func TestNewBoundFromTile(t *testing.T) {
	bounds, _ := NewBoundsFromTile(maptile.New(7, 8, 9))

	// check 9 tiles around bounds
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			t.Run(fmt.Sprintf("i %d j %d", i, j), func(t *testing.T) {
				n := centroid(mustBounds(t, uint32(7+i), uint32(8+j), 9))
				if i == 0 && j == 0 {
					if !bounds.ContainsNode(n) {
						t.Errorf("should contain point")
					}
				} else {
					if bounds.ContainsNode(n) {
						t.Errorf("should not contain point")
					}
				}
			})
		}
	}
}

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

func mustBounds(t *testing.T, x, y uint32, z maptile.Zoom) *Bounds {
	bounds, err := NewBoundsFromTile(maptile.New(x, y, z))
	if err != nil {
		t.Fatalf("invalid bounds: %v", err)
	}

	return bounds
}

func centroid(b *Bounds) *Node {
	return &Node{
		Lon: (b.MinLon + b.MaxLon) / 2,
		Lat: (b.MinLat + b.MaxLat) / 2,
	}
}
