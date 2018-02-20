package osmpbf

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf/internal/osmpbf"
)

type elementInfo struct {
	Version   int32
	UID       int32
	Timestamp time.Time
	Changeset int64
	User      string
	Visible   bool
}

// dataDecoder is a decoder for Blob with OSMData (PrimitiveBlock).
type dataDecoder struct {
	q []osm.Object
}

func (dec *dataDecoder) Decode(blob *osmpbf.Blob) ([]osm.Object, error) {
	dec.q = make([]osm.Object, 0, 8000) // typical PrimitiveBlock contains 8k OSM entities

	data, err := getData(blob)
	if err != nil {
		return nil, err
	}

	primitiveBlock := &osmpbf.PrimitiveBlock{}
	if err := proto.Unmarshal(data, primitiveBlock); err != nil {
		return nil, err
	}

	dec.parsePrimitiveBlock(primitiveBlock)
	return dec.q, nil
}

func (dec *dataDecoder) parsePrimitiveBlock(pb *osmpbf.PrimitiveBlock) {
	for _, pg := range pb.GetPrimitivegroup() {
		dec.parsePrimitiveGroup(pb, pg)
	}
}

func (dec *dataDecoder) parsePrimitiveGroup(pb *osmpbf.PrimitiveBlock, pg *osmpbf.PrimitiveGroup) {
	dec.parseNodes(pb, pg.GetNodes())
	dec.parseDenseNodes(pb, pg.GetDense())
	dec.parseWays(pb, pg.GetWays())
	dec.parseRelations(pb, pg.GetRelations())
}

func (dec *dataDecoder) parseNodes(pb *osmpbf.PrimitiveBlock, nodes []*osmpbf.Node) {
	if len(nodes) == 0 {
		return
	}

	panic("nodes are not supported, currently untested")
	// st := pb.GetStringtable().GetS()
	// granularity := int64(pb.GetGranularity())
	// dateGranularity := int64(pb.GetDateGranularity())

	// latOffset := pb.GetLatOffset()
	// lonOffset := pb.GetLonOffset()

	// for _, node := range nodes {
	// 	info := extractInfo(st, node.GetInfo(), dateGranularity)
	// 	dec.q = append(dec.q, osm.Element{
	// 		Node: &osm.Node{
	// 			ID:          osm.NodeID(node.GetId()),
	// 			Lat:         1e-9 * float64((latOffset + (granularity * node.GetLat()))),
	// 			Lon:         1e-9 * float64((lonOffset + (granularity * node.GetLon()))),
	// 			User:        info.User,
	// 			UserID:      osm.UserID(info.UID),
	// 			Visible:     info.Visible,
	// 			ChangesetID: osm.ChangesetID(info.Changeset),
	// 			Timestamp:   info.Timestamp,
	// 			Tags:        extractOSMTags(st, node.GetKeys(), node.GetVals()),
	// 		},
	// 	})
	// }
}

func (dec *dataDecoder) parseDenseNodes(pb *osmpbf.PrimitiveBlock, dn *osmpbf.DenseNodes) {
	st := pb.GetStringtable().GetS()
	granularity := int64(pb.GetGranularity())

	latOffset := pb.GetLatOffset()
	lonOffset := pb.GetLonOffset()
	ids := dn.GetId()
	lats := dn.GetLat()
	lons := dn.GetLon()
	di := dn.GetDenseinfo()

	tu := tagUnpacker{st, dn.GetKeysVals(), 0}
	state := &denseInfoState{
		DenseInfo:       di,
		StringTable:     st,
		DateGranularity: int64(pb.GetDateGranularity()),
	}

	var id, lat, lon int64
	for index := range ids {
		id = ids[index] + id
		lat = lats[index] + lat
		lon = lons[index] + lon
		info := state.Next()

		dec.q = append(dec.q, &osm.Node{
			ID:          osm.NodeID(id),
			Lat:         1e-9 * float64((latOffset + (granularity * lat))),
			Lon:         1e-9 * float64((lonOffset + (granularity * lon))),
			User:        info.User,
			UserID:      osm.UserID(info.UID),
			Visible:     info.Visible,
			Version:     int(info.Version),
			ChangesetID: osm.ChangesetID(info.Changeset),
			Timestamp:   info.Timestamp,
			Tags:        tu.Next(),
		})
	}
}

