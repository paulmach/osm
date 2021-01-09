osm/osmpbf [![Go Reference](https://pkg.go.dev/badge/github.com/paulmach/osm.svg)](https://pkg.go.dev/github.com/paulmach/osm/osmpbf)
==========

Package osmpbf provides a scanner for decoding large [OSM PBF](https://wiki.openstreetmap.org/wiki/PBF_Format) files.
They are typically found at [planet.osm.org](https://planet.openstreetmap.org/) or [Geofabrik Download](https://download.geofabrik.de/).

### Example:

```go
file, err := os.Open("./delaware-latest.osm.pbf")
if err != nil {
	panic(err)
}
defer f.Close()

// The third parameter is the number of parallel decoders to use.
scanner := osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
defer scanner.Close()

for scanner.Scan() {
	switch o := scanner.Object().(type)
	case *osm.Node:

	case *osm.Way:

	case *osm.Relation:
}

if err := scanner.Err(); err != nil {
	panic(err)
}
```

**Note:** Scanners are **not** safe for parallel use. One should feed the
objects into a channel and have workers read from that.

### Skipping Types

Sometimes only ways or relations are needed. In this case reading and creating
those objects can be skipped completely. After creating the Scanner set the appropriate
attributes to true.

```
type Scanner struct {
	// Skip element types that are not needed. The data is skipped
	// at the encoded protobuf level, but each block still
	// needs to be decompressed.
	SkipNodes     bool
	SkipWays      bool
	SkipRelations bool

	// contains filtered or unexported fields
}
```

### Using cgo/czlib for decompression

OSM PBF files are a set of blocks that are zlib compressed. When using the pure golang
implementation this can account for about 1/3 of the read time. When cgo is enabled
the package will used [czlib](https://github.com/DataDog/czlib).

```
$ CGO_ENABLED=0 go test -bench . > disabled.txt
$ CGO_ENABLED=1 go test -bench . > enabled.txt
$ benchcmp disabled.txt enabled.txt
benchmark                        old ns/op     new ns/op     delta
BenchmarkLondon-12               312294630     229927205     -26.37%
BenchmarkLondon_nodes-12         246562457     160021768     -35.10%
BenchmarkLondon_ways-12          216803544     134747327     -37.85%
BenchmarkLondon_relations-12     158722633     80560144      -49.24%

benchmark                        old allocs     new allocs     delta
BenchmarkLondon-12               2469128        2416804        -2.12%
BenchmarkLondon_nodes-12         1056166        1003850        -4.95%
BenchmarkLondon_ways-12          1845032        1792716        -2.84%
BenchmarkLondon_relations-12     509090         456772         -10.28%

benchmark                        old bytes     new bytes     delta
BenchmarkLondon-12               963734544     954877896     -0.92%
BenchmarkLondon_nodes-12         658337435     649482060     -1.35%
BenchmarkLondon_ways-12          441674734     432819378     -2.00%
BenchmarkLondon_relations-12     187941609     179086389     -4.71%
```
