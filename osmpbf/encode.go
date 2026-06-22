package osmpbf

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf/internal/osmpbf"
	"google.golang.org/protobuf/proto"
)

// Encoder is an interface for encoding osm data. The format written is the osm
// pbf format.
type Encoder interface {
	io.Closer
	Encode(obj osm.Object) error
}

func NewEncoder(w io.Writer) (Encoder, error) {
	encoder := &encoder{
		writer:             w,
		reverseStringTable: make(map[string]int),
		compress:           true,
	}

	blockHeader := &osmpbf.HeaderBlock{
		RequiredFeatures: []string{"OsmSchema-V0.6", "DenseNodes"},
	}
	blockHeaderData, err := proto.Marshal(blockHeader)
	if err != nil {
		return nil, err
	}

	_, err = encoder.write(blockHeaderData, osmHeaderType)
	if err != nil {
		return nil, err
	}

	return encoder, nil
}

type encoder struct {
	writer             io.Writer
	reverseStringTable map[string]int
	lastWrittenType    osm.Type
	entities           []osm.Object
	mu                 sync.Mutex
	compress           bool
}

func (e *encoder) Close() error {
	return e.flush()
}

func (e *encoder) flush() error {
	if len(e.entities) == 0 {
		return nil
	}

	block := &osmpbf.PrimitiveBlock{}
	encode(block, e.reverseStringTable, e.entities, e.compress)
	e.entities = e.entities[:0]
	for k := range e.reverseStringTable {
		delete(e.reverseStringTable, k)
	}

	blockData, err := proto.Marshal(block)
	if err != nil {
		return nil
	}
	_, err = e.write(blockData, osmDataType)
	return err
}

func (e *encoder) write(data []byte, osmType string) (n int, err error) {
	blob := &osmpbf.Blob{}
	blob.RawSize = proto.Int32(int32(len(data)))
	if e.compress {
		target := &bytes.Buffer{}
		// compress the data
		writer := zlibWriter(target)
		writer.Write(data)
		writer.Close()
		blob.ZlibData = target.Bytes()
	} else {
		blob.Raw = data
	}

	blobData, err := proto.Marshal(blob)
	if err != nil {
		return 0, nil
	}

	blobHeader := &osmpbf.BlobHeader{
		Type:     proto.String(osmType),
		Datasize: proto.Int32(int32(len(blobData))),
	}

	blobHeaderData, err := proto.Marshal(blobHeader)
	if err != nil {
		return 0, nil
	}

	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(blobHeaderData)))

	if _, err = e.writer.Write(size); err != nil {
		return 0, nil
	}

	if _, err = e.writer.Write(blobHeaderData); err != nil {
		return 0, nil
	}

	return e.writer.Write(blobData)
}

func (e *encoder) Encode(obj osm.Object) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.lastWrittenType != obj.ObjectID().Type() {
		if err := e.flush(); err != nil {
			return err
		}
		e.lastWrittenType = obj.ObjectID().Type()
	}

	e.entities = append(e.entities, obj)
	if len(e.entities) >= 8000 {
		return e.flush()
	}
	return nil
}

