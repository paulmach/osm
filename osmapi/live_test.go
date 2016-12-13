package osmapi

import (
	"testing"

	osm "github.com/paulmach/go.osm"

	"golang.org/x/net/context"
)

func TestNodes(t *testing.T) {
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
	ctx := context.Background()
	lat, lon := 37.79, -122.27
	o, err := Map(ctx, lon-0.001, lat-0.001, lon+0.001, lat+0.001)
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
