package annotate

import (
	"context"
	"testing"

	"github.com/onXmaps/osm"
)

func TestChange_create(t *testing.T) {
	ctx := context.Background()

	ds := (&osm.OSM{}).HistoryDatasource()

	t.Run("new node", func(t *testing.T) {
		change := &osm.Change{
			Create: &osm.OSM{
				Nodes: osm.Nodes{{ID: 3, Version: 1}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionCreate {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Nodes[0].ID; v != 3 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Nodes[0].Version; v != 1 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Nodes[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}
	})

	t.Run("new way", func(t *testing.T) {
		change := &osm.Change{
			Create: &osm.OSM{
				Ways: osm.Ways{{ID: 3, Version: 1}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionCreate {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Ways[0].ID; v != 3 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Ways[0].Version; v != 1 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Ways[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}
	})

	t.Run("new relation", func(t *testing.T) {
		change := &osm.Change{
			Create: &osm.OSM{
				Relations: osm.Relations{{ID: 3, Version: 1}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionCreate {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Relations[0].ID; v != 3 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Relations[0].Version; v != 1 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Relations[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}
	})
}

func TestChange_modify(t *testing.T) {
	ctx := context.Background()

	ds := (&osm.OSM{
		Nodes: osm.Nodes{
			{ID: 1, Version: 1, Visible: true},
			{ID: 1, Version: 2, Visible: true},
			{ID: 5, Version: 1, Visible: true},
			{ID: 5, Version: 2, Visible: false},
		},
		Ways: osm.Ways{
			{ID: 2, Version: 1, Visible: true},
			{ID: 2, Version: 2, Visible: true},
		},
		Relations: osm.Relations{
			{ID: 3, Version: 1, Visible: true},
			{ID: 3, Version: 2, Visible: true},
		},
	}).HistoryDatasource()

	t.Run("undelete a node", func(t *testing.T) {
		change := &osm.Change{
			Modify: &osm.OSM{
				Nodes: osm.Nodes{{ID: 5, Version: 3, Visible: true}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionModify {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Old.Nodes[0].ID; v != 5 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Old.Nodes[0].Version; v != 2 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Old.Nodes[0].Visible; v {
			t.Errorf("incorrect visible: %v", v)
		}

		if v := a.New.Nodes[0].ID; v != 5 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.New.Nodes[0].Version; v != 3 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.New.Nodes[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}
	})

	t.Run("new node version", func(t *testing.T) {
		change := &osm.Change{
			Modify: &osm.OSM{
				Nodes: osm.Nodes{{ID: 1, Version: 3}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionModify {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Old.Nodes[0].ID; v != 1 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Old.Nodes[0].Version; v != 2 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Old.Nodes[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}

		if v := a.New.Nodes[0].ID; v != 1 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.New.Nodes[0].Version; v != 3 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.New.Nodes[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}
	})

	t.Run("new way version", func(t *testing.T) {
		change := &osm.Change{
			Modify: &osm.OSM{
				Ways: osm.Ways{{ID: 2, Version: 3}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionModify {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Old.Ways[0].ID; v != 2 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Old.Ways[0].Version; v != 2 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Old.Ways[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}

		if v := a.New.Ways[0].ID; v != 2 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.New.Ways[0].Version; v != 3 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.New.Ways[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}
	})

	t.Run("new relation version", func(t *testing.T) {
		change := &osm.Change{
			Modify: &osm.OSM{
				Relations: osm.Relations{{ID: 3, Version: 3}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionModify {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Old.Relations[0].ID; v != 3 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Old.Relations[0].Version; v != 2 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Old.Relations[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}

		if v := a.New.Relations[0].ID; v != 3 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.New.Relations[0].Version; v != 3 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.New.Relations[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}
	})
}

func TestChange_delete(t *testing.T) {
	ctx := context.Background()

	ds := (&osm.OSM{
		Nodes: osm.Nodes{
			{ID: 1, Version: 1, Visible: true},
			{ID: 1, Version: 2, Visible: true},
		},
		Ways: osm.Ways{
			{ID: 2, Version: 1, Visible: true},
			{ID: 2, Version: 2, Visible: true},
		},
		Relations: osm.Relations{
			{ID: 3, Version: 1, Visible: true},
			{ID: 3, Version: 2, Visible: true},
		},
	}).HistoryDatasource()

	t.Run("delete node", func(t *testing.T) {
		change := &osm.Change{
			Delete: &osm.OSM{
				Nodes: osm.Nodes{{ID: 1, Version: 3, Visible: false}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionDelete {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Old.Nodes[0].ID; v != 1 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Old.Nodes[0].Version; v != 2 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Old.Nodes[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}

		if v := a.New.Nodes[0].ID; v != 1 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.New.Nodes[0].Version; v != 3 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.New.Nodes[0].Visible; v {
			t.Errorf("incorrect visible: %v", v)
		}
	})

	t.Run("delete way", func(t *testing.T) {
		change := &osm.Change{
			Delete: &osm.OSM{
				Ways: osm.Ways{{ID: 2, Version: 3, Visible: false}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionDelete {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Old.Ways[0].ID; v != 2 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Old.Ways[0].Version; v != 2 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Old.Ways[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}

		if v := a.New.Ways[0].ID; v != 2 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.New.Ways[0].Version; v != 3 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.New.Ways[0].Visible; v {
			t.Errorf("incorrect visible: %v", v)
		}
	})

	t.Run("delete relation", func(t *testing.T) {
		change := &osm.Change{
			Delete: &osm.OSM{
				Relations: osm.Relations{{ID: 3, Version: 3, Visible: false}},
			},
		}

		diff, err := Change(ctx, change, ds)
		if err != nil {
			t.Fatalf("change error: %v", err)
		}

		a := diff.Actions[0]
		if a.Type != osm.ActionDelete {
			t.Errorf("invalid type: %v", a.Type)
		}

		if v := a.Old.Relations[0].ID; v != 3 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.Old.Relations[0].Version; v != 2 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.Old.Relations[0].Visible; !v {
			t.Errorf("incorrect visible: %v", v)
		}

		if v := a.New.Relations[0].ID; v != 3 {
			t.Errorf("incorrect id: %v", v)
		}

		if v := a.New.Relations[0].Version; v != 3 {
			t.Errorf("incorrect version: %v", v)
		}

		if v := a.New.Relations[0].Visible; v {
			t.Errorf("incorrect visible: %v", v)
		}
	})
}
