package osm

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/paulmach/go.osm/osmpb"
)

const locMultiple = 10000000.0

var memberTypeMap = map[MemberType]osmpb.Relation_MemberType{
	NodeMember:     osmpb.Relation_NODE,
	WayMember:      osmpb.Relation_WAY,
	RelationMember: osmpb.Relation_RELATION,
}

var memberTypeMapRev = map[osmpb.Relation_MemberType]MemberType{
	osmpb.Relation_NODE:     NodeMember,
	osmpb.Relation_WAY:      WayMember,
	osmpb.Relation_RELATION: RelationMember,
}

func marshalNode(node *Node, ss *stringSet, includeChangeset bool) *osmpb.Node {
	keys, vals := node.Tags.KeyValues(ss)
	encoded := &osmpb.Node{
		Id:   proto.Int64(int64(node.ID)),
		Keys: keys,
		Vals: vals,
		Info: &osmpb.Info{
			Version:   proto.Int32(int32(node.Version)),
			Timestamp: proto.Int64(node.Timestamp.Unix()),
			Visible:   proto.Bool(node.Visible),
		},
		// geoToInt64
		Lat: proto.Int64(geoToInt64(node.Lat)),
		Lng: proto.Int64(geoToInt64(node.Lng)),
	}

	if includeChangeset {
		encoded.Info.Changeset = proto.Int64(int64(node.ChangesetID))
		encoded.Info.Uid = proto.Int32(int32(node.UserID))
		encoded.Info.UserSid = proto.Uint32(ss.Add(node.User))
	}

	return encoded
}

func unmarshalNode(encoded *osmpb.Node, ss []string) (*Node, error) {
	tags, err := tagsFromStrings(ss, encoded.GetKeys(), encoded.GetVals())
	if err != nil {
		return nil, err
	}

	info := encoded.GetInfo()
	return &Node{
		ID:          NodeID(encoded.GetId()),
		User:        ss[info.GetUserSid()],
		UserID:      UserID(info.GetUid()),
		Visible:     info.GetVisible(),
		Version:     int(info.GetVersion()),
		ChangesetID: ChangesetID(info.GetChangeset()),
		Timestamp:   time.Unix(info.GetTimestamp(), 0).UTC(),
		Tags:        tags,
		Lat:         float64(encoded.GetLat()) / locMultiple,
		Lng:         float64(encoded.GetLng()) / locMultiple,
	}, nil
}

func marshalNodes(nodes Nodes, ss *stringSet, includeChangeset bool) *osmpb.DenseNodes {
	dense := denseNodesValues(nodes)
	encoded := &osmpb.DenseNodes{
		Id: encodeInt64(dense.IDs),
		DenseInfo: &osmpb.DenseInfo{
			Version:   dense.Versions,
			Timestamp: encodeInt64(dense.Timestamps),
			Visible:   dense.Visibles,
		},
		Lat: encodeInt64(dense.Lats),
		Lng: encodeInt64(dense.Lngs),
	}

	if dense.TagCount > 0 {
		encoded.KeysVals = encodeNodesTags(nodes, ss, dense.TagCount)
	}

	if includeChangeset {
		csinfo := nodesChangesetInfo(nodes, ss)
		encoded.DenseInfo.Changeset = encodeInt64(csinfo.Changesets)
		encoded.DenseInfo.Uid = encodeInt32(csinfo.Uids)
		encoded.DenseInfo.UserSid = encodeInt32(csinfo.UserSid)
	}

	return encoded
}

