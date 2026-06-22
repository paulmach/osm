package osmpbf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"
	"testing"

	"github.com/paulmach/osm"
)

func TestEncodeDecode(t *testing.T) {
	buffer := bytes.Buffer{}
	writer, err := NewEncoder(&buffer)
	if err != nil {
		t.Fatal(err)
	}

	// write node
	err = writer.Encode(en)
	if err != nil {
		t.Fatal(err)
	}
	// write way
	err = writer.Encode(ew)
	if err != nil {
		t.Fatal(err)
	}
	// write relation
	err = writer.Encode(er)
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()

	d := newDecoder(context.Background(), &Scanner{}, &buffer)
	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		t.Fatal(err)
	}

	for {
		e, err := d.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}

		switch v := e.(type) {
		case *osm.Node:
			err = nodeEquals(en, v)
			if err != nil {
				t.Fatal(err)
			}
		case *osm.Way:
			err = wayEquals(ew, v)
			if err != nil {
				t.Fatal(err)
			}
		case *osm.Relation:
			err = relationEquals(er, v)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	d.Close()
}

func nodeEquals(en, node *osm.Node) error {
	if node.ID != en.ID {
		return fmt.Errorf("node id mismatch: %d != %d", node.ID, en.ID)
	}
	if node.Lat != en.Lat {
		return fmt.Errorf("node lat mismatch: %f != %f", node.Lat, en.Lat)
	}
	if node.Lon != en.Lon {
		return fmt.Errorf("node lon mismatch: %f != %f", node.Lon, en.Lon)
	}
	if node.User != en.User {
		return fmt.Errorf("node user mismatch: %s != %s", node.User, en.User)
	}
	if node.UserID != en.UserID {
		return fmt.Errorf("node user id mismatch: %d != %d", node.UserID, en.UserID)
	}
	if node.Visible != en.Visible {
		return fmt.Errorf("node visible mismatch: %v != %v", node.Visible, en.Visible)
	}
	if node.Version != en.Version {
		return fmt.Errorf("node version mismatch: %d != %d", node.Version, en.Version)
	}
	if node.ChangesetID != en.ChangesetID {
		return fmt.Errorf("node changeset id mismatch: %d != %d", node.ChangesetID, en.ChangesetID)
	}
	if node.Timestamp != en.Timestamp {
		return fmt.Errorf("node timestamp mismatch: %s != %s", node.Timestamp, en.Timestamp)
	}
	if len(node.Tags) != len(en.Tags) {
		return fmt.Errorf("node tags length mismatch: %d != %d", len(node.Tags), len(en.Tags))
	}
	for k, v := range node.Tags {
		if en.Tags[k] != v {
			return fmt.Errorf("node tag mismatch: %s != %s", en.Tags[k], v)
		}
	}
	return nil
}

func wayEquals(ew, way *osm.Way) error {
	if way.ID != ew.ID {
		return fmt.Errorf("way id mismatch: %d != %d", way.ID, ew.ID)
	}
	if way.User != ew.User {
		return fmt.Errorf("way user mismatch: %s != %s", way.User, ew.User)
	}
	if way.UserID != ew.UserID {
		return fmt.Errorf("way user id mismatch: %d != %d", way.UserID, ew.UserID)
	}
	if way.Visible != ew.Visible {
		return fmt.Errorf("way visible mismatch: %v != %v", way.Visible, ew.Visible)
	}
	if way.Version != ew.Version {
		return fmt.Errorf("way version mismatch: %d != %d", way.Version, ew.Version)
	}
	if way.ChangesetID != ew.ChangesetID {
		return fmt.Errorf("way changeset id mismatch: %d != %d", way.ChangesetID, ew.ChangesetID)
	}
	if way.Timestamp != ew.Timestamp {
		return fmt.Errorf("way timestamp mismatch: %s != %s", way.Timestamp, ew.Timestamp)
	}
	if len(way.Tags) != len(ew.Tags) {
		return fmt.Errorf("way tags length mismatch: %d != %d", len(way.Tags), len(ew.Tags))
	}
	for k, v := range way.Tags {
		if ew.Tags[k] != v {
			return fmt.Errorf("way tag mismatch: %s != %s", ew.Tags[k], v)
		}
	}
	if len(way.Nodes) != len(ew.Nodes) {
		return fmt.Errorf("way nodes length mismatch: %d != %d", len(way.Nodes), len(ew.Nodes))
	}
	for i, v := range way.Nodes {
		if ew.Nodes[i] != v {
			return fmt.Errorf("way node mismatch: %v != %v", ew.Nodes[i], v)
		}
	}
	return nil
}

