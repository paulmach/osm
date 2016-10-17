package osmpbf_test

import (
	"fmt"
	"math"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/paulmach/go.osm"
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

	nodes := 0
	ways := 0
	relations := 0

	minLat := math.MaxFloat64
	maxLat := -math.MaxFloat64
	minLon := math.MaxFloat64
	maxLon := -math.MaxFloat64

	minNodeID := osm.NodeID(math.MaxInt64)
	maxNodeID := osm.NodeID(0)

	minWayID := osm.WayID(math.MaxInt64)
	maxWayID := osm.WayID(0)

	minRelationID := osm.RelationID(math.MaxInt64)
	maxRelationID := osm.RelationID(0)

	minTS := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC) // TODO: update in year 2100
	maxTS := time.Time{}

	var (
		maxTags     int
		maxTagsType osm.ElementType
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

		e := scanner.Element()
		if e.Node != nil {
			nodes++
			ts = e.Node.Timestamp

			if e.Node.Lat > maxLat {
				maxLat = e.Node.Lat
			}

			if e.Node.Lat < minLat {
				minLat = e.Node.Lat
			}

			if e.Node.Lon > maxLon {
				maxLon = e.Node.Lon
			}

			if e.Node.Lon < minLon {
				minLon = e.Node.Lon
			}

			if e.Node.ID > maxNodeID {
				maxNodeID = e.Node.ID
			}

			if e.Node.ID < minNodeID {
				minNodeID = e.Node.ID
			}

			if l := len(e.Node.Tags); l > maxTags {
				maxTags = l
				maxTagsType = osm.NodeType
				maxTagsID = int64(e.Node.ID)
			}
		} else if e.Way != nil {
			ways++
			ts = e.Way.Timestamp

			if e.Way.ID > maxWayID {
				maxWayID = e.Way.ID
			}

			if e.Way.ID < minWayID {
				minWayID = e.Way.ID
			}

			if l := len(e.Way.Tags); l > maxTags {
				maxTags = l
				maxTagsType = osm.WayType
				maxTagsID = int64(e.Way.ID)
			}

			if l := len(e.Way.Nodes); l > maxNodeRefs {
				maxNodeRefs = l
				maxNodeRefsID = e.Way.ID
			}
		} else if e.Relation != nil {
			relations++
			ts = e.Relation.Timestamp

			if e.Relation.ID > maxRelationID {
				maxRelationID = e.Relation.ID
			}

			if e.Relation.ID < minRelationID {
				minRelationID = e.Relation.ID
			}

			if l := len(e.Relation.Tags); l > maxTags {
				maxTags = l
				maxTagsType = osm.RelationType
				maxTagsID = int64(e.Relation.ID)
			}

			if l := len(e.Relation.Members); l > maxRelRefs {
				maxRelRefs = l
				maxRelRefsID = e.Relation.ID
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
