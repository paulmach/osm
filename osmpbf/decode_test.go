package osmpbf

import (
	"context"
	"fmt"
	"io"
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
)

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

var (
	IDsExpectedOrder = []string{
		// Start of dense nodes.
		"node/44", "node/47", "node/52", "node/58", "node/60",
		"node/79", // Just because way/79 is already there
		"node/2740703694", "node/2740703695", "node/2740703697",
		"node/2740703699", "node/2740703701",
		// End of dense nodes.

		// Start of ways.
		"way/73", "way/74", "way/75", "way/79", "way/482",
		"way/268745428", "way/268745431", "way/268745434", "way/268745436",
		"way/268745439",
		// End of ways.

		// Start of relations.
		"relation/69", "relation/94", "relation/152", "relation/245",
		"relation/332", "relation/3593436", "relation/3595575",
		"relation/3595798", "relation/3599126", "relation/3599127",
		// End of relations
	}

	IDs map[string]bool

	enc uint64 = 2729006
	ewc uint64 = 459055
	erc uint64 = 12833

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

	ew = &osm.Way{
		ID: 4257116,
		Nodes: osm.WayNodes{
			{ID: 21544864},
			{ID: 333731851},
			{ID: 333731852},
			{ID: 333731850},
			{ID: 333731855},
			{ID: 333731858},
			{ID: 333731854},
			{ID: 108047},
			{ID: 769984352},
			{ID: 21544864},
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

func downloadTestOSMFile(t *testing.T) {
	if _, err := os.Stat(London); os.IsNotExist(err) {
		out, err := os.Create(London)
		if err != nil {
			t.Fatal(err)
		}
		defer out.Close()

		resp, err := http.Get(LondonURL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("test status code invalid: %v", resp.StatusCode)
		}

		if _, err := io.Copy(out, resp.Body); err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
}

func TestDecode(t *testing.T) {
	downloadTestOSMFile(t)

	f, err := os.Open(London)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	d := newDecoder(context.Background(), &Scanner{}, f)
	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		t.Fatal(err)
	}

	var n *osm.Node
	var w *osm.Way
	var r *osm.Relation
	var nc, wc, rc uint64
	var id string
	idsOrder := make([]string, 0, len(IDsExpectedOrder))
	for {
		e, err := d.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}

		switch v := e.(type) {
		case *osm.Node:
			nc++
			if v.ID == en.ID {
				n = v
			}
			id = fmt.Sprintf("node/%d", v.ID)
			if _, ok := IDs[id]; ok {
				idsOrder = append(idsOrder, id)
			}
		case *osm.Way:
			wc++
			if v.ID == ew.ID {
				w = v
			}
			id = fmt.Sprintf("way/%d", v.ID)
			if _, ok := IDs[id]; ok {
				idsOrder = append(idsOrder, id)
			}
		case *osm.Relation:
			rc++
			if v.ID == er.ID {
				r = v
			}
			id = fmt.Sprintf("relation/%d", v.ID)
			if _, ok := IDs[id]; ok {
				idsOrder = append(idsOrder, id)
			}
		}
	}
	d.Close()

	if !reflect.DeepEqual(en, n) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", en, n)
	}
	if !reflect.DeepEqual(ew, w) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", ew, w)
	}
	if !reflect.DeepEqual(er, r) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", er, r)
	}
	if enc != nc || ewc != wc || erc != rc {
		t.Errorf("\nExpected %7d nodes, %7d ways, %7d relations\nGot %7d nodes, %7d ways, %7d relations.",
			enc, ewc, erc, nc, wc, rc)
	}
	if !reflect.DeepEqual(IDsExpectedOrder, idsOrder) {
		t.Errorf("\nExpected: %v\nGot:      %v", IDsExpectedOrder, idsOrder)
	}
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
