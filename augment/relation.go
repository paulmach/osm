package augment

import (
	"context"
	"fmt"
	"time"

	osm "github.com/paulmach/go.osm"
	"github.com/paulmach/go.osm/augment/internal/core"
)

// Relations computes the updates for the given relations
// and augments members with stuff like changeset and lon/lat data.
// The input relations are modified to include this information.
func Relations(
	ctx context.Context,
	relations osm.Relations,
	datasource Datasource,
	threshold time.Duration,
) error {
	parents, children, err := convertRelationData(ctx, relations, datasource)
	if err != nil {
		return mapErrors(err)
	}

	updatesForParents, err := core.Compute(parents, children, threshold)
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
	datasource Datasource,
) ([]core.Parent, *core.Histories, error) {
	relations.SortByIDVersion()

	parents := make([]core.Parent, len(relations))
	histories := &core.Histories{}

	for i, r := range relations {
		parents[i] = &parentRelation{Relation: r}

		for j, m := range r.Members {
			childID := m.ElementID()
			if histories.Get(childID) != nil {
				continue
			}

			var list core.ChildList
			switch childID.Type {
			case osm.NodeType:
				nodes, err := datasource.NodeHistory(ctx, childID.NodeID())
				if err != nil {
					return nil, nil, err
				}

				list = nodesToChildList(nodes)
				histories.Set(childID, list)
			case osm.WayType:
				ways, err := datasource.WayHistory(ctx, childID.WayID())
				if err != nil {
					return nil, nil, err
				}

				list = waysToChildList(ways)
				histories.Set(childID, list)
			case osm.RelationType:
				relations, err := datasource.RelationHistory(ctx, childID.RelationID())
				if err != nil {
					return nil, nil, err
				}

				list = relationsToChildList(relations)
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
	list := make(core.ChildList, len(ways))

	ways.SortByIDVersion()
	for i, w := range ways {
		list[i] = &childWay{
			Index: i,
			Way:   w,
		}
	}

	return list
}

func relationsToChildList(relations osm.Relations) core.ChildList {
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

func (r parentRelation) ID() osm.ElementID {
	return r.Relation.ElementID()
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

func (r parentRelation) Refs() osm.ElementIDs {
	return r.Relation.Members.ElementIDs()
}

func (r parentRelation) Children() core.ChildList {
	return r.children
}

func (r *parentRelation) SetChildren(list core.ChildList) {
	r.children = list

	for i, child := range list {
		switch t := child.(type) {
		case *childNode:
			r.Relation.Members[i].Version = t.Node.Version
			r.Relation.Members[i].ChangesetID = t.Node.ChangesetID
			r.Relation.Members[i].Lat = t.Node.Lat
			r.Relation.Members[i].Lon = t.Node.Lon
		case *childWay:
			r.Relation.Members[i].Version = t.Way.Version
			r.Relation.Members[i].ChangesetID = t.Way.ChangesetID
		case *childRelation:
			r.Relation.Members[i].Version = t.Relation.Version
			r.Relation.Members[i].ChangesetID = t.Relation.ChangesetID
		default:
			panic(fmt.Sprintf("unsupported type %T", child))
		}
	}
}
