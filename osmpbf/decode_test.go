package osmpbf

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/paulmach/osm"
)

const (
	// Originally downloaded from http://download.geofabrik.de/europe/great-britain/england/greater-london.html
	London    = "greater-london-140324.osm.pbf"
	LondonURL = "https://gist.githubusercontent.com/paulmach/853d57b83d408480d3b148b07954c110/raw/853f33f4dbe4246915134f1cde8edb30241ecc10/greater-london-140324.osm.pbf"
	// Created based on the above file, by running `osmium add-locations-to-ways`.
	LondonLocations    = "greater-london-140324-low.osm.pbf"
	LondonLocationsURL = "https://gist.github.com/paulmach/853d57b83d408480d3b148b07954c110/raw/d3dd351fcb202e3db1c77b44313c1ba0d71b43b3/greater-london-140324-low.osm.pbf"

	coordinatesPrecision = 1e7
)

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func stripCoordinates(w *osm.Way) *osm.Way {
	if w == nil {
		return nil
	}

	ws := new(osm.Way)
	*ws = *w
	ws.Nodes = make(osm.WayNodes, len(w.Nodes))
	for i, n := range w.Nodes {
		n.Lat, n.Lon = 0, 0
		ws.Nodes[i] = n
	}
	return ws
}

func roundCoordinates(w *osm.Way) {
	if w == nil {
		return
	}
	for i := range w.Nodes {
		w.Nodes[i].Lat = math.Round(w.Nodes[i].Lat*coordinatesPrecision) / coordinatesPrecision
		w.Nodes[i].Lon = math.Round(w.Nodes[i].Lon*coordinatesPrecision) / coordinatesPrecision
	}
}

type OSMFileTest struct {
	*testing.T
	FileName                               string
	FileURL                                string
	ExpNode                                *osm.Node
	ExpWay                                 *osm.Way
	ExpRel                                 *osm.Relation
	ExpNodeCount, ExpWayCount, ExpRelCount uint64
	IDsExpOrder                            []string
}

var (
	idsExpectedOrderNodes = []string{
		"node/44", "node/47", "node/52", "node/58", "node/60",
		"node/79", // Just because way/79 is already there
		"node/2740703694", "node/2740703695", "node/2740703697",
		"node/2740703699", "node/2740703701",
	}
	idsExpectedOrderWays = []string{
		"way/73", "way/74", "way/75", "way/79", "way/482",
		"way/268745428", "way/268745431", "way/268745434", "way/268745436",
		"way/268745439",
	}
	idsExpectedOrderRelations = []string{
		"relation/69", "relation/94", "relation/152", "relation/245",
		"relation/332", "relation/3593436", "relation/3595575",
		"relation/3595798", "relation/3599126", "relation/3599127",
	}
	IDsExpectedOrderNoNodes = append(idsExpectedOrderWays, idsExpectedOrderRelations...)
	IDsExpectedOrder        = append(idsExpectedOrderNodes, IDsExpectedOrderNoNodes...)

	IDs map[string]bool

	enc  uint64 = 2729006
	encl uint64 = 244523
	ewc  uint64 = 459055
	erc  uint64 = 12833

	en = &osm.Node{
		ID:  18088578,
		Lat: 51.5442632,
		Lon: -0.2010027,
		Tags: osm.Tags([]osm.Tag{
			{Key: "alt_name", Value: "The King's Head"},
			{Key: "amenity", Value: "pub"},
			{Key: "created_by", Value: "JOSM"},
			{Key: "name", Value: "The Luminaire"},
			{Key: "note", Value: "Live music venue too"},
		}),
		Version:     2,
		Timestamp:   parseTime("2009-05-20T10:28:54Z"),
		ChangesetID: 1260468,
		UserID:      508,
		User:        "Welshie",
		Visible:     true,
	}

	ewl = &osm.Way{
		ID: 4257116,
		Nodes: osm.WayNodes{
			{ID: 21544864, Lat: 51.5230531, Lon: -0.1408525},
			{ID: 333731851, Lat: 51.5224309, Lon: -0.1402297},
			{ID: 333731852, Lat: 51.5224107, Lon: -0.1401878},
			{ID: 333731850, Lat: 51.522422, Lon: -0.1401375},
			{ID: 333731855, Lat: 51.522792, Lon: -0.1392477},
			{ID: 333731858, Lat: 51.5228209, Lon: -0.1392124},
			{ID: 333731854, Lat: 51.5228579, Lon: -0.1392339},
			{ID: 108047, Lat: 51.5234407, Lon: -0.1398771},
			{ID: 769984352, Lat: 51.5232469, Lon: -0.1403648},
			{ID: 21544864, Lat: 51.5230531, Lon: -0.1408525},
		},
		Tags: osm.Tags([]osm.Tag{
			{Key: "area", Value: "yes"},
			{Key: "highway", Value: "pedestrian"},
			{Key: "name", Value: "Fitzroy Square"},
		}),
		Version:     7,
		Timestamp:   parseTime("2013-08-07T12:08:39Z"),
		ChangesetID: 17253164,
		UserID:      1016290,
		User:        "Amaroussi",
		Visible:     true,
	}

	ew = stripCoordinates(ewl)

	er = &osm.Relation{
		ID: 7677,
		Members: osm.Members{
			{Ref: 4875932, Type: osm.TypeWay, Role: "outer"},
			{Ref: 4894305, Type: osm.TypeWay, Role: "inner"},
		},
		Tags: osm.Tags([]osm.Tag{
			{Key: "created_by", Value: "Potlatch 0.9c"},
			{Key: "type", Value: "multipolygon"},
		}),
		Version:     4,
		Timestamp:   parseTime("2008-07-19T15:04:03Z"),
		ChangesetID: 540201,
		UserID:      3876,
		User:        "Edgemaster",
		Visible:     true,
	}
)

