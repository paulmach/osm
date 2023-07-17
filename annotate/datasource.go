package annotate

import (
	"context"

	"github.com/onXmaps/osm"
	"github.com/onXmaps/osm/annotate/internal/core"
	"github.com/onXmaps/osm/annotate/shared"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
)

type wayDatasource struct {
	NodeHistoryDatasourcer
}

type wayChildDatasource struct {
	NodeHistoryAsChildrenDatasourcer
}

func newWayDatasourcer(ds NodeHistoryDatasourcer) core.Datasourcer {
	if d, ok := ds.(NodeHistoryAsChildrenDatasourcer); ok {
		return &wayChildDatasource{d}
	}

	return &wayDatasource{ds}
}

func (wds *wayDatasource) Get(ctx context.Context, id osm.FeatureID) (core.ChildList, error) {
	if id.Type() != osm.TypeNode {
		panic("only node types supported")
	}

	nodes, err := wds.NodeHistory(ctx, id.NodeID())
	if err != nil {
		return nil, err
	}

	return nodesToChildList(nodes), nil
}

func (wds *wayChildDatasource) Get(ctx context.Context, id osm.FeatureID) (core.ChildList, error) {
	if id.Type() != osm.TypeNode {
		panic("only node types supported")
	}

	return wds.NodeHistoryAsChildren(ctx, id.NodeID())
}

type relationDatasource struct {
	osm.HistoryDatasourcer
}

type relationChildDatasource struct {
	HistoryAsChildrenDatasourcer
}

func newRelationDatasourcer(ds osm.HistoryDatasourcer) core.Datasourcer {
	if d, ok := ds.(HistoryAsChildrenDatasourcer); ok {
		return &relationChildDatasource{d}
	}

	return &relationDatasource{ds}
}

func (rds *relationDatasource) Get(ctx context.Context, id osm.FeatureID) (core.ChildList, error) {

	switch id.Type() {
	case osm.TypeNode:
		nodes, err := rds.NodeHistory(ctx, id.NodeID())
		if err != nil {
			return nil, err
		}

		return nodesToChildList(nodes), nil
	case osm.TypeWay:
		ways, err := rds.WayHistory(ctx, id.WayID())
		if err != nil {
			return nil, err
		}

		return waysToChildList(ways), nil
	case osm.TypeRelation:
		relations, err := rds.RelationHistory(ctx, id.RelationID())
		if err != nil {
			return nil, err
		}

		return relationsToChildList(relations), nil
	}

	return nil, &UnsupportedMemberTypeError{
		MemberType: id.Type(),
	}
}

func (rds *relationChildDatasource) Get(ctx context.Context, id osm.FeatureID) (core.ChildList, error) {

	switch id.Type() {
	case osm.TypeNode:
		return rds.NodeHistoryAsChildren(ctx, id.NodeID())
	case osm.TypeWay:
		return rds.WayHistoryAsChildren(ctx, id.WayID())
	case osm.TypeRelation:
		return rds.RelationHistoryAsChildren(ctx, id.RelationID())
	}

	return nil, &UnsupportedMemberTypeError{
		MemberType: id.Type(),
	}
}

func nodesToChildList(nodes osm.Nodes) core.ChildList {
	if len(nodes) == 0 {
		return nil
	}

	list := make(core.ChildList, len(nodes))
	nodes.SortByIDVersion()
	for i, n := range nodes {
		c := shared.FromNode(n)
		c.VersionIndex = i
		list[i] = c
	}

	return list
}

func waysToChildList(ways osm.Ways) core.ChildList {
	if len(ways) == 0 {
		return nil
	}

	list := make(core.ChildList, len(ways))
	ways.SortByIDVersion()
	for i, w := range ways {
		c := shared.FromWay(w)
		c.VersionIndex = i

		if i != 0 {
			c.ReverseOfPrevious = IsReverse(w, ways[i-1])
		}

		list[i] = c
	}

	return list
}

// IsReverse checks to see if this way update was a "reversal". It is very tricky
// to generally answer this question but easier for a relation minor update.
// Since the relation wasn't updated we assume things are still connected and
// can just check the endpoints.
func IsReverse(w1, w2 *osm.Way) bool {
	if len(w1.Nodes) < 2 || len(w2.Nodes) < 2 {
		return false
	}

	// check if either is a ring
	if w1.Nodes[0].ID == w1.Nodes[len(w1.Nodes)-1].ID ||
		w2.Nodes[0].ID == w2.Nodes[len(w2.Nodes)-1].ID {

		r1 := orb.Ring(w1.LineString())
		r2 := orb.Ring(w2.LineString())
		return planar.Area(r1)*planar.Area(r2) < 0
	}

	// not a ring so see if endpoint were flipped
	return w1.Nodes[0].ID == w2.Nodes[len(w2.Nodes)-1].ID &&
		w2.Nodes[0].ID == w1.Nodes[len(w1.Nodes)-1].ID
}

func relationsToChildList(relations osm.Relations) core.ChildList {
	if len(relations) == 0 {
		return nil
	}

	list := make(core.ChildList, len(relations))
	relations.SortByIDVersion()
	for i, r := range relations {
		c := shared.FromRelation(r)
		c.VersionIndex = i
		list[i] = c
	}

	return list
}
