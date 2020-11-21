package osmpbf

import (
	"time"

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
	data []byte
	q    []osm.Object
}

func (dec *dataDecoder) Decode(blob *osmpbf.Blob) ([]osm.Object, error) {
	dec.q = make([]osm.Object, 0, 8000) // typical PrimitiveBlock contains 8k OSM entities

	var err error
	dec.data, err = getData(blob, dec.data)
	if err != nil {
		return nil, err
	}

	primitiveBlock := &osmpbf.PrimitiveBlock{}
	if err := primitiveBlock.Unmarshal(dec.data); err != nil {
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
	myNodes := make([]osm.Node, len(ids), len(ids))
	for index := range ids {
		id = ids[index] + id
		lat = lats[index] + lat
		lon = lons[index] + lon
		state.Next()

		myNodes[index].ID = osm.NodeID(id)
		myNodes[index].Lat = 1e-9 * float64((latOffset + (granularity * lat)))
		myNodes[index].Lon = 1e-9 * float64((lonOffset + (granularity * lon)))
		myNodes[index].User = state.info.User
		myNodes[index].UserID = osm.UserID(state.info.UID)
		myNodes[index].Visible = state.info.Visible
		myNodes[index].Version = int(state.info.Version)
		myNodes[index].ChangesetID = osm.ChangesetID(state.info.Changeset)
		myNodes[index].Timestamp = state.info.Timestamp
		myNodes[index].Tags = tu.Next()

		dec.q = append(dec.q, &myNodes[index])
	}
}

func (dec *dataDecoder) parseWays(pb *osmpbf.PrimitiveBlock, ways []*osmpbf.Way) {
	st := pb.GetStringtable().GetS()
	dateGranularity := int64(pb.GetDateGranularity())
	myWays := make([]osm.Way, len(ways))
	for i, way := range ways {
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

		myWays[i].ID = osm.WayID(way.Id)
		myWays[i].User = info.User
		myWays[i].UserID = osm.UserID(info.UID)
		myWays[i].Visible = info.Visible
		myWays[i].Version = int(info.Version)
		myWays[i].ChangesetID = osm.ChangesetID(info.Changeset)
		myWays[i].Timestamp = info.Timestamp
		myWays[i].Nodes = nodeIDs
		myWays[i].Tags = extractTags(st, way.Keys, way.Vals)

		dec.q = append(dec.q, &myWays[i])
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

	info elementInfo
}

func (s *denseInfoState) Next() {
	s.info = elementInfo{Visible: true}

	if versions := s.DenseInfo.GetVersion(); len(versions) > 0 {
		s.info.Version = versions[s.index]
	}

	if timestamps := s.DenseInfo.GetTimestamp(); len(timestamps) > 0 {
		s.timestamp = timestamps[s.index] + s.timestamp
		millisec := time.Duration(s.timestamp*s.DateGranularity) * time.Millisecond
		s.info.Timestamp = time.Unix(0, millisec.Nanoseconds()).UTC()
	}

	if changesets := s.DenseInfo.GetChangeset(); len(changesets) > 0 {
		s.changeset = changesets[s.index] + s.changeset
		s.info.Changeset = s.changeset
	}

	if uids := s.DenseInfo.GetUid(); len(uids) > 0 {
		s.uid = uids[s.index] + s.uid
		s.info.UID = s.uid
	}

	if userSids := s.DenseInfo.GetUserSid(); len(userSids) > 0 {
		s.userSid = userSids[s.index] + s.userSid
		s.info.User = s.StringTable[s.userSid]
	}

	if visibles := s.DenseInfo.GetVisible(); len(visibles) > 0 {
		s.info.Visible = visibles[s.index]
	}

	s.index++
}
