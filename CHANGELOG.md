# Changelog

All notable changes to this project will be documented in this file.

## [v0.8.0](https://github.com/paulmach/osm/compare/v0.7.1...v0.8.0) - 2024-01-08

### Changed

-   go 1.16 is required, updated usages of `ioutil` for similar functions in `io` and `os`

### Fixed

-   correctly JSON unmarshal elements with a type tag by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/53

## [v0.7.1](https://github.com/paulmach/osm/compare/v0.7.0...v0.7.1) - 2022-11-29

### Added

-   osm: add Tags.FindTag and Tags.HasTag methods by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/45

### Fixed

-   osm: support version as json number or string by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/46

## [v0.7.0](https://github.com/paulmach/osm/compare/v0.6.0...v0.7.0) - 2022-08-17

### Changed

-   remove node/ways/relations marshaling into this packages custom binary format by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/40

## [v0.6.0](https://github.com/paulmach/osm/compare/v0.5.0...v0.6.0) - 2022-08-16

### Added

-   json: ability to unmarshal osmjson by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/39
-   json: add support for external json implementations by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/39

## [v0.5.0](https://github.com/paulmach/osm/compare/v0.4.0...v0.5.0) - 2022-06-07

### Added

-   replication: ability to get changeset state by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/37
-   replication: search for state/sequence number by timestamp by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/38

## [v0.4.0](https://github.com/paulmach/osm/compare/v0.3.0...v0.4.0) - 2022-05-26

### Changed

-   protobuf: port to google protobuf by [@OlafFlebbeBosch](https://github.com/OlafFlebbeBoch) in https://github.com/paulmach/osm/pull/36

## [v0.3.0](https://github.com/paulmach/osm/compare/v0.2.2...v0.3.0) - 2022-04-21

### Added

-   osmpbf: preallocation node tags array by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/33
-   osmpbf: support "sparse" dense nodes by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/32
-   osmpbf: add filter functions by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/30

## [v0.2.2](https://github.com/paulmach/osm/compare/v0.2.1...v0.2.2) - 2021-04-27

### Fixed

-   osmpbf: fixed memory allocation issues by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/26

## [v0.2.1](https://github.com/paulmach/osm/compare/v0.2.0...v0.2.1) - 2021-02-04

### Changed

-   osmpbf: reduces memory usage when decoding by [@oflebbe](https://github.com/oflebbe) in https://github.com/paulmach/osm/pull/22
-   Fix some more typos by [@meyermarcel](https://github.com/meyermarcel) in https://github.com/paulmach/osm/pull/23

## [v0.2.0](https://github.com/paulmach/osm/compare/v0.1.1...v0.2.0) - 2021-01-09

### Changed

-   osmpbf: ability to efficiently skip types when decoding by [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/18
-   osmpbf: use [protoscan](https://github.com/paulmach/protoscan) for a 10%ish performance improvement
-   osmpbf: use cgo/czlib to decode protobufs (if cgo enabled), 20% faster on benchmarks [@paulmach](https://github.com/paulmach) in https://github.com/paulmach/osm/pull/19
-   deprecated node/ways/relations marshaling into this packages custom binary format [`8fcda5`](https://github.com/paulmach/osm/commit/8fcda5dc49b4767df63eccb5a25f3e63d5b17f4d)
