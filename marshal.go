package osm

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/paulmach/go.osm/internal/osmpb"
)

const locMultiple = 10000000.0

var memberTypeMap = map[ElementType]osmpb.MemberType{
	NodeType:     osmpb.MemberType_NODE,
	WayType:      osmpb.MemberType_WAY,
	RelationType: osmpb.MemberType_RELATION,
}

var memberTypeMapRev = map[osmpb.MemberType]ElementType{
	osmpb.MemberType_NODE:     NodeType,
	osmpb.MemberType_WAY:      WayType,
	osmpb.MemberType_RELATION: RelationType,
}

func marshalNode(node *Node, ss *stringSet, includeChangeset bool) *osmpb.Node {
	keys, vals := node.Tags.keyValues(ss)
	encoded := &osmpb.Node{
		Id:   int64(node.ID),
		Keys: keys,
		Vals: vals,
		Info: &osmpb.Info{
			Version:   int32(node.Version),
			Timestamp: timeToUnix(node.Timestamp),
			Visible:   proto.Bool(node.Visible),
		},
		// geoToInt64
		Lat: geoToInt64(node.Lat),
		Lon: geoToInt64(node.Lon),
	}

	if includeChangeset {
		encoded.Info.ChangesetId = int64(node.ChangesetID)
		encoded.Info.UserId = int32(node.UserID)
		encoded.Info.UserSid = ss.Add(node.User)
	}

	return encoded
}

func unmarshalNode(encoded *osmpb.Node, ss []string, cs *Changeset) (*Node, error) {
	tags, err := tagsFromStrings(ss, encoded.GetKeys(), encoded.GetVals())
	if err != nil {
		return nil, err
	}

	info := encoded.GetInfo()
	n := &Node{
		ID:          NodeID(encoded.GetId()),
		User:        ss[info.GetUserSid()],
		UserID:      UserID(info.GetUserId()),
		Visible:     info.GetVisible(),
		Version:     int(info.GetVersion()),
		ChangesetID: ChangesetID(info.GetChangesetId()),
		Timestamp:   unixToTime(info.GetTimestamp()),
		Tags:        tags,
		Lat:         float64(encoded.GetLat()) / locMultiple,
		Lon:         float64(encoded.GetLon()) / locMultiple,
	}

	if cs != nil {
		n.ChangesetID = cs.ID
		n.UserID = cs.UserID
		n.User = cs.User
	}

	return n, nil
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
		Lon: encodeInt64(dense.Lons),
	}

	if dense.TagCount > 0 {
		encoded.KeysVals = encodeNodesTags(nodes, ss, dense.TagCount)
	}

	if includeChangeset {
		csinfo := nodesChangesetInfo(nodes, ss)
		encoded.DenseInfo.ChangesetId = encodeInt64(csinfo.Changesets)
		encoded.DenseInfo.UserId = encodeInt32(csinfo.UserIDs)
		encoded.DenseInfo.UserSid = encodeInt32(csinfo.UserSid)
	}

	return encoded
}

func unmarshalNodes(encoded *osmpb.DenseNodes, ss []string, cs *Changeset) (Nodes, error) {
	encoded.Id = decodeInt64(encoded.Id)
	encoded.Lat = decodeInt64(encoded.Lat)
	encoded.Lon = decodeInt64(encoded.Lon)
	encoded.DenseInfo.Timestamp = decodeInt64(encoded.DenseInfo.Timestamp)
	encoded.DenseInfo.ChangesetId = decodeInt64(encoded.DenseInfo.ChangesetId)
	encoded.DenseInfo.UserId = decodeInt32(encoded.DenseInfo.UserId)
	encoded.DenseInfo.UserSid = decodeInt32(encoded.DenseInfo.UserSid)

	tagLoc := 0
	nodes := make(Nodes, len(encoded.Id))
	for i := range encoded.Id {
		n := &Node{
			ID:        NodeID(encoded.Id[i]),
			Lat:       float64(encoded.Lat[i]) / locMultiple,
			Lon:       float64(encoded.Lon[i]) / locMultiple,
			Visible:   encoded.DenseInfo.Visible[i],
			Version:   int(encoded.DenseInfo.Version[i]),
			Timestamp: unixToTime(encoded.DenseInfo.Timestamp[i]),
		}

		if cs != nil {
			n.ChangesetID = cs.ID
			n.UserID = cs.UserID
			n.User = cs.User
		} else {
			if len(encoded.DenseInfo.ChangesetId) > 0 {
				n.ChangesetID = ChangesetID(encoded.DenseInfo.ChangesetId[i])
			}

			if len(encoded.DenseInfo.UserId) > 0 {
				n.UserID = UserID(encoded.DenseInfo.UserId[i])
			}

			if len(encoded.DenseInfo.UserSid) > 0 {
				n.User = ss[encoded.DenseInfo.UserSid[i]]
			}
		}

		if encoded.KeysVals != nil {
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
		}

		nodes[i] = n
	}

	return nodes, nil
}

