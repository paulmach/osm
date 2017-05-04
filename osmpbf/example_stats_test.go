package osmpbf_test

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	osm "github.com/paulmach/go.osm"
	"github.com/paulmach/go.osm/osmpbf"
)

// ExampleStats demonstrates how to read a full file and gather some stats.
// This is similar to `osmconvert --out-statistics`
func Example_stats() {
	f, err := os.Open("../testdata/delaware-latest.osm.pbf")
	if err != nil {
		fmt.Printf("could not open file: %v", err)
		os.Exit(1)
	}

	nodes, ways, relations := 0, 0, 0

	minLat, maxLat := math.MaxFloat64, -math.MaxFloat64
	minLon, maxLon := math.MaxFloat64, -math.MaxFloat64

	maxVersion := 0
	minNodeID, maxNodeID := osm.NodeID(math.MaxInt64), osm.NodeID(0)
	minWayID, maxWayID := osm.WayID(math.MaxInt64), osm.WayID(0)

	minRelationID, maxRelationID := osm.RelationID(math.MaxInt64), osm.RelationID(0)

	minTS, maxTS := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC), time.Time{}

	var (
		maxTags     int
		maxTagsType osm.Type
		maxTagsID   int64
	)

	var (
		maxNodeRefs   int
		maxNodeRefsID osm.WayID
	)

	var (
		maxRelRefs   int
		maxRelRefsID osm.RelationID
	)

	scanner := osmpbf.New(context.Background(), f, 3)
	for scanner.Scan() {
		var ts time.Time

		switch e := scanner.Element().(type) {
		case *osm.Node:
			nodes++
			ts = e.Timestamp

			if e.Lat > maxLat {
				maxLat = e.Lat
			}

			if e.Lat < minLat {
				minLat = e.Lat
			}

			if e.Lon > maxLon {
				maxLon = e.Lon
			}

			if e.Lon < minLon {
				minLon = e.Lon
			}

			if e.ID > maxNodeID {
				maxNodeID = e.ID
			}

			if e.ID < minNodeID {
				minNodeID = e.ID
			}

			if e.Version > maxVersion {
				maxVersion = e.Version
			}

			if l := len(e.Tags); l > maxTags {
				maxTags = l
				maxTagsType = osm.TypeNode
				maxTagsID = int64(e.ID)
			}
		case *osm.Way:
			ways++
			ts = e.Timestamp

			if e.ID > maxWayID {
				maxWayID = e.ID
			}

			if e.ID < minWayID {
				minWayID = e.ID
			}

			if e.Version > maxVersion {
				maxVersion = e.Version
			}

			if l := len(e.Tags); l > maxTags {
				maxTags = l
				maxTagsType = osm.TypeWay
				maxTagsID = int64(e.ID)
			}

			if l := len(e.Nodes); l > maxNodeRefs {
				maxNodeRefs = l
				maxNodeRefsID = e.ID
			}
		case *osm.Relation:
			relations++
			ts = e.Timestamp

			if e.ID > maxRelationID {
				maxRelationID = e.ID
			}

			if e.ID < minRelationID {
				minRelationID = e.ID
			}

			if e.Version > maxVersion {
				maxVersion = e.Version
			}

			if l := len(e.Tags); l > maxTags {
				maxTags = l
				maxTagsType = osm.TypeRelation
				maxTagsID = int64(e.ID)
			}

			if l := len(e.Members); l > maxRelRefs {
				maxRelRefs = l
				maxRelRefsID = e.ID
			}
		}

		if ts.After(maxTS) {
			maxTS = ts
		}

		if ts.Before(minTS) {
			minTS = ts
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("scanner returned error: %v", err)
		os.Exit(1)
	}

	fmt.Println("timestamp min:", minTS.Format(time.RFC3339))
	fmt.Println("timestamp max:", maxTS.Format(time.RFC3339))
	fmt.Printf("lon min: %0.7f\n", minLon)
	fmt.Printf("lon max: %0.7f\n", maxLon)
	fmt.Printf("lat min: %0.7f\n", minLat)
	fmt.Printf("lat max: %0.7f\n", maxLat)
	fmt.Println("nodes:", nodes)
	fmt.Println("ways:", ways)
	fmt.Println("relations:", relations)
	fmt.Println("version max:", maxVersion)
	fmt.Println("node id min:", minNodeID)
	fmt.Println("node id max:", maxNodeID)
	fmt.Println("way id min:", minWayID)
	fmt.Println("way id max:", maxWayID)
	fmt.Println("relation id min:", minRelationID)
	fmt.Println("relation id max:", maxRelationID)
	fmt.Println("keyval pairs max:", maxTags)
	fmt.Println("keyval pairs max object:", maxTagsType, maxTagsID)
	fmt.Println("noderefs max:", maxNodeRefs)
	fmt.Println("noderefs max object: way", maxNodeRefsID)
	fmt.Println("relrefs max:", maxRelRefs)
	fmt.Println("relrefs max object: relation", maxRelRefsID)

	// Output:
	// timestamp min: 2007-10-16T15:59:24Z
	// timestamp max: 2016-08-10T17:32:02Z
	// lon min: -76.1748935
	// lon max: -74.4929376
	// lat min: 38.0273717
	// lat max: 39.9688859
	// nodes: 723870
	// ways: 73144
	// relations: 1644
	// version max: 421
	// node id min: 75385503
	// node id max: 4343778904
	// way id min: 9650669
	// way id max: 436488690
	// relation id min: 82010
	// relation id max: 6462005
	// keyval pairs max: 276
	// keyval pairs max object: relation 148838
	// noderefs max: 1811
	// noderefs max object: way 318739264
	// relrefs max: 7177
	// relrefs max object: relation 4799100
}
