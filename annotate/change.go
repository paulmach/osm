package annotate

import (
	"context"

	"github.com/onXmaps/osm"
	"github.com/onXmaps/osm/annotate/internal/core"
)

// Change will annotate a change into a diff. It will use the
// HistoryDatasourcer to figure out the previous version of all
// the elements and build the diff.
// The IgnoreMissingChildren option can be used to handle missing histories.
// In this case the change will be considered a "new" action type.
func Change(
	ctx context.Context,
	change *osm.Change,
	ds osm.HistoryDatasourcer,
	opts ...Option,
) (*osm.Diff, error) {
	computeOpts := &core.Options{}
	for _, o := range opts {
		err := o(computeOpts)
		if err != nil {
			return nil, err
		}
	}
	ignoreMissing := computeOpts.IgnoreMissingChildren

	actions := make([]osm.Action, 0, osmCount(change.Create)+osmCount(change.Modify)+osmCount(change.Delete))

	// creates are all "new" things
	if o := change.Create; o != nil {
		for _, n := range o.Nodes {
			n.Visible = true
			actions = append(actions, osm.Action{
				Type: osm.ActionCreate,
				OSM:  &osm.OSM{Nodes: osm.Nodes{n}},
			})
		}

		for _, w := range o.Ways {
			w.Visible = true
			actions = append(actions, osm.Action{
				Type: osm.ActionCreate,
				OSM:  &osm.OSM{Ways: osm.Ways{w}},
			})
		}

		for _, r := range o.Relations {
			r.Visible = true
			actions = append(actions, osm.Action{
				Type: osm.ActionCreate,
				OSM:  &osm.OSM{Relations: osm.Relations{r}},
			})
		}
	}

	// modify
	actions, err := addUpdate(ctx, actions, change.Modify, osm.ActionModify, ds, ignoreMissing)
	if err != nil {
		return nil, err
	}

	// delete
	actions, err = addUpdate(ctx, actions, change.Delete, osm.ActionDelete, ds, ignoreMissing)
	if err != nil {
		return nil, err
	}

	return &osm.Diff{Actions: actions}, nil
}

func addUpdate(
	ctx context.Context,
	actions []osm.Action,
	o *osm.OSM,
	actionType osm.ActionType,
	ds osm.HistoryDatasourcer,
	ignoreMissing bool,
) ([]osm.Action, error) {
	if o == nil {
		return actions, nil
	}

	currentVisible := true
	if actionType == osm.ActionDelete {
		currentVisible = false
	}

	for _, n := range o.Nodes {
		old, err := findPreviousNode(ctx, n, ds, ignoreMissing)
		if e := checkErr(ds, ignoreMissing, err, n.FeatureID()); e != nil {
			return nil, e
		}

		if old == nil {
			n.Visible = true
			actions = append(actions, osm.Action{
				Type: osm.ActionCreate,
				OSM:  &osm.OSM{Nodes: osm.Nodes{n}},
			})
			continue
		}

		n.Visible = currentVisible
		actions = append(actions, osm.Action{
			Type: actionType,
			Old:  &osm.OSM{Nodes: osm.Nodes{old}},
			New:  &osm.OSM{Nodes: osm.Nodes{n}},
		})
	}

	for _, w := range o.Ways {
		old, err := findPreviousWay(ctx, w, ds, ignoreMissing)
		if e := checkErr(ds, ignoreMissing, err, w.FeatureID()); e != nil {
			return nil, e
		}

		if old == nil {
			w.Visible = true
			actions = append(actions, osm.Action{
				Type: osm.ActionCreate,
				OSM:  &osm.OSM{Ways: osm.Ways{w}},
			})
			continue
		}

		w.Visible = currentVisible
		actions = append(actions, osm.Action{
			Type: actionType,
			Old:  &osm.OSM{Ways: osm.Ways{old}},
			New:  &osm.OSM{Ways: osm.Ways{w}},
		})
	}

	for _, r := range o.Relations {
		old, err := findPreviousRelation(ctx, r, ds, ignoreMissing)
		if e := checkErr(ds, ignoreMissing, err, r.FeatureID()); e != nil {
			return nil, e
		}

		if old == nil {
			r.Visible = true
			actions = append(actions, osm.Action{
				Type: osm.ActionCreate,
				OSM:  &osm.OSM{Relations: osm.Relations{r}},
			})
			continue
		}

		r.Visible = currentVisible
		actions = append(actions, osm.Action{
			Type: actionType,
			Old:  &osm.OSM{Relations: osm.Relations{old}},
			New:  &osm.OSM{Relations: osm.Relations{r}},
		})
	}

	return actions, nil
}

func osmCount(o *osm.OSM) int {
	if o == nil {
		return 0
	}

	return len(o.Nodes) + len(o.Ways) + len(o.Relations)
}

func checkErr(ds osm.HistoryDatasourcer, ignoreMissing bool, err error, id osm.FeatureID) error {
	if err == nil {
		return nil
	}

	if ds.NotFound(err) {
		if ignoreMissing {
			return nil
		}

		return &NoVisibleChildError{ID: id}
	}

	return err
}

func findPreviousNode(
	ctx context.Context,
	n *osm.Node,
	ds osm.HistoryDatasourcer,
	ignoreMissing bool,
) (*osm.Node, error) {
	nodes, err := ds.NodeHistory(ctx, n.ID)
	if err != nil {
		return nil, err
	}

	loc, max := -1, -1
	for i, node := range nodes {
		if v := node.Version; v < n.Version && v > max {
			max = v
			loc = i
		}
	}

	if loc == -1 {
		// no version before ours
		if ignoreMissing {
			return nil, nil
		}
		return nil, &NoVisibleChildError{ID: n.FeatureID()}
	}

	return nodes[loc], nil
}

func findPreviousWay(
	ctx context.Context,
	w *osm.Way,
	ds osm.HistoryDatasourcer,
	ignoreMissing bool,
) (*osm.Way, error) {
	ways, err := ds.WayHistory(ctx, w.ID)
	if err != nil {
		return nil, err
	}

	loc, max := -1, -1
	for i, way := range ways {
		if v := way.Version; v < w.Version && v > max {
			max = v
			loc = i
		}
	}

	if loc == -1 {
		// no version before ours
		if ignoreMissing {
			return nil, nil
		}
		return nil, &NoVisibleChildError{ID: w.FeatureID()}
	}

	return ways[loc], nil
}

func findPreviousRelation(
	ctx context.Context,
	r *osm.Relation,
	ds osm.HistoryDatasourcer,
	ignoreMissing bool,
) (*osm.Relation, error) {
	relations, err := ds.RelationHistory(ctx, r.ID)
	if err != nil {
		return nil, err
	}

	loc, max := -1, -1
	for i, relation := range relations {
		if v := relation.Version; v < r.Version && v > max {
			max = v
			loc = i
		}
	}

	if loc == -1 {
		// no version before ours
		if ignoreMissing {
			return nil, nil
		}
		return nil, &NoVisibleChildError{ID: r.FeatureID()}
	}

	return relations[loc], nil
}