func encode(block *osmpbf.PrimitiveBlock, reverseStringTable map[string]int, osmGeos []osm.Object, compressed bool) {
	groupIdx := 0
	i := 0

	if block.Stringtable != nil && block.Stringtable.S != nil {
		block.Stringtable.S = nil
	}

	for i < len(osmGeos) {
		var group *osmpbf.PrimitiveGroup = nil
		if groupIdx < len(block.Primitivegroup) {
			group = block.Primitivegroup[groupIdx]
			if group == nil {
				group = &osmpbf.PrimitiveGroup{}
			}
			if group.Dense != nil {
				if group.Dense.Denseinfo != nil {
					if group.Dense.Denseinfo.Changeset != nil {
						group.Dense.Denseinfo.Changeset = nil
					}
					if group.Dense.Denseinfo.Timestamp != nil {
						group.Dense.Denseinfo.Timestamp = nil
					}
					if group.Dense.Denseinfo.Uid != nil {
						group.Dense.Denseinfo.Uid = nil
					}
					if group.Dense.Denseinfo.UserSid != nil {
						group.Dense.Denseinfo.UserSid = nil
					}
					if group.Dense.Denseinfo.Version != nil {
						group.Dense.Denseinfo.Version = nil
					}
				}
				if group.Dense.Id != nil {
					group.Dense.Id = nil
				}
				if group.Dense.KeysVals != nil {
					group.Dense.KeysVals = nil
				}
				if group.Dense.Lat != nil {
					group.Dense.Lat = nil
				}
				if group.Dense.Lon != nil {
					group.Dense.Lon = nil
				}
			}
			if group.Changesets != nil {
				group.Changesets = nil
			}
			if group.Ways != nil {
				group.Ways = nil
			}
			if group.Relations != nil {
				group.Relations = nil
			}
		} else {
			group = &osmpbf.PrimitiveGroup{}
			block.Primitivegroup = append(block.Primitivegroup, group)
		}

		currentNodeCount := 0
		groupType := osmGeos[i].ObjectID().Type()
		previousNode := &osm.Node{}
		if groupType == osm.TypeNode && compressed && group.Dense == nil {
			group.Dense = &osmpbf.DenseNodes{Denseinfo: &osmpbf.DenseInfo{}}
		}
		for i < len(osmGeos) && osmGeos[i].ObjectID().Type() == groupType {
			switch groupType {
			case osm.TypeNode:
				if compressed {
					currentNode := osmGeos[i].(*osm.Node)
					EncodeDenseNode(block, reverseStringTable, group.Dense, currentNode, previousNode)
					previousNode = currentNode
				} else {
					if currentNodeCount < len(group.Nodes) {
						EncodeNode(block, reverseStringTable, group.Nodes[currentNodeCount], osmGeos[i].(*osm.Node))
					} else {
						pbfNode := &osmpbf.Node{}
						group.Nodes = append(group.Nodes, EncodeNode(block, reverseStringTable, pbfNode, osmGeos[i].(*osm.Node)))
					}
					currentNodeCount++
				}
			case osm.TypeWay:
				group.Ways = append(group.Ways, EncodeWay(block, reverseStringTable, osmGeos[i].(*osm.Way)))
			case osm.TypeRelation:
				group.Relations = append(group.Relations, EncodeRelation(block, reverseStringTable, osmGeos[i].(*osm.Relation)))
			}
			i++
		}

		if group.Nodes != nil {
			for currentNodeCount < len(group.Nodes) {
				group.Nodes = group.Nodes[:len(group.Nodes)-1]
			}
		}

		groupIdx++
	}

	if groupIdx < len(block.Primitivegroup) {
		block.Primitivegroup = block.Primitivegroup[:groupIdx]
	}
}

func EncodeDenseNode(block *osmpbf.PrimitiveBlock, reverseStringTable map[string]int, groupDense *osmpbf.DenseNodes, current, previous *osm.Node) {
	groupDense.Id = append(groupDense.Id, int64(current.ID-previous.ID))
	// cast block.Granularity to int64
	granularity := block.GetGranularity()
	currentLat := EncodeLatLon(current.Lat, block.GetLatOffset(), granularity)
	currentLon := EncodeLatLon(current.Lon, block.GetLonOffset(), granularity)
	previousLat := EncodeLatLon(previous.Lat, block.GetLatOffset(), granularity)
	previousLon := EncodeLatLon(previous.Lon, block.GetLonOffset(), granularity)
	latDiff := currentLat - previousLat
	lonDiff := currentLon - previousLon
	groupDense.Lat = append(groupDense.Lat, latDiff)
	groupDense.Lon = append(groupDense.Lon, lonDiff)

	if len(current.Tags) > 0 {
		for _, nodeTag := range current.Tags {
			groupDense.KeysVals = append(groupDense.KeysVals, EncodeString(block, reverseStringTable, nodeTag.Key))
			groupDense.KeysVals = append(groupDense.KeysVals, EncodeString(block, reverseStringTable, nodeTag.Value))
		}
	}
	groupDense.KeysVals = append(groupDense.KeysVals, 0)

	if groupDense.Denseinfo != nil {
		groupDense.Denseinfo.Changeset = append(groupDense.Denseinfo.Changeset, int64(current.ChangesetID-previous.ChangesetID))
		dateGranularity := block.GetDateGranularity()
		if !current.Timestamp.IsZero() {
			currentTimeStamp := EncodeTimestamp(current.Timestamp, dateGranularity)
			if !previous.Timestamp.IsZero() {
				previousTimeStamp := EncodeTimestamp(previous.Timestamp, dateGranularity)
				groupDense.Denseinfo.Timestamp = append(groupDense.Denseinfo.Timestamp, currentTimeStamp-previousTimeStamp)
			} else {
				groupDense.Denseinfo.Timestamp = append(groupDense.Denseinfo.Timestamp, currentTimeStamp)
			}
		}
		groupDense.Denseinfo.Uid = append(groupDense.Denseinfo.Uid, int32(current.UserID-previous.UserID))
		groupDense.Denseinfo.Version = append(groupDense.Denseinfo.Version, int32(current.Version-previous.Version))
		var previousUserNameId int32 = 0
		if previous.User != "" {
			previousUserNameId = EncodeString(block, reverseStringTable, previous.User)
		}
		currentUserNameId := EncodeString(block, reverseStringTable, current.User)
		groupDense.Denseinfo.UserSid = append(groupDense.Denseinfo.UserSid, currentUserNameId-previousUserNameId)
	}
}

