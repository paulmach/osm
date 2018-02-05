osm/osmapi [![Godoc Reference](https://godoc.org/github.com/paulmach/osm/osmapi?status.png)](https://godoc.org/github.com/paulmach/osm/osmapi)
==========

Package osmapi provides an interface to the [OSM v0.6 API](https://wiki.openstreetmap.org/wiki/API_v0.6).

Usage:

	node, err := osmapi.Node(ctx, 1010)

This call issues a request to [api.openstreetmap.org/api/0.6/node/1010](https://api.openstreetmap.org/api/0.6/node/1010)
and returns a parsed `osm.Node` object with all the methods attached.

### List of methods

```go
func Map(context.Context, bounds *osm.Bounds) (*osm.OSM, error)

func Node(context.Context, osm.NodeID) (*osm.Node, error)
func Nodes(context.Context, []osm.NodeID) (osm.Nodes, error)
func NodeVersion(context.Context, osm.NodeID, v int) (*osm.Node, error)
func NodeHistory(context.Context, osm.NodeID) (osm.Nodes, error)

func NodeWays(context.Context, osm.NodeID) (osm.Ways, error)
func NodeRelations(context.Context, osm.NodeID) (osm.Relations, error)

func Way(context.Context, osm.WayID) (*osm.Way, error)
func Ways(context.Context, []osm.WayID) (osm.Ways, error)
func WayFull(context.Context, osm.WayID) (*osm.OSM, error)
func WayVersion(context.Context, osm.WayID, v int) (*osm.Way, error)
func WayHistory(context.Context, osm.WayID) (osm.Ways, error)

func WayRelations(context.Context, osm.WayID) (osm.Relations, error)

func Relation(context.Context, osm.RelationID) (*osm.Relation, error)
func Relations(context.Context, []osm.RelationID) (osm.Relations, error)
func RelationFull(context.Context, osm.RelationID) (*osm.OSM, error)
func RelationVersion(context.Context, osm.RelationID, v int) (*osm.Relation, error)
func RelationHistory(context.Context, osm.RelationID) (osm.Relations, error)

func RelationRelations(context.Context, osm.RelationID) (osm.Relations, error)

func Changeset(context.Context, osm.ChangesetID) (*osm.Changeset, error)
func ChangesetWithDiscussion(context.Context, osm.ChangesetID) (*osm.Changeset, error)
func ChangesetDownload(context.Context, osm.ChangesetID) (*osm.Change, error)
```

See the [godoc reference](https://godoc.org/github.com/paulmach/osm/osmapi)
for more details.

### Rate limiting

This package can make sure of [`x/time/rate.Limiter`](https://godoc.org/golang.org/x/time/rate#Limiter)
 to throttle requests to the official api. Example usage:

	// 10 qps
	osmapi.DefaultDatasource.Limiter = rate.NewLimiter(10, 1)
