package osm

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/paulmach/orb"
)

func TestRelation_ids(t *testing.T) {
	r := Relation{ID: 12, Version: 2}

	if id := r.FeatureID(); id != RelationID(12).FeatureID() {
		t.Errorf("incorrect feature id: %v", id)
	}

	if id := r.ElementID(); id != RelationID(12).ElementID(2) {
		t.Errorf("incorrect element id: %v", id)
	}
}

func TestRelation_MarshalJSON(t *testing.T) {
	r := Relation{
		ID: 123,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`{"type":"relation","id":123,"visible":false,"timestamp":"0001-01-01T00:00:00Z","members":[]}`)) {
		t.Errorf("incorrect json: %v", string(data))
	}

	// with members
	r = Relation{
		ID:      123,
		Members: Members{{Type: "node", Ref: 123, Role: "outer", Version: 1}},
	}

	data, err = json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`{"type":"relation","id":123,"visible":false,"timestamp":"0001-01-01T00:00:00Z","members":[{"type":"node","ref":123,"role":"outer","version":1}]}`)) {
		t.Errorf("incorrect json: %v", string(data))
	}
}

func TestRelation_ApplyUpdatesUpTo(t *testing.T) {
	updates := Updates{
		{Index: 0, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC), Version: 11},
		{Index: 1, Timestamp: time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC), Version: 12},
		{Index: 2, Timestamp: time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC), Version: 13, Lat: 10, Lon: 20},
	}
	r := Relation{
		ID:      123,
		Members: Members{{Version: 1}, {Version: 2}, {Version: 3}},
	}

	r.Updates = updates
	r.ApplyUpdatesUpTo(time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC))
	if r.Members[0].Version != 1 || r.Members[1].Version != 2 || r.Members[2].Version != 3 {
		t.Errorf("incorrect members, got %v", r.Members)
	}

	r.Updates = updates
	r.ApplyUpdatesUpTo(time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC))
	if r.Members[0].Version != 11 || r.Members[1].Version != 2 || r.Members[2].Version != 13 {
		t.Errorf("incorrect members, got %v", r.Members)
	}

	if r.Members[2].Lat != 10 {
		t.Errorf("did not apply lat data")
	}

	if r.Members[2].Lon != 20 {
		t.Errorf("did not apply lon data")
	}

	if l := len(r.Updates); l != 1 {
		t.Errorf("incorrect number of updates: %v", l)
	}

	if r.Updates[0].Index != 1 {
		t.Errorf("incorrect updates: %v", r.Updates)
	}
}

func TestRelation_ApplyUpdate(t *testing.T) {
	r := Relation{
		ID:      123,
		Members: Members{{Ref: 1, Type: TypeWay, Orientation: orb.CW}},
	}

	err := r.applyUpdate(Update{
		Index:       0,
		Version:     1,
		ChangesetID: 2,
		Reverse:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := Member{
		Ref:         1,
		Type:        TypeWay,
		Version:     1,
		ChangesetID: 2,
		Orientation: orb.CCW,
	}

	if !reflect.DeepEqual(r.Members[0], expected) {
		t.Errorf("incorrect update, got %v", r.Members[0])
	}
}

func TestRelation_ApplyUpdate_error(t *testing.T) {
	r := Relation{
		ID:      123,
		Members: Members{{Ref: 1, Type: TypeNode}},
	}

	err := r.applyUpdate(Update{
		Index: 1,
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if e, ok := err.(*UpdateIndexOutOfRangeError); !ok {
		t.Errorf("incorrect error, got %v", e)
	}
}

func TestRelation_MarshalXML(t *testing.T) {
	r := Relation{
		ID: 123,
	}

	data, err := xml.Marshal(r)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	expected := `<relation id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"></relation>`
	if !bytes.Equal(data, []byte(expected)) {
		t.Errorf("incorrect marshal, got: %s", string(data))
	}

	// members
	r.Members = Members{{Type: "node", Ref: 123, Role: "child"}}
	data, err = xml.Marshal(r)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`<relation id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><member type="node" ref="123" role="child"></member></relation>`)) {
		t.Errorf("not marshalled correctly: %s", string(data))
	}

	// members with nodes
	r.Members = Members{{Type: "way", Ref: 123, Role: "child", Nodes: WayNodes{{Lat: 1, Lon: 2}, {Lat: 3, Lon: 4}}}}
	data, err = xml.Marshal(r)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`<relation id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><member type="way" ref="123" role="child"><nd lat="1" lon="2"></nd><nd lat="3" lon="4"></nd></member></relation>`)) {
		t.Errorf("not marshalled correctly: %s", string(data))
	}

	// minor relation
	r.Members = nil
	r.Updates = []Update{
		{
			Index:       0,
			Version:     1,
			ChangesetID: 123,
			Timestamp:   time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	data, err = xml.Marshal(r)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`<relation id="123" user="" uid="0" visible="false" version="0" changeset="0" timestamp="0001-01-01T00:00:00Z"><update index="0" version="1" timestamp="2012-01-01T00:00:00Z" changeset="123"></update></relation>`)) {
		t.Errorf("not marshalled correctly: %s", string(data))
	}

	// blanket xml test
	data, err = os.ReadFile("testdata/relation-updates.osm")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	osm := &OSM{}
	err = xml.Unmarshal(data, &osm)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	relation := osm.Relations[0]

	var i1 interface{}
	err = xml.Unmarshal(data, &i1)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	data, err = xml.Marshal(relation)
	if err != nil {
		t.Errorf("marshal error: %v", err)
	}

	var i2 interface{}
	err = xml.Unmarshal(data, &i2)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(i1, i2) {
		t.Errorf("interfaces not equal")
		t.Logf("%+v", i1)
		t.Logf("%+v", i2)
	}
}