func EncodeNode(block *osmpbf.PrimitiveBlock, reverseStringTable map[string]int, pbfNode *osmpbf.Node, node *osm.Node) *osmpbf.Node {
	castId := int64(node.ID)
	pbfNode.Id = &castId
	zero := int32(0)
	pbfNode.Info = &osmpbf.Info{
		Version: &zero,
	}
	if node.ChangesetID != 0 {
		changesetId := int64(node.ChangesetID)
		pbfNode.Info.Changeset = &changesetId
	}
	if !node.Timestamp.IsZero() {
		timeStamp := EncodeTimestamp(node.Timestamp, block.GetDateGranularity())
		pbfNode.Info.Timestamp = &timeStamp
	}
	if node.UserID != 0 {
		userId := int32(node.UserID)
		pbfNode.Info.Uid = &userId
	}
	userSid := uint32(EncodeString(block, reverseStringTable, node.User))
	pbfNode.Info.UserSid = &userSid
	if node.Version != 0 {
		nodeVersion := int32(node.Version)
		pbfNode.Info.Version = &nodeVersion
	}
	lat := EncodeLatLon(node.Lat, block.GetLatOffset(), block.GetGranularity())
	pbfNode.Lat = &lat
	lon := EncodeLatLon(node.Lon, block.GetLonOffset(), block.GetGranularity())
	pbfNode.Lon = &lon

	if len(node.Tags) > 0 {
		for _, tag := range node.Tags {
			pbfNode.Keys = append(pbfNode.Keys, uint32(EncodeString(block, reverseStringTable, tag.Key)))
			pbfNode.Vals = append(pbfNode.Vals, uint32(EncodeString(block, reverseStringTable, tag.Value)))
		}
	} else {
		pbfNode.Keys = nil
		pbfNode.Vals = nil
	}
	return pbfNode
}

func EncodeWay(block *osmpbf.PrimitiveBlock, reverseStringTable map[string]int, way *osm.Way) *osmpbf.Way {
	pbfWay := &osmpbf.Way{
		Id:   (*int64)(&way.ID),
		Info: &osmpbf.Info{},
	}
	if way.ChangesetID != 0 {
		pbfWay.Info.Changeset = (*int64)(&way.ChangesetID)
	}
	if !way.Timestamp.IsZero() {
		timestamp := EncodeTimestamp(way.Timestamp, block.GetDateGranularity())
		pbfWay.Info.Timestamp = &timestamp
	}
	if way.UserID != 0 {
		userId := int32(way.UserID)
		pbfWay.Info.Uid = &userId
	}
	userId := uint32(EncodeString(block, reverseStringTable, way.User))
	pbfWay.Info.UserSid = &userId
	if way.Version != 0 {
		wayVersion := int32(way.Version)
		pbfWay.Info.Version = &wayVersion
	}

	if len(way.Tags) > 0 {
		for _, tag := range way.Tags {
			pbfWay.Keys = append(pbfWay.Keys, uint32(EncodeString(block, reverseStringTable, tag.Key)))
			pbfWay.Vals = append(pbfWay.Vals, uint32(EncodeString(block, reverseStringTable, tag.Value)))
		}
	}

	if len(way.Nodes) > 0 {
		pbfWay.Refs = append(pbfWay.Refs, int64(way.Nodes[0].ID))
		for i := 1; i < len(way.Nodes); i++ {
			pbfWay.Refs = append(pbfWay.Refs, int64(way.Nodes[i].ID)-int64(way.Nodes[i-1].ID))
		}
	}
	return pbfWay
}

