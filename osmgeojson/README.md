osm/osmgeojson [![Godoc Reference](https://godoc.org/github.com/paulmach/osm/osmgeojson?status.svg)](https://godoc.org/github.com/paulmach/osm/osmgeojson)
==============

Package `osmgeojson` converts OSM data to GeoJSON. It is a **full** port of the
nodejs library [osmtogeojson](https://github.com/tyrasd/osmtogeojson) and sports
the same features and tests (plus more):

* real OSM polygon detection
* OSM multipolygon support, e.g. buildings with holes become proper multipolygons
* supports annotated geometries
* well tested

### Usage

```go
delta := 0.0001

lon, lat := -83.5997038, 41.5923682
bounds := &osm.Bounds{
	MinLat: lat - delta, MaxLat: lat + delta,
	MinLon: lon - delta, MaxLon: lon + delta,
}

o, _ := osmapi.Map(ctx, bounds)  // fetch data from the osm api.

// run the conversion
fc, err := osmgeojson.Convert(o, opts)

// marshal the json
gj, _ := json.MarshalIndent(fc, "", " ")
fmt.Println(string(gj))
```

### Options

The package provides several options to control what is included in the feature properties.
If possible, excluding some of the extra properties	can greatly improve the performance.
All of the options **default to false**, i.e. everything will be included.

* `NoID(yes bool)`

	Controls whether to set the feature.ID to "type/id" e.g. "node/475373687". For some use cases
	this may be of limited use since the feature.Properies "type" and "id" are also set.

* `NoMeta(yes bool)`

	Controls whether to populate the "meta" property which is a sub-map with the
	following values from the osm element: "timestamp", "version", "changeset", "user", "uid".

* `NoRelationMembership(yes bool)`

	Controls whether to include a list of the relations the osm element is a member of.
	This info is set as the "relation" property which is an array of objects with the
	following values from the relation: "id", "role", "tags".

* `IncludeInvalidPolygons(yes bool)`

	By default, inner rings of 'multipolygon' without a matching outer ring will be ignored.
	However, in some use cases the outer ring can be implied as the viewport bound and the inner rings
	can then be rendered correctly. Polygons with a nil first ring will be need to be updated such
	that the first ring is the viewport bound. This options will also include rings that do not
	have matching endpoints. Usually this means one or more of the outer ways are missing.


### Benchmarks

These benchmarks are meant to show the performance impact of the different options.
They were run on a 2012 MacBook Air with a 2 ghz processor and 8 gigs of ram.

```
BenchmarkConvert-4                        10000     2520891 ns/op     935697 B/op     11299 allocs/op
BenchmarkConvertAnnotated-4               10000     2196433 ns/op     853544 B/op     11239 allocs/op
BenchmarkConvert_NoID-4                   10000     2310816 ns/op     913915 B/op      9687 allocs/op
BenchmarkConvert_NoMeta-4                 10000     2026031 ns/op     716953 B/op      7546 allocs/op
BenchmarkConvert_NoRelationMembership-4   10000     2397634 ns/op     912454 B/op     10716 allocs/op
BenchmarkConvert_NoIDsMetaMembership-4    20000     1718224 ns/op     671984 B/op      5353 allocs/op
```

#### Similar libraries in other languages:

* [osmtogeojson](https://github.com/tyrasd/osmtogeojson) - Node