func TestMember_ids(t *testing.T) {
	cases := []struct {
		name string
		m    Member
		fid  FeatureID
		eid  ElementID
	}{
		{
			name: "node",
			m:    Member{Type: TypeNode, Ref: 12, Version: 2},
			fid:  NodeID(12).FeatureID(),
			eid:  NodeID(12).ElementID(2),
		},
		{
			name: "way",
			m:    Member{Type: TypeWay, Ref: 12, Version: 2},
			fid:  WayID(12).FeatureID(),
			eid:  WayID(12).ElementID(2),
		},
		{
			name: "relation",
			m:    Member{Type: TypeRelation, Ref: 12, Version: 2},
			fid:  RelationID(12).FeatureID(),
			eid:  RelationID(12).ElementID(2),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if id := tc.m.FeatureID(); id != tc.fid {
				t.Errorf("incorrect feature id: %v", id)
			}

			if id := tc.m.ElementID(); id != tc.eid {
				t.Errorf("incorrect element id: %v", id)
			}
		})
	}
}
func TestMembers_ids(t *testing.T) {
	ms := Members{
		{Type: TypeNode, Ref: 1, Version: 3},
		{Type: TypeWay, Ref: 2, Version: 4},
		{Type: TypeRelation, Ref: 3, Version: 5},
	}

	eids := ElementIDs{
		NodeID(1).ElementID(3),
		WayID(2).ElementID(4),
		RelationID(3).ElementID(5),
	}
	if ids := ms.ElementIDs(); !reflect.DeepEqual(ids, eids) {
		t.Errorf("incorrect element ids: %v", ids)
	}

	fids := FeatureIDs{
		NodeID(1).FeatureID(),
		WayID(2).FeatureID(),
		RelationID(3).FeatureID(),
	}
	if ids := ms.FeatureIDs(); !reflect.DeepEqual(ids, fids) {
		t.Errorf("incorrect feature ids: %v", ids)
	}
}

func TestRelations_ids(t *testing.T) {
	rs := Relations{
		{ID: 1, Version: 3},
		{ID: 2, Version: 4},
	}

	eids := ElementIDs{RelationID(1).ElementID(3), RelationID(2).ElementID(4)}
	if ids := rs.ElementIDs(); !reflect.DeepEqual(ids, eids) {
		t.Errorf("incorrect element ids: %v", ids)
	}

	fids := FeatureIDs{RelationID(1).FeatureID(), RelationID(2).FeatureID()}
	if ids := rs.FeatureIDs(); !reflect.DeepEqual(ids, fids) {
		t.Errorf("incorrect feature ids: %v", ids)
	}

	rids := []RelationID{1, 2}
	if ids := rs.IDs(); !reflect.DeepEqual(ids, rids) {
		t.Errorf("incorrect node id: %v", rids)
	}
}

func TestRelations_SortByIDVersion(t *testing.T) {
	rs := Relations{
		{ID: 7, Version: 3},
		{ID: 2, Version: 4},
		{ID: 5, Version: 2},
		{ID: 5, Version: 3},
		{ID: 5, Version: 4},
		{ID: 3, Version: 4},
		{ID: 4, Version: 4},
		{ID: 9, Version: 4},
	}

	rs.SortByIDVersion()

	eids := ElementIDs{
		RelationID(2).ElementID(4),
		RelationID(3).ElementID(4),
		RelationID(4).ElementID(4),
		RelationID(5).ElementID(2),
		RelationID(5).ElementID(3),
		RelationID(5).ElementID(4),
		RelationID(7).ElementID(3),
		RelationID(9).ElementID(4),
	}

	if ids := rs.ElementIDs(); !reflect.DeepEqual(ids, eids) {
		t.Errorf("incorrect sort: %v", eids)
	}
}
