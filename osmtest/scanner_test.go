package osmtest

import (
	"errors"
	"reflect"
	"testing"

	"github.com/onXmaps/osm"
)

func TestScanner(t *testing.T) {
	objs := osm.Objects{
		&osm.Node{ID: 1, Version: 4},
		&osm.Way{ID: 2, Version: 5},
		&osm.Relation{ID: 3, Version: 6},
	}

	scanner := NewScanner(objs)
	defer scanner.Close()

	expected := osm.ObjectIDs{
		osm.NodeID(1).ObjectID(4),
		osm.WayID(2).ObjectID(5),
		osm.RelationID(3).ObjectID(6),
	}

	ids := osm.ObjectIDs{}
	for scanner.Scan() {
		ids = append(ids, scanner.Object().ObjectID())
	}

	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("incorrect ids: %v", ids)
	}
}

func TestScanner_error(t *testing.T) {
	objs := osm.Objects{
		&osm.Node{ID: 1, Version: 4},
		&osm.Way{ID: 2, Version: 5},
		&osm.Relation{ID: 3, Version: 6},
	}

	scanner := NewScanner(objs)
	defer scanner.Close()

	if scanner.Err() != nil {
		t.Errorf("error should not be set initially")
	}

	if !scanner.Scan() {
		t.Errorf("should be true initially")
	}

	scanner.ScanError = errors.New("some error")

	if scanner.Scan() {
		t.Errorf("should be false after error")
	}

	if scanner.Err() == nil {
		t.Errorf("should return error if there is one")
	}
}