func EncodeRelation(block *osmpbf.PrimitiveBlock, reverseStringTable map[string]int, relation *osm.Relation) *osmpbf.Relation {
	pbfRelation := &osmpbf.Relation{
		Id: (*int64)(&relation.ID),
		Info: &osmpbf.Info{
			Changeset: (*int64)(&relation.ChangesetID),
		},
	}
	if !relation.Timestamp.IsZero() {
		timestamp := EncodeTimestamp(relation.Timestamp, block.GetDateGranularity())
		pbfRelation.Info.Timestamp = &timestamp
	}
	if relation.UserID != 0 {
		userId := int32(relation.UserID)
		pbfRelation.Info.Uid = &userId
	}
	userSid := uint32(EncodeString(block, reverseStringTable, relation.User))
	pbfRelation.Info.UserSid = &userSid
	relationVersion := int32(relation.Version)
	pbfRelation.Info.Version = &relationVersion

	if len(relation.Tags) > 0 {
		for _, tag := range relation.Tags {
			pbfRelation.Keys = append(pbfRelation.Keys, uint32(EncodeString(block, reverseStringTable, tag.Key)))
			pbfRelation.Vals = append(pbfRelation.Vals, uint32(EncodeString(block, reverseStringTable, tag.Value)))
		}
	}

	if len(relation.Members) > 0 {
		pbfRelation.Memids = append(pbfRelation.Memids, relation.Members[0].Ref)
		pbfRelation.RolesSid = append(pbfRelation.RolesSid, EncodeString(block, reverseStringTable, relation.Members[0].Role))
		switch relation.Members[0].Type {
		case osm.TypeNode:
			pbfRelation.Types = append(pbfRelation.Types, osmpbf.Relation_NODE)
		case osm.TypeWay:
			pbfRelation.Types = append(pbfRelation.Types, osmpbf.Relation_WAY)
		case osm.TypeRelation:
			pbfRelation.Types = append(pbfRelation.Types, osmpbf.Relation_RELATION)
		}
		for i := 1; i < len(relation.Members); i++ {
			pbfRelation.Memids = append(pbfRelation.Memids, relation.Members[i].Ref-relation.Members[i-1].Ref)
			pbfRelation.RolesSid = append(pbfRelation.RolesSid, EncodeString(block, reverseStringTable, relation.Members[i].Role))
			switch relation.Members[i].Type {
			case osm.TypeNode:
				pbfRelation.Types = append(pbfRelation.Types, osmpbf.Relation_NODE)
			case osm.TypeWay:
				pbfRelation.Types = append(pbfRelation.Types, osmpbf.Relation_WAY)
			case osm.TypeRelation:
				pbfRelation.Types = append(pbfRelation.Types, osmpbf.Relation_RELATION)
			}
		}
	}
	return pbfRelation
}

func EncodeLatLon(value float64, offset int64, granularity int32) int64 {
	return (int64(value/.000000001) - offset) / int64(granularity)
}

func EncodeTimestamp(timestamp time.Time, dateGranularity int32) int64 {
	ts := timestamp.UnixMilli() / int64(dateGranularity)
	if ts < 0 {
		return 0
	}
	return ts
}

func EncodeString(block *osmpbf.PrimitiveBlock, reverseStringTable map[string]int, value string) int32 {
	if value == "" {
		return 0
	}

	if block.Stringtable == nil {
		block.Stringtable = &osmpbf.StringTable{}
		block.Stringtable.S = append(block.Stringtable.S, "")
		reverseStringTable[""] = 0
	}

	if id, ok := reverseStringTable[value]; ok {
		return int32(id)
	}

	block.Stringtable.S = append(block.Stringtable.S, value)
	id := len(block.Stringtable.S) - 1
	reverseStringTable[value] = id
	return int32(id)
}
