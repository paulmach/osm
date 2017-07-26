package osmpbf

import "github.com/paulmach/osm"

func extractTags(stringTable []string, keyIDs, valueIDs []uint32) osm.Tags {
	if len(keyIDs) == 0 {
		return nil
	}

	tags := make(osm.Tags, 0, len(keyIDs))
	for index, keyID := range keyIDs {
		tags = append(tags, osm.Tag{
			Key:   stringTable[keyID],
			Value: stringTable[valueIDs[index]],
		})
	}

	return tags
}

type tagUnpacker struct {
	stringTable []string
	keysVals    []int32
	index       int
}

// Next creates the tags from the stringtable and array of IDs.
// Used in DenseNodes encoding.
func (tu *tagUnpacker) Next() osm.Tags {
	index := tu.index
	for index < len(tu.keysVals) {
		if tu.keysVals[index] == 0 {
			break
		}

		index += 2
	}

	count := index - tu.index
	if count == 0 {
		tu.index++
		return nil
	}

	tags := make(osm.Tags, 0, count/2)
	for tu.index < len(tu.keysVals) {
		keyID := tu.keysVals[tu.index]
		tu.index++
		if keyID == 0 {
			break
		}

		valID := tu.keysVals[tu.index]
		tu.index++

		tags = append(tags, osm.Tag{
			Key:   tu.stringTable[keyID],
			Value: tu.stringTable[valID],
		})
	}

	return tags
}
