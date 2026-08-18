package osmpbf

import (
	"bytes"
	"context"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/paulmach/osm/osmpbf/internal/osmpbf"
	"google.golang.org/protobuf/proto"
)

// writeFileBlock writes a single OSM PBF file block (a big-endian uint32
// BlobHeader length, the BlobHeader, then the uncompressed Blob) to buf.
func writeFileBlock(t *testing.T, buf *bytes.Buffer, blockType string, payload []byte) {
	t.Helper()

	blob, err := proto.Marshal(&osmpbf.Blob{Raw: payload})
	if err != nil {
		t.Fatal(err)
	}

	dataSize := int32(len(blob))
	header, err := proto.Marshal(&osmpbf.BlobHeader{
		Type:     proto.String(blockType),
		Datasize: &dataSize,
	})
	if err != nil {
		t.Fatal(err)
	}

	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(header)))
	buf.Write(size[:])
	buf.Write(header)
	buf.Write(blob)
}

// pbfWithGroup returns a minimal, valid-framing PBF stream: an OSMHeader block
// followed by an OSMData block whose PrimitiveBlock has a single-entry string
// table and the given primitive group.
func pbfWithGroup(t *testing.T, group *osmpbf.PrimitiveGroup) []byte {
	t.Helper()

	var buf bytes.Buffer

	header, err := proto.Marshal(&osmpbf.HeaderBlock{
		RequiredFeatures: []string{"OsmSchema-V0.6", "DenseNodes"},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeFileBlock(t, &buf, "OSMHeader", header)

	block, err := proto.Marshal(&osmpbf.PrimitiveBlock{
		Stringtable:    &osmpbf.StringTable{S: []string{""}},
		Primitivegroup: []*osmpbf.PrimitiveGroup{group},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeFileBlock(t, &buf, "OSMData", block)

	return buf.Bytes()
}

// A string table index that references an entry beyond the (single-entry)
// string table used by pbfWithGroup.
const outOfRangeIndex = 5

// TestOutOfRangeStringTableIndex verifies that a string table index pointing
// past the end of the table produces an error instead of panicking the
// decoder. The key/value/role/user fields of ways, relations and dense nodes
// are all indexes into the block's string table and are attacker controlled.
func TestOutOfRangeStringTableIndex(t *testing.T) {
	cases := []struct {
		name  string
		group *osmpbf.PrimitiveGroup
	}{
		{
			name: "way tag key",
			group: &osmpbf.PrimitiveGroup{Ways: []*osmpbf.Way{{
				Id:   proto.Int64(1),
				Keys: []uint32{outOfRangeIndex},
				Vals: []uint32{outOfRangeIndex},
			}}},
		},
		{
			name: "relation member role",
			group: &osmpbf.PrimitiveGroup{Relations: []*osmpbf.Relation{{
				Id:       proto.Int64(1),
				RolesSid: []int32{outOfRangeIndex},
				Memids:   []int64{1},
				Types:    []osmpbf.Relation_MemberType{osmpbf.Relation_NODE},
			}}},
		},
		{
			name: "dense node tag",
			group: &osmpbf.PrimitiveGroup{Dense: &osmpbf.DenseNodes{
				Id:       []int64{1},
				Lat:      []int64{0},
				Lon:      []int64{0},
				KeysVals: []int32{outOfRangeIndex, outOfRangeIndex, 0},
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := pbfWithGroup(t, tc.group)

			scanner := New(context.Background(), bytes.NewReader(data), 1)
			defer scanner.Close()

			for scanner.Scan() {
			}

			err := scanner.Err()
			if err == nil {
				t.Fatal("expected an error for the out-of-range string table index, got nil")
			}
			if !strings.Contains(err.Error(), "string table index") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
