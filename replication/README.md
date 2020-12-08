osm/replication [![Godoc Reference](https://godoc.org/github.com/paulmach/osm/replication?status.svg)](https://godoc.org/github.com/paulmach/osm/replication)
===============

Package `replication` handles fetching the Minute, Hour, Day and Changeset replication
and the associated state value from [Planet OSM](http://planet.osm.org).

For example, to fetch the current Minute replication state:

```go
num, fullState, err := replication.CurrentMinuteState(ctx)
```

This is the data in [http://planet.osm.org/replication/minute/state.txt](http://planet.osm.org/replication/minute/state.txt)
updated every minute.

Once you know the change number you want, fetch the change using:

```go
change, err := replication.Minute(ctx, num)
```
