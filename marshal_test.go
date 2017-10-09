package osm

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestMarshal_Node(t *testing.T) {
	c := loadChange(t, "testdata/changeset_38162210.osc")
	n := c.Create.Nodes[12]

	// verify it's a good test
	if len(n.Tags) == 0 {
		t.Fatalf("test should have some tags")
	}

	o := &OSM{}
	o.Append(n)

	checkMarshal(t, o)

	// with committed at
	tp := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	o.Nodes[0].Committed = &tp

	checkMarshal(t, o)
}

func TestMarshal_NodeRoundoff(t *testing.T) {
	c := loadChange(t, "testdata/changeset_38162210.osc")
	n := c.Create.Nodes[194]

	o := &OSM{}
	o.Append(n)

	checkMarshal(t, o)
}

func TestMarshal_Nodes(t *testing.T) {
	c := loadChange(t, "testdata/changeset_38162210.osc")
	ns1 := c.Create.Nodes

	o := &OSM{Nodes: ns1}
	checkMarshal(t, o)

	// nodes with no tags
	for _, n := range o.Nodes {
		n.Tags = nil
	}

	checkMarshal(t, o)
}

func TestMarshal_Way(t *testing.T) {
	c := loadChange(t, "testdata/changeset_38162210.osc")
	w := c.Create.Ways[5]

	// verify it's a good test
	if len(w.Tags) == 0 {
		t.Fatalf("test should have some tags")
	}

	o := &OSM{}
	o.Append(w)

	checkMarshal(t, o)
}

func TestMarshal_WayUpdates(t *testing.T) {
	o := loadOSM(t, "testdata/way-updates.osm")
	checkMarshal(t, o)

	// with no updates
	o.Ways[0].Updates = nil
	checkMarshal(t, o)
}

func TestMarshal_Relation(t *testing.T) {
	c := loadChange(t, "testdata/changeset_38162206.osc")
	r := c.Create.Relations[0]

	// verify it's a good test
	if len(r.Tags) == 0 {
		t.Fatalf("test should have some tags")
	}

	o := &OSM{}
	o.Append(r)

	checkMarshal(t, o)
}

func TestMarshal_RelationUpdates(t *testing.T) {
	o := loadOSM(t, "testdata/relation-updates.osm")
	checkMarshal(t, o)

	// with no updates
	o.Relations[0].Updates = nil
	checkMarshal(t, o)
}

func TestMarshal_RelationMemberLocation(t *testing.T) {
	o := &OSM{
		Relations: Relations{
			{
				ID: 123,
				Members: Members{
					{Type: TypeNode, Ref: 1, Version: 2, Lat: 3, Lon: 4},
					{Type: TypeWay, Ref: 2, Version: 3, Lat: 4, Lon: 5},
				},
			},
		},
	}

	checkMarshal(t, o)
}

func TestProtobufRelation_Orientation(t *testing.T) {
	o := &OSM{
		Relations: Relations{
			{
				ID: 123,
				Members: Members{
					{Type: TypeNode, Ref: 1, Version: 2, Orientation: 1},
					{Type: TypeWay, Ref: 2, Version: 3, Orientation: 2},
				},
				Updates: Updates{
					{Index: 0, Reverse: true},
				},
			},
		},
	}

	checkMarshal(t, o)
}

func BenchmarkMarshalXML(b *testing.B) {
	filename := "testdata/changeset_38162206.osc"
	data := readFile(b, filename)

	c := &Change{}
	err := xml.Unmarshal(data, c)
	if err != nil {
		b.Fatalf("unable to unmarshal: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := xml.Marshal(c)
		if err != nil {
			b.Fatalf("unable to marshal: %v", err)
		}
	}
}

func BenchmarkMarshalProto(b *testing.B) {
	cs := &Changeset{
		ID:     38162206,
		UserID: 2744209,
		User:   "grah735",
		Change: loadChange(b, "testdata/changeset_38162206.osc"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := cs.Marshal()
		if err != nil {
			b.Fatalf("unable to marshal: %v", err)
		}
	}
}

func BenchmarkMarshalWayUpdates(b *testing.B) {
	o := loadOSM(b, "testdata/way-updates.osm")
	ways := o.Ways

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data, _ := ways.Marshal()
		UnmarshalWays(data)
	}
}

func BenchmarkMarshalRelationUpdates(b *testing.B) {
	o := loadOSM(b, "testdata/relation-updates.osm")
	relations := o.Relations

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data, _ := relations.Marshal()
		UnmarshalRelations(data)
	}
}