func (dec *dataDecoder) parseWays(pb *osmpbf.PrimitiveBlock, ways []*osmpbf.Way) {
	st := pb.GetStringtable().GetS()
	dateGranularity := int64(pb.GetDateGranularity())

	for _, way := range ways {
		var (
			prev    int64
			nodeIDs osm.WayNodes
		)

		info := extractInfo(st, way.Info, dateGranularity)
		if refs := way.GetRefs(); len(refs) > 0 {
			nodeIDs = make(osm.WayNodes, len(refs))
			for i, r := range refs {
				prev = r + prev // delta encoding
				nodeIDs[i] = osm.WayNode{ID: osm.NodeID(prev)}
			}
		}

		dec.q = append(dec.q, &osm.Way{
			ID:          osm.WayID(way.Id),
			User:        info.User,
			UserID:      osm.UserID(info.UID),
			Visible:     info.Visible,
			Version:     int(info.Version),
			ChangesetID: osm.ChangesetID(info.Changeset),
			Timestamp:   info.Timestamp,
			Nodes:       nodeIDs,
			Tags:        extractTags(st, way.Keys, way.Vals),
		})
	}
}

// Make relation members from stringtable and three parallel arrays of IDs.
func extractMembers(stringTable []string, rel *osmpbf.Relation) osm.Members {
	memIDs := rel.GetMemids()
	types := rel.GetTypes()
	roleIDs := rel.GetRolesSid()

	var memID int64
	if len(memIDs) == 0 {
		return nil
	}

	members := make(osm.Members, len(memIDs))
	for index := range memIDs {
		memID = memIDs[index] + memID // delta encoding

		var memType osm.Type
		switch types[index] {
		case osmpbf.Relation_NODE:
			memType = osm.TypeNode
		case osmpbf.Relation_WAY:
			memType = osm.TypeWay
		case osmpbf.Relation_RELATION:
			memType = osm.TypeRelation
		}

		members[index] = osm.Member{
			Type: memType,
			Ref:  memID,
			Role: stringTable[roleIDs[index]],
		}
	}

	return members
}

func (dec *dataDecoder) parseRelations(pb *osmpbf.PrimitiveBlock, relations []*osmpbf.Relation) {
	st := pb.GetStringtable().GetS()
	dateGranularity := int64(pb.GetDateGranularity())

	for _, rel := range relations {
		members := extractMembers(st, rel)
		info := extractInfo(st, rel.GetInfo(), dateGranularity)

		dec.q = append(dec.q, &osm.Relation{
			ID:          osm.RelationID(rel.Id),
			User:        info.User,
			UserID:      osm.UserID(info.UID),
			Visible:     info.Visible,
			Version:     int(info.Version),
			ChangesetID: osm.ChangesetID(info.Changeset),
			Timestamp:   info.Timestamp,
			Tags:        extractTags(st, rel.GetKeys(), rel.GetVals()),
			Members:     members,
		})
	}
}

func extractInfo(stringTable []string, i *osmpbf.Info, dateGranularity int64) elementInfo {
	info := elementInfo{Visible: true}

	if i != nil {
		info.Version = i.GetVersion()

		millisec := time.Duration(i.GetTimestamp()*dateGranularity) * time.Millisecond
		info.Timestamp = time.Unix(0, millisec.Nanoseconds()).UTC()

		info.Changeset = i.GetChangeset()
		info.UID = i.GetUid()
		info.User = stringTable[i.GetUserSid()]

		if i.Visible != nil {
			info.Visible = i.GetVisible()
		}
	}

	return info
}

type denseInfoState struct {
	DenseInfo       *osmpbf.DenseInfo
	StringTable     []string
	DateGranularity int64

	index     int
	timestamp int64
	changeset int64
	uid       int32
	userSid   int32
}

func (s *denseInfoState) Next() elementInfo {
	info := elementInfo{Visible: true}

	if versions := s.DenseInfo.GetVersion(); len(versions) > 0 {
		info.Version = versions[s.index]
	}

	if timestamps := s.DenseInfo.GetTimestamp(); len(timestamps) > 0 {
		s.timestamp = timestamps[s.index] + s.timestamp
		millisec := time.Duration(s.timestamp*s.DateGranularity) * time.Millisecond
		info.Timestamp = time.Unix(0, millisec.Nanoseconds()).UTC()
	}

	if changesets := s.DenseInfo.GetChangeset(); len(changesets) > 0 {
		s.changeset = changesets[s.index] + s.changeset
		info.Changeset = s.changeset
	}

	if uids := s.DenseInfo.GetUid(); len(uids) > 0 {
		s.uid = uids[s.index] + s.uid
		info.UID = s.uid
	}

	if userSids := s.DenseInfo.GetUserSid(); len(userSids) > 0 {
		s.userSid = userSids[s.index] + s.userSid
		info.User = s.StringTable[s.userSid]
	}

	if visibles := s.DenseInfo.GetVisible(); len(visibles) > 0 {
		info.Visible = visibles[s.index]
	}

	s.index++
	return info
}
