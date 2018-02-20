package osm

import (
	"encoding/xml"
	"testing"
)

func TestDiff(t *testing.T) {
	data := readFile(t, "testdata/annotated_diff.xml")

	diff := &Diff{}
	err := xml.Unmarshal(data, &diff)
	if err != nil {
		t.Errorf("unable to unmarshal: %v", err)
	}

	if l := len(diff.Actions); l != 1094 {
		t.Fatalf("incorrect number of actions, got %d", l)
	}

	// create way
	if at := diff.Actions[1075].Type; at != ActionCreate {
		t.Errorf("not a create action, %v", at)
	}

	way := diff.Actions[1075].Ways[0]

	if id := way.ID; id != 180669361 {
		t.Errorf("incorrect way id, got %v", id)
	}

	// modify relation
	if at := diff.Actions[1088].Type; at != ActionModify {
		t.Errorf("not a modify action, %v", at)
	}

	oldRelation := diff.Actions[1088].Old.Relations[0]
	newRelation := diff.Actions[1088].New.Relations[0]

	if oldRelation.ID != newRelation.ID {
		t.Errorf("modify diff is not correct")
		t.Logf("old: %v", oldRelation)
		t.Logf("new: %v", newRelation)
	}

	// delete node
	if at := diff.Actions[44].Type; at != ActionDelete {
		t.Fatalf("not a delete action, %v", at)
	}

	oldNode := diff.Actions[44].Old.Nodes[0]
	newNode := diff.Actions[44].New.Nodes[0]

	if oldNode.ID != newNode.ID {
		t.Errorf("delete diff is not correct")
		t.Logf("old: %v", oldNode)
		t.Logf("new: %v", newNode)
	}

	if newNode.Visible {
		t.Errorf("new node must not be visible")
		t.Logf("old: %v", oldNode)
		t.Logf("new: %v", newNode)
	}
}