func marshalWay(way *Way, ss *stringSet, includeChangeset bool) *osmpb.Way {
	keys, vals := way.Tags.keyValues(ss)
	encoded := &osmpb.Way{
		Id:   int64(way.ID),
		Keys: keys,
		Vals: vals,
		Info: &osmpb.Info{
			Version:   int32(way.Version),
			Timestamp: timeToUnix(way.Timestamp),
			Visible:   proto.Bool(way.Visible),
		},
	}

	if len(way.Nodes) > 0 {
		// legacy simple refs encoding.
		if way.Nodes[0].Version == 0 {
			encoded.Refs = encodeWayNodes(way.Nodes)
		} else {
			encoded.DenseMembers = encodeDenseWayNodes(way.Nodes)
		}
	}

	if len(way.Minors) > 0 {
		encoded.MinorVersion = encodeMinorWays(way.Minors)
	}

	if includeChangeset {
		encoded.Info.ChangesetId = int64(way.ChangesetID)
		encoded.Info.UserId = int32(way.UserID)
		encoded.Info.UserSid = ss.Add(way.User)
	}

	return encoded
}

func unmarshalWay(encoded *osmpb.Way, ss []string, cs *Changeset) (*Way, error) {
	tags, err := tagsFromStrings(ss, encoded.GetKeys(), encoded.GetVals())
	if err != nil {
		return nil, err
	}

	info := encoded.GetInfo()
	w := &Way{
		ID:          WayID(encoded.GetId()),
		User:        ss[info.GetUserSid()],
		UserID:      UserID(info.GetUserId()),
		Visible:     info.GetVisible(),
		Version:     int(info.GetVersion()),
		ChangesetID: ChangesetID(info.GetChangesetId()),
		Timestamp:   unixToTime(info.GetTimestamp()),
		Tags:        tags,
	}

	if len(encoded.Refs) > 0 {
		w.Nodes = decodeWayNodes(encoded.GetRefs())
	} else if encoded.DenseMembers != nil {
		w.Nodes = decodeDenseWayNodes(encoded.GetDenseMembers())
	}

	if len(encoded.MinorVersion) > 0 {
		w.Minors = decodeMinorWays(encoded.MinorVersion)
	}

	if cs != nil {
		w.ChangesetID = cs.ID
		w.UserID = cs.UserID
		w.User = cs.User
	}

	return w, nil
}

func marshalRelation(relation *Relation, ss *stringSet, includeChangeset bool) *osmpb.Relation {
	l := len(relation.Members)
	roles := make([]uint32, l)
	refs := make([]int64, l)
	types := make([]osmpb.MemberType, l)

	for i, m := range relation.Members {
		roles[i] = ss.Add(m.Role)
		refs[i] = int64(m.Ref)
		types[i] = memberTypeMap[m.Type]
	}

	keys, vals := relation.Tags.keyValues(ss)
	encoded := &osmpb.Relation{
		Id:   int64(relation.ID),
		Keys: keys,
		Vals: vals,
		Info: &osmpb.Info{
			Version:   int32(relation.Version),
			Timestamp: timeToUnix(relation.Timestamp),
			Visible:   proto.Bool(relation.Visible),
		},
		Roles: roles,
		Refs:  encodeInt64(refs),
		Types: types,
	}

	if len(relation.Members) > 0 && relation.Members[0].Version != 0 {
		encoded.DenseMembers = encodeDenseMembers(relation.Members)
	}

	if len(relation.Minors) > 0 {
		encoded.MinorVersion = encodeMinorRelations(relation.Minors)
	}

	if includeChangeset {
		encoded.Info.ChangesetId = int64(relation.ChangesetID)
		encoded.Info.UserId = int32(relation.UserID)
		encoded.Info.UserSid = ss.Add(relation.User)
	}

	return encoded
}

