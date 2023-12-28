package osm

import (
	"encoding/xml"
	"os"
	"reflect"
	"testing"
)

func TestDiff_MarshalXML(t *testing.T) {
	data := []byte(`<osm>
 <action type="delete">
  <old>
   <node id="1896619025" lat="0" lon="0" user="" uid="0" visible="true" version="2" changeset="0" timestamp="0001-01-01T00:00:00Z"></node>
  </old>
  <new>
   <node id="1896619025" lat="0" lon="0" user="" uid="0" visible="false" version="3" changeset="0" timestamp="0001-01-01T00:00:00Z"></node>
  </new>
 </action>
 <action type="create">
  <node id="1911156719" lat="0" lon="0" user="" uid="0" visible="false" version="1" changeset="0" timestamp="0001-01-01T00:00:00Z"></node>
 </action>
</osm>`)

	diff := &Diff{}
	err := xml.Unmarshal(data, &diff)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	if l := len(diff.Actions); l != 2 {
		t.Errorf("incorrect num of actions: %v", l)
	}

	marshalled, err := xml.MarshalIndent(diff, "", " ")
	if err != nil {
		t.Errorf("marshal error: %v", err)
	}

	if !reflect.DeepEqual(marshalled, data) {
		t.Errorf("incorrect marshal")
		t.Logf("%v", string(marshalled))
		t.Logf("%v", string(data))
	}

	// specifics
	diff = &Diff{}
	_, err = xml.Marshal(diff)
	if err != nil {
		t.Errorf("unable to marshal: %v", err)
	}

	// create
	diff.Actions = append(diff.Actions, Action{
		Type: ActionCreate,
		OSM:  &OSM{Nodes: Nodes{{ID: 1}}},
	})
	_, err = xml.Marshal(diff)
	if err != nil {
		t.Errorf("unable to marshal: %v", err)
	}

	// modify
	diff.Actions = append(diff.Actions, Action{
		Type: ActionModify,
		Old:  &OSM{Nodes: Nodes{{ID: 1}}},
		New:  &OSM{Nodes: Nodes{{ID: 1}}},
	})
	_, err = xml.Marshal(diff)
	if err != nil {
		t.Errorf("unable to marshal: %v", err)
	}
}

func TestDiff(t *testing.T) {
	data, err := os.ReadFile("testdata/annotated_diff.xml")
	if err != nil {
		t.Fatalf("unable to read file: %v", err)
	}

	diff := &Diff{}
	err = xml.Unmarshal(data, &diff)
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

	// should marshal the unmarshalled data
	_, err = xml.Marshal(diff)
	if err != nil {
		t.Errorf("unable to marshal: %v", err)
	}
}

func BenchmarkDiff_Marshal(b *testing.B) {
	data, err := os.ReadFile("testdata/annotated_diff.xml")
	if err != nil {
		b.Fatalf("unable to read file: %v", err)
	}

	diff := &Diff{}
	err = xml.Unmarshal(data, &diff)
	if err != nil {
		b.Fatalf("unmarshal error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := xml.Marshal(diff)
		if err != nil {
			b.Fatalf("marshal error: %v", err)
		}
	}
}

func BenchmarkDiff_Unmarshal(b *testing.B) {
	data, err := os.ReadFile("testdata/annotated_diff.xml")
	if err != nil {
		b.Fatalf("unable to read file: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diff := &Diff{}
		err := xml.Unmarshal(data, &diff)
		if err != nil {
			b.Fatalf("unmarshal error: %v", err)
		}
	}
}