func unmarshalNodes(encoded *osmpb.DenseNodes, ss []string) (Nodes, error) {
	encoded.Id = decodeInt64(encoded.Id)
	encoded.Lat = decodeInt64(encoded.Lat)
	encoded.Lng = decodeInt64(encoded.Lng)
	encoded.DenseInfo.Timestamp = decodeInt64(encoded.DenseInfo.Timestamp)
	encoded.DenseInfo.Changeset = decodeInt64(encoded.DenseInfo.Changeset)
	encoded.DenseInfo.Uid = decodeInt32(encoded.DenseInfo.Uid)
	encoded.DenseInfo.UserSid = decodeInt32(encoded.DenseInfo.UserSid)

	tagLoc := 0
	nodes := make(Nodes, 0, len(encoded.Id))
	for i := range encoded.Id {
		n := &Node{
			ID:        NodeID(encoded.Id[i]),
			Lat:       float64(encoded.Lat[i]) / locMultiple,
			Lng:       float64(encoded.Lng[i]) / locMultiple,
			Visible:   encoded.DenseInfo.Visible[i],
			Version:   int(encoded.DenseInfo.Version[i]),
			Timestamp: time.Unix(encoded.DenseInfo.Timestamp[i], 0).UTC(),
		}

		if len(encoded.DenseInfo.Changeset) > 0 {
			n.ChangesetID = ChangesetID(encoded.DenseInfo.Changeset[i])
		}

		if len(encoded.DenseInfo.Uid) > 0 {
			n.UserID = UserID(encoded.DenseInfo.Uid[i])
		}

		if len(encoded.DenseInfo.UserSid) > 0 {
			n.User = ss[encoded.DenseInfo.UserSid[i]]
		}

		if encoded.KeysVals[tagLoc] == 0 {
			tagLoc++
		} else {
			for encoded.KeysVals[tagLoc] != 0 {
				// TODO: bound check these key-values
				n.Tags = append(n.Tags, Tag{
					Key:   ss[encoded.KeysVals[tagLoc]],
					Value: ss[encoded.KeysVals[tagLoc+1]],
				})

				tagLoc += 2
			}
			tagLoc++
		}
		nodes = append(nodes, n)
	}

	return nodes, nil
}

func marshalWay(way *Way, ss *stringSet, includeChangeset bool) *osmpb.Way {
	keys, vals := way.Tags.KeyValues(ss)
	encoded := &osmpb.Way{
		Id:   proto.Int64(int64(way.ID)),
		Keys: keys,
		Vals: vals,
		Info: &osmpb.Info{
			Version:   proto.Int32(int32(way.Version)),
			Timestamp: proto.Int64(way.Timestamp.Unix()),
			Visible:   proto.Bool(way.Visible),
		},
		Refs: encodeNodeRef(way.NodeRefs),
	}

	if includeChangeset {
		encoded.Info.Changeset = proto.Int64(int64(way.ChangesetID))
		encoded.Info.Uid = proto.Int32(int32(way.UserID))
		encoded.Info.UserSid = proto.Uint32(ss.Add(way.User))
	}

	return encoded
}

func unmarshalWay(encoded *osmpb.Way, ss []string) (*Way, error) {
	tags, err := tagsFromStrings(ss, encoded.GetKeys(), encoded.GetVals())
	if err != nil {
		return nil, err
	}

	info := encoded.GetInfo()
	return &Way{
		ID:          WayID(encoded.GetId()),
		User:        ss[info.GetUserSid()],
		UserID:      UserID(info.GetUid()),
		Visible:     info.GetVisible(),
		Version:     int(info.GetVersion()),
		ChangesetID: ChangesetID(info.GetChangeset()),
		Timestamp:   time.Unix(info.GetTimestamp(), 0).UTC(),
		NodeRefs:    decodeNodeRef(encoded.GetRefs()),
		Tags:        tags,
	}, nil
}

func marshalRelation(relation *Relation, ss *stringSet, includeChangeset bool) *osmpb.Relation {
	l := len(relation.Members)
	roles := make([]uint32, l, l)
	refs := make([]int64, l, l)
	types := make([]osmpb.Relation_MemberType, l, l)

	for i, m := range relation.Members {
		roles[i] = ss.Add(m.Role)
		refs[i] = int64(m.Ref)
		types[i] = memberTypeMap[m.Type]
	}

	keys, vals := relation.Tags.KeyValues(ss)
	encoded := &osmpb.Relation{
		Id:   proto.Int64(int64(relation.ID)),
		Keys: keys,
		Vals: vals,
		Info: &osmpb.Info{
			Version:   proto.Int32(int32(relation.Version)),
			Timestamp: proto.Int64(relation.Timestamp.Unix()),
			Visible:   proto.Bool(relation.Visible),
		},
		Roles: roles,
		Refs:  encodeInt64(refs),
		Types: types,
	}

	if includeChangeset {
		encoded.Info.Changeset = proto.Int64(int64(relation.ChangesetID))
		encoded.Info.Uid = proto.Int32(int32(relation.UserID))
		encoded.Info.UserSid = proto.Uint32(ss.Add(relation.User))
	}

	return encoded
}

func unmarshalRelation(encoded *osmpb.Relation, ss []string) (*Relation, error) {
	tags, err := tagsFromStrings(ss, encoded.GetKeys(), encoded.GetVals())
	if err != nil {
		return nil, err
	}

	info := encoded.GetInfo()
	return &Relation{
		ID:          RelationID(encoded.GetId()),
		User:        ss[info.GetUserSid()],
		UserID:      UserID(info.GetUid()),
		Visible:     info.GetVisible(),
		Version:     int(info.GetVersion()),
		ChangesetID: ChangesetID(info.GetChangeset()),
		Timestamp:   time.Unix(info.GetTimestamp(), 0).UTC(),
		Members:     decodeMembers(ss, encoded.GetRoles(), encoded.GetRefs(), encoded.GetTypes()),
		Tags:        tags,
	}, nil
}

