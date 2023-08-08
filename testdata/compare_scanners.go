package main

import (
	"compress/bzip2"
	"context"
	"log"
	"math"
	"os"
	"reflect"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/paulmach/osm/osmxml"
)

func main() {
	// these can be downloaded at http://download.geofabrik.de/north-america.html
	pbffile := "delaware-latest.osm.pbf"
	bz2file := "delaware-latest.osm.bz2"

	o1 := readPbf(pbffile)
	o2 := readBz2(bz2file)

	log.Printf("Are they the same? %v", reflect.DeepEqual(o1, o2))

	log.Printf("nodes: %v %v", len(o1.Nodes), len(o2.Nodes))
	if len(o1.Nodes) == len(o2.Nodes) {
		for i := range o1.Nodes {
			if !reflect.DeepEqual(o1.Nodes[i], o2.Nodes[i]) {
				log.Printf("unequal nodes")
				log.Printf("%v", o1.Nodes[i])
				log.Printf("%v", o2.Nodes[i])
			}
		}
	}

	log.Printf("ways: %v %v", len(o1.Ways), len(o2.Ways))
	if len(o1.Ways) == len(o2.Ways) {
		for i := range o1.Ways {
			if !reflect.DeepEqual(o1.Ways[i], o2.Ways[i]) {
				log.Printf("unequal ways")
				log.Printf("%v", o1.Ways[i])
				log.Printf("%v", o2.Ways[i])
			}
		}
	}

	log.Printf("relations: %v %v", len(o1.Relations), len(o2.Relations))
	if len(o1.Relations) == len(o2.Relations) {
		for i := range o1.Relations {
			if !reflect.DeepEqual(o1.Relations[i], o2.Relations[i]) {
				log.Printf("unequal relations")
				log.Printf("%v", o1.Relations[i])
				log.Printf("%v", o2.Relations[i])
			}
		}
	}
}

func readPbf(filename string) *osm.OSM {
	log.Printf("Reading pbf file %v", filename)
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := osmpbf.New(context.Background(), f, 1)
	return scanner2osm(scanner)
}

func readBz2(filename string) *osm.OSM {
	log.Printf("Reading bz2 file %v", filename)
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bzip2.NewReader(f)
	scanner := osmxml.New(context.Background(), r)
	return scanner2osm(scanner)
}

func scanner2osm(scanner osm.Scanner) *osm.OSM {
	o := &osm.OSM{}
	for scanner.Scan() {
		e := scanner.Element()

		if e.Node != nil {
			e.Node.Lat = math.Floor(e.Node.Lat*1e7+0.5) / 1e7
			e.Node.Lon = math.Floor(e.Node.Lon*1e7+0.5) / 1e7
			e.Node.Visible = true
			e.Node.Tags.SortByKeyValue()
			o.Nodes = append(o.Nodes, e.Node)
		}

		if e.Way != nil {
			e.Way.Visible = true
			e.Way.Tags.SortByKeyValue()
			o.Ways = append(o.Ways, e.Way)
		}

		if e.Relation != nil {
			e.Relation.Visible = true
			e.Relation.Tags.SortByKeyValue()
			o.Relations = append(o.Relations, e.Relation)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	o.Nodes.SortByIDVersion()
	o.Ways.SortByIDVersion()
	o.Relations.SortByIDVersion()

	return o
}