func BenchmarkMarshalProtoGZIP(b *testing.B) {
	cs := &Changeset{
		ID:     38162206,
		UserID: 2744209,
		User:   "grah735",
		Change: loadChange(b, "testdata/changeset_38162206.osc"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data, err := cs.Marshal()
		if err != nil {
			b.Fatalf("unable to marshal: %v", err)
		}

		w, _ := gzip.NewWriterLevel(&bytes.Buffer{}, gzip.BestCompression)
		_, err = w.Write(data)
		if err != nil {
			b.Fatalf("unable to write: %v", err)
		}

		err = w.Close()
		if err != nil {
			b.Fatalf("unable to close: %v", err)
		}
	}
}

func BenchmarkUnmarshalXML(b *testing.B) {
	filename := "testdata/changeset_38162206.osc"
	data := readFile(b, filename)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c := &Change{}
		err := xml.Unmarshal(data, c)
		if err != nil {
			b.Fatalf("unable to unmarshal: %v", err)
		}
	}
}

func BenchmarkUnmarshalProto(b *testing.B) {
	cs := &Changeset{
		ID:     38162206,
		UserID: 2744209,
		User:   "grah735",
		Change: loadChange(b, "testdata/changeset_38162206.osc"),
	}

	data, err := cs.Marshal()
	if err != nil {
		b.Fatalf("unable to marshal: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := UnmarshalChangeset(data)
		if err != nil {
			b.Fatalf("unable to unmarshal: %v", err)
		}
	}
}
func BenchmarkUnmarshalProtoGZIP(b *testing.B) {
	cs := &Changeset{
		ID:     38162206,
		UserID: 2744209,
		User:   "grah735",
		Change: loadChange(b, "testdata/changeset_38162206.osc"),
	}

	data, err := cs.Marshal()
	if err != nil {
		b.Fatalf("unable to marshal: %v", err)
	}

	buff := &bytes.Buffer{}
	w, _ := gzip.NewWriterLevel(buff, gzip.BestCompression)
	_, err = w.Write(data)
	if err != nil {
		b.Fatalf("unable to write: %v", err)
	}

	err = w.Close()
	if err != nil {
		b.Fatalf("unable to close: %v", err)
	}

	data = buff.Bytes()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, _ := gzip.NewReader(bytes.NewReader(data))
		ungzipped, _ := ioutil.ReadAll(r)

		_, err := UnmarshalChangeset(ungzipped)
		if err != nil {
			b.Fatalf("unable to unmarshal: %v", err)
		}
	}
}

func loadChange(t testing.TB, filename string) *Change {
	data := readFile(t, filename)

	c := &Change{}
	err := xml.Unmarshal(data, &c)
	if err != nil {
		t.Fatalf("unable to unmarshal %s: %v", filename, err)
	}

	cleanXMLNameFromChange(c)
	return c
}

func loadOSM(t testing.TB, filename string) *OSM {
	data := readFile(t, filename)

	o := &OSM{}
	err := xml.Unmarshal(data, &o)
	if err != nil {
		t.Fatalf("unable to unmarshal %s: %v", filename, err)
	}

	cleanXMLNameFromOSM(o)
	return o
}

func readFile(t testing.TB, filename string) []byte {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("unable to open %s: %v", filename, err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("unable to read file %s: %v", filename, err)
	}

	return data
}

func checkMarshal(t testing.TB, o1 *OSM) *OSM {
	t.Helper()

	data, err := o1.Marshal()
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	o2, err := UnmarshalOSM(data)
	if err != nil {
		t.Fatalf("unable to unmarshal: %v", err)
	}

	if !reflect.DeepEqual(o1, o2) {
		t.Errorf("results not equal")

		// verify nodes
		ns1 := o1.Nodes
		ns2 := o2.Nodes
		if len(ns1) != len(ns2) {
			t.Fatalf("different number of nodes: %d != %d", len(ns1), len(ns2))
		}

		for i := range ns1 {
			if !reflect.DeepEqual(ns1[i], ns2[i]) {
				t.Errorf("nodes %d are not equal", i)
				t.Logf("%+v", ns1[i])
				t.Logf("%+v", ns2[i])
			}
		}

		// verify ways
		ws1 := o1.Ways
		ws2 := o2.Ways
		if len(ws1) != len(ws2) {
			t.Fatalf("different number of ways: %d != %d", len(ws1), len(ws2))
		}

		for i := range ws1 {
			if !reflect.DeepEqual(ws1[i], ws2[i]) {
				t.Errorf("ways %d are not equal", i)
				t.Logf("%+v", ws1[i])
				t.Logf("%+v", ws2[i])
			}
		}

		// verify relations
		rs1 := o1.Relations
		rs2 := o2.Relations
		if len(rs1) != len(rs2) {
			t.Fatalf("different number of ways: %d != %d", len(rs1), len(rs2))
		}

		for i := range ws1 {
			if !reflect.DeepEqual(rs1[i], rs2[i]) {
				t.Errorf("relations %d are not equal", i)
				t.Logf("%+v", rs1[i])
				t.Logf("%+v", rs2[i])
			}
		}
	}

	return o2
}