func init() {
	IDs = make(map[string]bool)
	for _, id := range IDsExpectedOrder {
		IDs[id] = false
	}
}

func (ft *OSMFileTest) downloadTestOSMFile() {
	if _, err := os.Stat(ft.FileName); os.IsNotExist(err) {
		out, err := os.Create(ft.FileName)
		if err != nil {
			ft.Fatal(err)
		}
		defer out.Close()

		resp, err := http.Get(ft.FileURL)
		if err != nil {
			ft.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ft.Fatalf("test status code invalid: %v", resp.StatusCode)
		}

		if _, err := io.Copy(out, resp.Body); err != nil {
			ft.Fatal(err)
		}
	} else if err != nil {
		ft.Fatal(err)
	}
}

func (ft *OSMFileTest) testDecode() {
	ft.downloadTestOSMFile()

	f, err := os.Open(ft.FileName)
	if err != nil {
		ft.Fatal(err)
	}
	defer f.Close()

	d := newDecoder(context.Background(), &Scanner{}, f)
	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		ft.Fatal(err)
	}

	var n *osm.Node
	var w *osm.Way
	var r *osm.Relation
	var nc, wc, rc uint64
	idsOrder := make([]string, 0, len(IDsExpectedOrder))
	for {
		e, err := d.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			ft.Fatal(err)
		}

		switch v := e.(type) {
		case *osm.Node:
			nc++
			if v.ID == ft.ExpNode.ID {
				n = v
			}
			id := fmt.Sprintf("node/%d", v.ID)
			if _, ok := IDs[id]; ok {
				idsOrder = append(idsOrder, id)
			}
		case *osm.Way:
			wc++
			if v.ID == ft.ExpWay.ID {
				w = v
			}
			id := fmt.Sprintf("way/%d", v.ID)
			if _, ok := IDs[id]; ok {
				idsOrder = append(idsOrder, id)
			}
		case *osm.Relation:
			rc++
			if v.ID == ft.ExpRel.ID {
				r = v
			}
			id := fmt.Sprintf("relation/%d", v.ID)
			if _, ok := IDs[id]; ok {
				idsOrder = append(idsOrder, id)
			}
		}
	}
	d.Close()

	if !reflect.DeepEqual(ft.ExpNode, n) {
		ft.Errorf("\nExpected: %#v\nActual:   %#v", ft.ExpNode, n)
	}
	roundCoordinates(w)
	if !reflect.DeepEqual(ft.ExpWay, w) {
		ft.Errorf("\nExpected: %#v\nActual:   %#v", ft.ExpWay, w)
	}
	if !reflect.DeepEqual(ft.ExpRel, r) {
		ft.Errorf("\nExpected: %#v\nActual:   %#v", ft.ExpRel, r)
	}
	if ft.ExpNodeCount != nc || ft.ExpWayCount != wc || ft.ExpRelCount != rc {
		ft.Errorf("\nExpected %7d nodes, %7d ways, %7d relations\nGot %7d nodes, %7d ways, %7d relations.",
			ft.ExpNodeCount, ft.ExpWayCount, ft.ExpRelCount, nc, wc, rc)
	}
	if !reflect.DeepEqual(ft.IDsExpOrder, idsOrder) {
		ft.Errorf("\nExpected: %v\nGot:      %v", ft.IDsExpOrder, idsOrder)
	}
}

func TestDecode(t *testing.T) {
	ft := &OSMFileTest{
		T:            t,
		FileName:     London,
		FileURL:      LondonURL,
		ExpNode:      en,
		ExpWay:       ew,
		ExpRel:       er,
		ExpNodeCount: enc,
		ExpWayCount:  ewc,
		ExpRelCount:  erc,
		IDsExpOrder:  IDsExpectedOrder,
	}
	ft.testDecode()
}

func TestDecodeLocations(t *testing.T) {
	ft := &OSMFileTest{
		T:            t,
		FileName:     LondonLocations,
		FileURL:      LondonLocationsURL,
		ExpNode:      en,
		ExpWay:       ewl,
		ExpRel:       er,
		ExpNodeCount: encl,
		ExpWayCount:  ewc,
		ExpRelCount:  erc,
		IDsExpOrder:  IDsExpectedOrderNoNodes,
	}
	ft.testDecode()
}

func TestDecode_Close(t *testing.T) {
	f, err := os.Open(Delaware)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// should close at start
	f.Seek(0, 0)
	d := newDecoder(context.Background(), &Scanner{}, f)
	d.Start(5)

	err = d.Close()
	if err != nil {
		t.Errorf("close error: %v", err)
	}

	// should close after partial read
	f.Seek(0, 0)
	d = newDecoder(context.Background(), &Scanner{}, f)
	d.Start(5)

	d.Next()
	d.Next()

	err = d.Close()
	if err != nil {
		t.Errorf("close error: %v", err)
	}

	// should close after full read
	f.Seek(0, 0)
	d = newDecoder(context.Background(), &Scanner{}, f)
	d.Start(5)

	elements := 0
	for {
		_, err := d.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Errorf("next error: %v", err)
		}

		elements++
	}

	if elements < 2 {
		t.Errorf("did not read enough elements: %v", elements)
	}

	// should close at end of read
	err = d.Close()
	if err != nil {
		t.Errorf("close error: %v", err)
	}
}
