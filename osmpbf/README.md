# osm/osmpbf [![Go Reference](https://pkg.go.dev/badge/github.com/paulmach/osm.svg)](https://pkg.go.dev/github.com/paulmach/osm/osmpbf)

Package osmpbf provides a scanner for decoding large [OSM PBF](https://wiki.openstreetmap.org/wiki/PBF_Format) files.
They are typically found at [planet.osm.org](https://planet.openstreetmap.org/) or [Geofabrik Download](https://download.geofabrik.de/).

## Example:

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

## Using cgo libdeflate for decompression

OSM PBF files are a set of blocks that are zlib compressed. When using the pure golang
implementation this can account for about 1/3 of the read time. When cgo is enabled
the package [go-libdeflate](https://github.com/4kills/libdeflate) will used.

Previous versions used the lib czlib based on zlib. libdeflate is more performant
and more memory efficient for uncompressing.

```
$ CGO_ENABLED=0 go test -bench . > disabled.txt
$ CGO_ENABLED=1 go test -bench . > enabled.txt
$ benchcmp disabled.txt enabled.txt
benchmark                              old ns/op     new ns/op     delta
BenchmarkLondon-8                      361519289     275254714     -23.86%
BenchmarkLondon_withFiltersTrue-8      392469042     263935960     -32.75%
BenchmarkLondon_withFiltersFalse-8     310824940     200477972     -35.50%
BenchmarkLondon_nodes-8                295277528     180614979     -38.83%
BenchmarkLondon_ways-8                 257494509     140700970     -45.36%
BenchmarkLondon_relations-8            189490128     75263200      -60.28%

benchmark                              old allocs     new allocs     delta
BenchmarkLondon-8                      4863784        4808526        -1.14%
BenchmarkLondon_withFiltersTrue-8      4863786        4808515        -1.14%
BenchmarkLondon_withFiltersFalse-8     1419995        1364724        -3.89%
BenchmarkLondon_nodes-8                3450825        3395559        -1.60%
BenchmarkLondon_ways-8                 1851359        1796099        -2.98%
BenchmarkLondon_relations-8            515422         460152         -10.72%

benchmark                              old bytes     new bytes     delta
BenchmarkLondon-8                      947061317     924789892     -2.35%
BenchmarkLondon_withFiltersTrue-8      947061146     924787588     -2.35%
BenchmarkLondon_withFiltersFalse-8     388725836     366452840     -5.73%
BenchmarkLondon_nodes-8                641663624     619391213     -3.47%
BenchmarkLondon_ways-8                 460631859     438360054     -4.84%
BenchmarkLondon_relations-8            206899749     184626277     -10.77%