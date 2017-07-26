package osmapi

import (
	"context"
	"os"
	"testing"

	"github.com/paulmach/osm"

	"golang.org/x/time/rate"
)

var _ RateLimiter = &rate.Limiter{}

func TestNodes(t *testing.T) {
	if os.Getenv("LIVE_TEST") != "true" {
		t.Skipf("skipping live test, set LIVE_TEST=true to enable")
	}

	ctx := context.Background()
	nodes, err := Nodes(ctx, []osm.NodeID{2640249171, 2640249172, 2640249173})
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(nodes); l != 3 {
		t.Errorf("incorrect number of nodes, got %d", l)
	}
}

func TestWays(t *testing.T) {
	if os.Getenv("LIVE_TEST") != "true" {
		t.Skipf("skipping live test, set LIVE_TEST=true to enable")
	}

	ctx := context.Background()
	ways, err := Ways(ctx, []osm.WayID{106994776, 106994777, 106994778})
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(ways); l != 3 {
		t.Errorf("incorrect number of ways, got %d", l)
	}
}

func TestRelations(t *testing.T) {
	if os.Getenv("LIVE_TEST") != "true" {
		t.Skipf("skipping live test, set LIVE_TEST=true to enable")
	}

	ctx := context.Background()
	relations, err := Relations(ctx, []osm.RelationID{2714790, 2714791, 2714792})
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(relations); l != 3 {
		t.Errorf("incorrect number of relations, got %d", l)
	}
}

func TestMap(t *testing.T) {
	if os.Getenv("LIVE_TEST") != "true" {
		t.Skipf("skipping live test, set LIVE_TEST=true to enable")
	}

	ctx := context.Background()
	lat, lon := 37.79, -122.27

	b := &osm.Bounds{
		MinLat: lat - 0.001,
		MaxLat: lat + 0.001,
		MinLon: lon - 0.001,
		MaxLon: lon + 0.001,
	}
	o, err := Map(ctx, b)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if len(o.Nodes) == 0 {
		t.Errorf("no nodes returned")
	}

	if len(o.Ways) == 0 {
		t.Errorf("no ways returned")
	}

	if len(o.Relations) == 0 {
		t.Errorf("no relations returned")
	}
}
