osm/replication [![Godoc Reference](https://godoc.org/github.com/paulmach/osm/replication?status.png)](https://godoc.org/github.com/paulmach/osm/replication)
===============

Package `replication` handles fetch the Minute, Hour, Day and Changeset replication
and the assocated state value from [Planet OSM](http://planet.osm.org).

For example, to fetch the current Minute replication state:

	num, fullState, err := replication.CurrentMinuteState(ctx)

This is the data in [this file](http://planet.osm.org/replication/minute/state.txt)
updated every minute.

Once you know the change you want, fetch the change using:

	change, err := replication.Minute(ctx, num)
