# osm/osmpbf [![Go Reference](https://pkg.go.dev/badge/github.com/paulmach/osm.svg)](https://pkg.go.dev/github.com/paulmach/osm/osmpbf)

Package osmpbf provides a scanner for decoding large [OSM PBF](https://wiki.openstreetmap.org/wiki/PBF_Format) files.
They are typically found at [planet.osm.org](https://planet.openstreetmap.org/) or [Geofabrik Download](https://download.geofabrik.de/).

## Example:

```go
file, err := os.Open("./delaware-latest.osm.pbf")
if err != nil {
	panic(err)
}
defer file.Close()

// The third parameter is the number of parallel decoders to use.
scanner := osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
defer scanner.Close()

for scanner.Scan() {
	switch o := scanner.Object().(type) {
	case *osm.Node:

	case *osm.Way:

	case *osm.Relation:

	}
}

if err := scanner.Err(); err != nil {
	panic(err)
}
```

**Note:** Scanners are **not** safe for parallel use. One should feed the
objects into a channel and have workers read from that.

## Skipping Types

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

## Filtering Elements

The above skips all elements of a type. To filter based on the element's tags or
other values, use the filter functions. These filter functions are called in parallel
and not in a predefined order. This can be a performant way to filter for elements
with a certain set of tags.

```
type Scanner struct {
	// If the Filter function is false, the element well be skipped
	// at the decoding level. The functions should be fast, they block the
	// decoder, there are `procs` number of concurrent decoders.
	// Elements can be stored if the function returns true. Memory is
	// reused if the filter returns false.
	FilterNode     func(*osm.Node) bool
	FilterWay      func(*osm.Way) bool
	FilterRelation func(*osm.Relation) bool

	// contains filtered or unexported fields
}
```

## OSM PBF files with node locations on ways

This package supports reading OSM PBF files where the ways have been annotated with the coordinates of each node. Such files can be generated using [osmium](https://osmcode.org/osmium-tool), with the [add-locations-to-ways](https://docs.osmcode.org/osmium/latest/osmium-add-locations-to-ways.html) subcommand. This feature makes it possible to work with the ways and their geometries without having to keep all node locations in some index (which takes work and memory resources).

Coordinates are stored in the `Lat` and `Lon` fields of each `WayNode`. There is no need to specify an explicit option; when the node locations are present on the ways, they are loaded automatically. For more info about the OSM PBF format extension, see [the original blog post](https://blog.jochentopf.com/2016-04-20-node-locations-on-ways.html).

## Using cgo/czlib for decompression

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
