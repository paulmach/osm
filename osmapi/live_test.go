package osmapi

import (
	"context"
	"os"
	"testing"

	"github.com/onXmaps/osm"

	"golang.org/x/time/rate"
)

var _ RateLimiter = &rate.Limiter{}

func liveOnly(t testing.TB) {
	if os.Getenv("LIVE_TEST") != "true" {
		t.Skipf("skipping live test, set LIVE_TEST=true to enable")
	}
}

func TestNode(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	node, err := Node(ctx, 2640249171)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if node.ID != 2640249171 {
		t.Errorf("incorrect node: %v", node)
	}
}

func TestNodes(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	nodes, err := Nodes(ctx, []osm.NodeID{2640249171, 2640249172, 2640249173})
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(nodes); l != 3 {
		t.Errorf("incorrect number of nodes: %d", l)
	}
}

func TestNodeVersion(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	node, err := NodeVersion(ctx, 2640249171, 3)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if node.ID != 2640249171 {
		t.Errorf("incorrect node: %v", node)
	}

	if node.Version != 3 {
		t.Errorf("incorrect node version: %v", node)
	}
}

func TestNodeHistory(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	nodes, err := NodeHistory(ctx, 2640249171)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(nodes); l != 3 {
		t.Errorf("incorrect number of nodes: %v", l)
	}
}

func TestNodeWays(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	ways, err := NodeWays(ctx, 2640249171)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(ways); l != 1 {
		t.Errorf("should be part of 1 way: %v", l)
	}
}

func TestNodeRelations(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	relations, err := NodeRelations(ctx, 2640249171)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if len(relations) != 2 {
		t.Errorf("should be part of 2 relations: %v", relations)
	}
}

func TestWay(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	way, err := Way(ctx, 106994776)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if way.ID != 106994776 {
		t.Errorf("incorrect way version: %v", way)
	}
}

func TestWays(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	ways, err := Ways(ctx, []osm.WayID{106994776, 106994777, 106994778})
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(ways); l != 3 {
		t.Errorf("incorrect number of ways: %d", l)
	}
}

func TestWayVersion(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	way, err := WayVersion(ctx, 106994776, 17)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if way.ID != 106994776 {
		t.Errorf("incorrect way: %v", way)
	}

	if way.Version != 17 {
		t.Errorf("incorrect way version: %v", way)
	}

	if l := len(way.Nodes); l != 4 {
		t.Errorf("incorrect number of way nodes: %v", l)
	}

	if l := len(way.Tags); l != 11 {
		t.Errorf("incorrect number of tags: %v", l)
	}
}

func TestWayHistory(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	ways, err := WayHistory(ctx, 106994776)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(ways); l != 17 {
		t.Errorf("incorrect number of ways: %d", l)
	}
}

func TestWayFull(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	o, err := WayFull(ctx, 106994776)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(o.Relations); l != 0 {
		t.Errorf("incorrect number of relations: %d", l)
	}

	if l := len(o.Ways); l != 1 {
		t.Errorf("incorrect number of ways: %d", l)
	}

	if l := len(o.Nodes); l != 4 {
		t.Errorf("incorrect number of nodes: %d", l)
	}
}

func TestWayRelations(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	relations, err := WayRelations(ctx, 106994776)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(relations); l != 4 {
		t.Errorf("incorrect number of relations: %d", l)
	}
}

func TestRelation(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	relation, err := Relation(ctx, 2714790)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if relation.ID != 2714790 {
		t.Errorf("incorrect relation: %v", relation)
	}
}

func TestRelations(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	relations, err := Relations(ctx, []osm.RelationID{2714790, 2714791, 2714792})
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(relations); l != 3 {
		t.Errorf("incorrect number of relations: %d", l)
	}
}

func TestRelationVersion(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	relation, err := RelationVersion(ctx, 2714790, 42)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if relation.ID != 2714790 {
		t.Errorf("incorrect relation: %v", relation)
	}

	if relation.Version != 42 {
		t.Errorf("incorrect version: %v", relation)
	}
}

func TestRelationRelations(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	relations, err := RelationRelations(ctx, 2714790)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(relations); l != 1 {
		t.Errorf("incorrect number of relations: %d", l)
	}
}

func TestRelationHistory(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	relations, err := RelationHistory(ctx, 2714790)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(relations); l < 42 {
		t.Errorf("incorrect number of relations: %d", l)
	}
}

func TestRelationFull(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	o, err := RelationFull(ctx, 2714790)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(o.Relations); l != 1 {
		t.Errorf("incorrect number of relations: %d", l)
	}

	if l := len(o.Ways); l < 100 {
		t.Errorf("incorrect number of ways: %d", l)
	}

	if l := len(o.Nodes); l < 383 {
		t.Errorf("incorrect number of nodes: %d", l)
	}
}

func TestMap(t *testing.T) {
	liveOnly(t)

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

func TestChangeset(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	c, err := Changeset(ctx, 56344850)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if c.ID != 56344850 {
		t.Errorf("incorrect id: %v", c.ID)
	}

	if c.Comment() != "remove duplicate node" {
		t.Errorf("incorrect comment: %v", c.Comment())
	}
}

func TestChangesetDownload(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	c, err := ChangesetDownload(ctx, 56344850)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(c.Delete.Nodes); l != 1 {
		t.Errorf("should be 1 node delete: %v", l)
	}

	if l := len(c.Modify.Ways); l != 1 {
		t.Errorf("should be 1 way modify: %v", l)
	}
}

func TestNotes(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	bound := &osm.Bounds{MinLon: -0.65094, MinLat: 51.312159, MaxLon: 0.374908, MaxLat: 51.669148}
	notes, err := Notes(ctx, bound, Limit(3), MaxDaysClosed(7))
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(notes); l != 3 {
		t.Errorf("incorrect number of notes: %d", l)
	}
}

func TestNote(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	note, err := Note(ctx, 123)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if note.ID != 123 {
		t.Errorf("incorrect note: %v", note)
	}
}

func TestNotesSearch(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	notes, err := NotesSearch(ctx, "Spam", Limit(2))
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(notes); l != 2 {
		t.Errorf("incorrect number of notes: %d", l)
	}
}

func TestUser(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	user, err := User(ctx, 91499)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if user.ID != 91499 {
		t.Errorf("incorrect user: %v", user)
	}
}