func unmarshalRelation(encoded *osmpb.Relation, ss []string, cs *Changeset) (*Relation, error) {
	tags, err := tagsFromStrings(ss, encoded.GetKeys(), encoded.GetVals())
	if err != nil {
		return nil, err
	}

	info := encoded.GetInfo()
	r := &Relation{
		ID:          RelationID(encoded.GetId()),
		User:        ss[info.GetUserSid()],
		UserID:      UserID(info.GetUserId()),
		Visible:     info.GetVisible(),
		Version:     int(info.GetVersion()),
		ChangesetID: ChangesetID(info.GetChangesetId()),
		Timestamp:   unixToTime(info.GetTimestamp()),
		Members:     decodeMembers(ss, encoded.GetRoles(), encoded.GetRefs(), encoded.GetTypes()),
		Tags:        tags,
	}

	decodeDenseMembers(r.Members, encoded.GetDenseMembers())

	if len(encoded.MinorVersion) > 0 {
		r.Minors = decodeMinorRelations(encoded.MinorVersion)
	}

	if cs != nil {
		r.ChangesetID = cs.ID
		r.UserID = cs.UserID
		r.User = cs.User
	}

	return r, nil
}

type denseNodesResult struct {
	IDs        []int64
	Lats       []int64
	Lons       []int64
	Timestamps []int64
	Versions   []int32
	Visibles   []bool
	TagCount   int
}