func geoToInt64(l float64) int64 {
	// on rounding errors
	//
	// It is the case that 32.850314 * 10e6 = 32850313.999999996
	// Simpily casting this as an int will truncate towards zero
	// and result in an off by one. The true solution is to round
	// the scaled result, like so:
	//
	// int64(math.Floor(stream.BaseData[i][0]*factor + 0.5))
	//
	// However, the code below does the same thing in this context,
	// and is twice as fast:
	sign := 0.5
	if l < 0 {
		sign = -0.5
	}

	return int64(l*locMultiple + sign)
}

func encodeNodeRef(refs []NodeRef) []int64 {
	result := make([]int64, 0, len(refs))
	var prev int64

	for _, r := range refs {
		result = append(result, int64(r.Ref)-prev)
		prev = int64(r.Ref)
	}

	return result
}

func decodeMembers(ss []string, roles []uint32, refs []int64, types []osmpb.Relation_MemberType) []Member {
	result := make([]Member, len(roles), len(roles))

	decodeInt64(refs)
	for i := range roles {
		result[i].Role = ss[roles[i]]
		result[i].Ref = int(refs[i])
		result[i].Type = memberTypeMapRev[types[i]]
	}

	return result
}

func encodeInt32(vals []int32) []int32 {
	var prev int32
	for i, v := range vals {
		vals[i] = v - prev
		prev = v
	}

	return vals
}

func encodeInt64(vals []int64) []int64 {
	var prev int64
	for i, v := range vals {
		vals[i] = v - prev
		prev = v
	}

	return vals
}

func decodeNodeRef(diff []int64) []NodeRef {
	result := make([]NodeRef, 0, len(diff))
	var prev NodeID

	for _, d := range diff {
		result = append(result, NodeRef{Ref: NodeID(d) + prev})
		prev += NodeID(d)
	}

	return result
}

func decodeInt32(vals []int32) []int32 {
	var prev int32
	for i, v := range vals {
		prev += v
		vals[i] = prev
	}

	return vals
}

func decodeInt64(vals []int64) []int64 {
	var prev int64
	for i, v := range vals {
		prev += v
		vals[i] = prev
	}

	return vals
}

type denseNodesResult struct {
	IDs        []int64
	Lats       []int64
	Lngs       []int64
	Timestamps []int64
	Versions   []int32
	Visibles   []bool
	TagCount   int
}

func denseNodesValues(ns Nodes) denseNodesResult {
	l := len(ns)
	ds := denseNodesResult{
		IDs:        make([]int64, l, l),
		Lats:       make([]int64, l, l),
		Lngs:       make([]int64, l, l),
		Timestamps: make([]int64, l, l),
		Versions:   make([]int32, l, l),
		Visibles:   make([]bool, l, l),
	}

	for i, n := range ns {
		ds.IDs[i] = int64(n.ID)
		ds.Lats[i] = geoToInt64(n.Lat)
		ds.Lngs[i] = geoToInt64(n.Lng)
		ds.Timestamps[i] = n.Timestamp.Unix()
		ds.Versions[i] = int32(n.Version)
		ds.Visibles[i] = n.Visible
		ds.TagCount += len(n.Tags)
	}

	return ds
}

func encodeNodesTags(ns Nodes, ss *stringSet, count int) []uint32 {
	r := make([]uint32, 0, 2*count+len(ns))
	for _, n := range ns {
		for _, t := range n.Tags {
			r = append(r, ss.Add(t.Key))
			r = append(r, ss.Add(t.Value))
		}
		r = append(r, 0)
	}

	return r
}

type changesetInfoResult struct {
	Changesets []int64
	Uids       []int32
	UserSid    []int32
}

func nodesChangesetInfo(ns Nodes, ss *stringSet) changesetInfoResult {
	l := len(ns)
	cs := changesetInfoResult{
		Changesets: make([]int64, l, l),
		Uids:       make([]int32, l, l),
		UserSid:    make([]int32, l, l),
	}

	for i, n := range ns {
		cs.Changesets[i] = int64(n.ChangesetID)
		cs.Uids[i] = int32(n.UserID)
		cs.UserSid[i] = int32(ss.Add(n.User))
	}

	return cs
}
