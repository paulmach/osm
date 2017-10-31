package annotate

import (
	"context"
	"fmt"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/annotate/internal/core"
)

// Relations computes the updates for the given relations
// and annotate members with stuff like changeset and lon/lat data.
// The input relations are modified to include this information.
func Relations(
	ctx context.Context,
	relations osm.Relations,
	datasource osm.HistoryDatasourcer,
	threshold time.Duration,
	opts ...Option,
) error {
	computeOpts := &core.Options{}
	for _, o := range opts {
		err := o(computeOpts)
		if err != nil {
			return err
		}
	}

	parents, children, err := convertRelationData(ctx, relations, datasource, computeOpts.IgnoreMissingChildren)
	if err != nil {
		return mapErrors(err)
	}

	updatesForParents, err := core.Compute(parents, children, threshold, computeOpts)
	if err != nil {
		return mapErrors(err)
	}

	for i, updates := range updatesForParents {
		relations[i].Updates = updates
	}

	return nil
}

func convertRelationData(
	ctx context.Context,
	relations osm.Relations,
	datasource osm.HistoryDatasourcer,
	ignoreNotFound bool,
) ([]core.Parent, *core.Histories, error) {
	relations.SortByIDVersion()

	parents := make([]core.Parent, len(relations))
	histories := &core.Histories{}

	for i, r := range relations {
		parents[i] = &parentRelation{Relation: r}

		for j, m := range r.Members {
			childID := m.FeatureID()
			if histories.Get(childID) != nil {
				continue
			}

			switch childID.Type {
			case osm.TypeNode:
				nodes, err := datasource.NodeHistory(ctx, childID.NodeID())
				if err != nil && (!datasource.NotFound(err) || !ignoreNotFound) {
					return nil, nil, err
				}

				list := nodesToChildList(nodes)
				histories.Set(childID, list)
			case osm.TypeWay:
				ways, err := datasource.WayHistory(ctx, childID.WayID())
				if err != nil && (!datasource.NotFound(err) || !ignoreNotFound) {
					return nil, nil, err
				}

				list := waysToChildList(ways)
				histories.Set(childID, list)
			case osm.TypeRelation:
				relations, err := datasource.RelationHistory(ctx, childID.RelationID())
				if err != nil && (!datasource.NotFound(err) || !ignoreNotFound) {
					return nil, nil, err
				}

				list := relationsToChildList(relations)
				histories.Set(childID, list)
			default:
				return nil, nil, &UnsupportedMemberTypeError{
					RelationID: r.ID,
					Index:      j,
					MemberType: m.Type,
				}
			}
		}
	}

	return parents, histories, nil
}

func waysToChildList(ways osm.Ways) core.ChildList {
	if len(ways) == 0 {
		return nil
	}

	list := make(core.ChildList, len(ways))
	ways.SortByIDVersion()
	for i, w := range ways {
		c := &childWay{
			Index: i,
			Way:   w,
		}

		if i != 0 {
			c.ReverseOfPrevious = isReverse(w, ways[i-1])
		}

		list[i] = c
	}

	return list
}

func isReverse(w1, w2 *osm.Way) bool {
	if len(w1.Nodes) < 2 || len(w2.Nodes) < 2 {
		return false
	}

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
		list[i] = &childRelation{
			Index:    i,
			Relation: r,
		}
	}

	return list
}

// A parentRelation wraps a osm.Relation into the core.Parent interface
// so that updates can be computed.
type parentRelation struct {
	Relation *osm.Relation
	children core.ChildList
}

func (r parentRelation) ID() osm.FeatureID {
	return r.Relation.FeatureID()
}

func (r parentRelation) ChangesetID() osm.ChangesetID {
	return r.Relation.ChangesetID
}

func (r parentRelation) Version() int {
	return r.Relation.Version
}

func (r parentRelation) Visible() bool {
	return r.Relation.Visible
}

func (r parentRelation) Timestamp() time.Time {
	return r.Relation.Timestamp
}

func (r parentRelation) Committed() time.Time {
	if r.Relation.Committed == nil {
		return time.Time{}
	}

	return *r.Relation.Committed
}

func (r parentRelation) Refs() osm.FeatureIDs {
	return r.Relation.Members.FeatureIDs()
}

func (r parentRelation) Children() core.ChildList {
	return r.children
}

func (r *parentRelation) SetChildren(list core.ChildList) {
	r.children = list

	var ways map[osm.WayID]*osm.Way
	if r.Relation.Polygon() {
		ways = make(map[osm.WayID]*osm.Way, len(r.Relation.Members))
	}

	for i, child := range list {
		if child == nil {
			continue
		}

		switch t := child.(type) {
		case *childNode:
			r.Relation.Members[i].Version = t.Node.Version
			r.Relation.Members[i].ChangesetID = t.Node.ChangesetID
			r.Relation.Members[i].Lat = t.Node.Lat
			r.Relation.Members[i].Lon = t.Node.Lon
		case *childWay:
			r.Relation.Members[i].Version = t.Way.Version
			r.Relation.Members[i].ChangesetID = t.Way.ChangesetID

			if ways != nil {
				ways[t.Way.ID] = t.Way
			}
		case *childRelation:
			r.Relation.Members[i].Version = t.Relation.Version
			r.Relation.Members[i].ChangesetID = t.Relation.ChangesetID
		default:
			panic(fmt.Sprintf("unsupported type %T", child))
		}
	}

	if r.Relation.Polygon() {
		orientation(r.Relation.Members, ways, r.Relation.CommittedAt())
	}
}