func denseNodesValues(ns Nodes) denseNodesResult {
	l := len(ns)
	ds := denseNodesResult{
		IDs:        make([]int64, l),
		Lats:       make([]int64, l),
		Lons:       make([]int64, l),
		Timestamps: make([]int64, l),
		Versions:   make([]int32, l),
		Visibles:   make([]bool, l),
	}

	for i, n := range ns {
		ds.IDs[i] = int64(n.ID)
		ds.Lats[i] = geoToInt64(n.Lat)
		ds.Lons[i] = geoToInt64(n.Lon)
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
	UserIDs    []int32
	UserSid    []int32
}

func nodesChangesetInfo(ns Nodes, ss *stringSet) changesetInfoResult {
	l := len(ns)
	cs := changesetInfoResult{
		Changesets: make([]int64, l),
		UserIDs:    make([]int32, l),
		UserSid:    make([]int32, l),
	}

	for i, n := range ns {
		cs.Changesets[i] = int64(n.ChangesetID)
		cs.UserIDs[i] = int32(n.UserID)
		cs.UserSid[i] = int32(ss.Add(n.User))
	}

	return cs
}

func encodeWayNodes(waynodes []WayNode) []int64 {
	result := make([]int64, len(waynodes))
	var prev int64

	for i, r := range waynodes {
		result[i] = int64(r.ID) - prev
		prev = int64(r.ID)
	}

	return result
}

func encodeDenseWayNodes(waynodes []WayNode) *osmpb.DenseMembers {
	l := len(waynodes)

	IDs := make([]int64, l)
	Versions := make([]int32, l)
	ChangesetIDs := make([]int64, l)
	Lats := make([]int64, l)
	Lons := make([]int64, l)

	for i, n := range waynodes {
		IDs[i] = int64(n.ID)
		Lats[i] = geoToInt64(n.Lat)
		Lons[i] = geoToInt64(n.Lon)
		Versions[i] = int32(n.Version)
		ChangesetIDs[i] = int64(n.ChangesetID)
	}

	return &osmpb.DenseMembers{
		Id:          encodeInt64(IDs),
		Version:     Versions,
		ChangesetId: encodeInt64(ChangesetIDs),
		Lat:         encodeInt64(Lats),
		Lon:         encodeInt64(Lons),
	}
}

func encodeDenseMinorWayNodes(waynodes []MinorWayNode) *osmpb.DenseMembers {
	if len(waynodes) == 0 {
		return nil
	}

	l := len(waynodes)
	Indexes := make([]int32, l)
	Versions := make([]int32, l)
	ChangesetIDs := make([]int64, l)
	Lats := make([]int64, l)
	Lons := make([]int64, l)

	for i, n := range waynodes {
		Indexes[i] = int32(n.Index)
		Lats[i] = geoToInt64(n.Lat)
		Lons[i] = geoToInt64(n.Lon)
		Versions[i] = int32(n.Version)
		ChangesetIDs[i] = int64(n.ChangesetID)
	}

	return &osmpb.DenseMembers{
		Index:       encodeInt32(Indexes),
		Version:     Versions,
		ChangesetId: encodeInt64(ChangesetIDs),
		Lat:         encodeInt64(Lats),
		Lon:         encodeInt64(Lons),
	}
}

func encodeMinorWays(ways []MinorWay) []*osmpb.MinorVersion {
	if len(ways) == 0 {
		return nil
	}

	result := make([]*osmpb.MinorVersion, len(ways))
	for i, w := range ways {
		result[i] = &osmpb.MinorVersion{
			Timestamp:    timeToUnix(w.Timestamp),
			DenseMembers: encodeDenseMinorWayNodes(w.MinorNodes),
		}
	}

	return result
}

func encodeDenseMembers(members []Member) *osmpb.DenseMembers {
	l := len(members)
	versions := make([]int32, l)
	minorVersions := make([]int32, l)
	changesetIDs := make([]int64, l)

	for i, m := range members {
		versions[i] = int32(m.Version)
		minorVersions[i] = int32(m.MinorVersion)
		changesetIDs[i] = int64(m.ChangesetID)
	}

	return &osmpb.DenseMembers{
		Version:      versions,
		MinorVersion: minorVersions,
		ChangesetId:  encodeInt64(changesetIDs),
	}
}

func encodeMinorRelations(relations []MinorRelation) []*osmpb.MinorVersion {
	if len(relations) == 0 {
		return nil
	}

	result := make([]*osmpb.MinorVersion, len(relations))
	for i, r := range relations {
		result[i] = &osmpb.MinorVersion{
			Timestamp:    timeToUnix(r.Timestamp),
			DenseMembers: encodeDenseMinorMembers(r.MinorMembers),
		}
	}

	return result
}

func encodeDenseMinorMembers(members []MinorRelationMember) *osmpb.DenseMembers {
	l := len(members)
	indexes := make([]int32, l)
	versions := make([]int32, l)
	minorVersions := make([]int32, l)
	changesetIDs := make([]int64, l)

	for i, m := range members {
		indexes[i] = int32(m.Index)
		versions[i] = int32(m.Version)
		minorVersions[i] = int32(m.MinorVersion)
		changesetIDs[i] = int64(m.ChangesetID)
	}

	return &osmpb.DenseMembers{
		Index:        encodeInt32(indexes),
		Version:      versions,
		MinorVersion: minorVersions,
		ChangesetId:  encodeInt64(changesetIDs),
	}
}

func decodeMembers(ss []string, roles []uint32, refs []int64, types []osmpb.MemberType) []Member {
	if len(roles) == 0 {
		return nil
	}

	result := make([]Member, len(roles))
	decodeInt64(refs)
	for i := range roles {
		result[i].Role = ss[roles[i]]
		result[i].Ref = int64(refs[i])
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

func decodeWayNodes(diff []int64) []WayNode {
	if len(diff) == 0 {
		return nil
	}

	result := make([]WayNode, len(diff))
	decodeInt64(diff)

	for i, d := range diff {
		result[i] = WayNode{ID: NodeID(d)}
	}

	return result
}

func decodeDenseWayNodes(encoded *osmpb.DenseMembers) []WayNode {
	if encoded == nil {
		return nil
	}

	result := make([]WayNode, len(encoded.Id))

	decodeInt64(encoded.Id)
	decodeInt64(encoded.ChangesetId)
	decodeInt64(encoded.Lat)
	decodeInt64(encoded.Lon)

	for i := range encoded.Id {
		result[i] = WayNode{
			ID:          NodeID(encoded.Id[i]),
			Version:     int(encoded.Version[i]),
			ChangesetID: ChangesetID(encoded.ChangesetId[i]),
			Lat:         float64(encoded.Lat[i]) / locMultiple,
			Lon:         float64(encoded.Lon[i]) / locMultiple,
		}
	}

	return result
}

func decodeDenseMinorWayNodes(encoded *osmpb.DenseMembers) []MinorWayNode {
	if encoded == nil || len(encoded.Index) == 0 {
		return nil
	}

	result := make([]MinorWayNode, len(encoded.Index))

	decodeInt32(encoded.Index)
	decodeInt64(encoded.ChangesetId)
	decodeInt64(encoded.Lat)
	decodeInt64(encoded.Lon)

	for i := range encoded.Index {
		result[i] = MinorWayNode{
			Index:       int(encoded.Index[i]),
			Version:     int(encoded.Version[i]),
			ChangesetID: ChangesetID(encoded.ChangesetId[i]),
			Lat:         float64(encoded.Lat[i]) / locMultiple,
			Lon:         float64(encoded.Lon[i]) / locMultiple,
		}
	}

	return result
}

func decodeMinorWays(encoded []*osmpb.MinorVersion) []MinorWay {
	if len(encoded) == 0 {
		return nil
	}

	result := make([]MinorWay, len(encoded))
	for i, mv := range encoded {
		result[i] = MinorWay{
			Timestamp:  unixToTime(mv.Timestamp),
			MinorNodes: decodeDenseMinorWayNodes(mv.DenseMembers),
		}
	}

	return result
}

func decodeDenseMembers(members []Member, encoded *osmpb.DenseMembers) {
	if encoded == nil || len(encoded.Version) == 0 {
		return
	}

	decodeInt64(encoded.ChangesetId)

	for i := range encoded.Version {
		members[i].Version = int(encoded.Version[i])
		members[i].MinorVersion = int(encoded.MinorVersion[i])
		members[i].ChangesetID = ChangesetID(encoded.ChangesetId[i])
	}

	return
}

func decodeMinorRelations(encoded []*osmpb.MinorVersion) []MinorRelation {
	if len(encoded) == 0 {
		return nil
	}

	result := make([]MinorRelation, len(encoded))
	for i, mv := range encoded {
		result[i] = MinorRelation{
			Timestamp:    unixToTime(mv.Timestamp),
			MinorMembers: decodeDenseMinorMembers(mv.DenseMembers),
		}
	}

	return result
}

func decodeDenseMinorMembers(encoded *osmpb.DenseMembers) []MinorRelationMember {
	if encoded == nil || len(encoded.Index) == 0 {
		return nil
	}

	result := make([]MinorRelationMember, len(encoded.Index))

	decodeInt32(encoded.Index)
	decodeInt64(encoded.ChangesetId)

	for i := range encoded.Index {
		result[i] = MinorRelationMember{
			Index:        int(encoded.Index[i]),
			Version:      int(encoded.Version[i]),
			MinorVersion: int(encoded.MinorVersion[i]),
			ChangesetID:  ChangesetID(encoded.ChangesetId[i]),
		}
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

func timeToUnix(t time.Time) int64 {
	u := t.Unix()
	if u <= 0 {
		return 0
	}

	return u
}

func timeToUnixPointer(t time.Time) *int64 {
	u := t.Unix()
	if u <= 0 {
		return nil
	}

	return &u
}

func unixToTime(u int64) time.Time {
	if u <= 0 {
		return time.Time{}
	}

	return time.Unix(u, 0).UTC()
}