func relationEquals(er, relation *osm.Relation) error {
	if relation.ID != er.ID {
		return fmt.Errorf("relation id mismatch: %d != %d", relation.ID, er.ID)
	}
	if relation.User != er.User {
		return fmt.Errorf("relation user mismatch: %s != %s", relation.User, er.User)
	}
	if relation.UserID != er.UserID {
		return fmt.Errorf("relation user id mismatch: %d != %d", relation.UserID, er.UserID)
	}
	if relation.Visible != er.Visible {
		return fmt.Errorf("relation visible mismatch: %v != %v", relation.Visible, er.Visible)
	}
	if relation.Version != er.Version {
		return fmt.Errorf("relation version mismatch: %d != %d", relation.Version, er.Version)
	}
	if relation.ChangesetID != er.ChangesetID {
		return fmt.Errorf("relation changeset id mismatch: %d != %d", relation.ChangesetID, er.ChangesetID)
	}
	if relation.Timestamp != er.Timestamp {
		return fmt.Errorf("relation timestamp mismatch: %s != %s", relation.Timestamp, er.Timestamp)
	}
	if len(relation.Tags) != len(er.Tags) {
		return fmt.Errorf("relation tags length mismatch: %d != %d", len(relation.Tags), len(er.Tags))
	}
	for k, v := range relation.Tags {
		if er.Tags[k] != v {
			return fmt.Errorf("relation tag mismatch: %s != %s", er.Tags[k], v)
		}
	}
	if len(relation.Members) != len(er.Members) {
		return fmt.Errorf("relation members length mismatch: %d != %d", len(relation.Members), len(er.Members))
	}
	for i, member := range relation.Members {
		expectedMember := er.Members[i]
		if member.ChangesetID != expectedMember.ChangesetID {
			return fmt.Errorf("relation member changeset id mismatch: %d != %d", member.ChangesetID, expectedMember.ChangesetID)
		}
		if member.Role != expectedMember.Role {
			return fmt.Errorf("relation member role mismatch: %s != %s", member.Role, expectedMember.Role)
		}
		if member.Type != expectedMember.Type {
			return fmt.Errorf("relation member type mismatch: %s != %s", member.Type, expectedMember.Type)
		}
		if member.Ref != expectedMember.Ref {
			return fmt.Errorf("relation member ref mismatch: %d != %d", member.Ref, expectedMember.Ref)
		}
		if member.Lat != expectedMember.Lat {
			return fmt.Errorf("relation member lat mismatch: %f != %f", member.Lat, expectedMember.Lat)
		}
		if member.Lon != expectedMember.Lon {
			return fmt.Errorf("relation member lon mismatch: %f != %f", member.Lon, expectedMember.Lon)
		}
		if member.Orientation != expectedMember.Orientation {
			return fmt.Errorf("relation member orientation mismatch: %d != %d", member.Orientation, expectedMember.Orientation)
		}
		if member.Version != expectedMember.Version {
			return fmt.Errorf("relation member version mismatch: %d != %d", member.Version, expectedMember.Version)
		}
		if len(member.Nodes) != len(expectedMember.Nodes) {
			return fmt.Errorf("relation member nodes length mismatch: %d != %d", len(member.Nodes), len(expectedMember.Nodes))
		}
		for j, node := range member.Nodes {
			if expectedMember.Nodes[j] != node {
				return fmt.Errorf("relation member node mismatch: %v != %v", expectedMember.Nodes[j], node)
			}
		}
	}
	return nil
}
