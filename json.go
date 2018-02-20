package osm

import "encoding/xml"

// xmlNameJSONTypeNode is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeNode xml.Name

func (x xmlNameJSONTypeNode) MarshalJSON() ([]byte, error) {
	return []byte(`"node"`), nil
}

// xmlNameJSONTypeWay is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeWay xml.Name

func (x xmlNameJSONTypeWay) MarshalJSON() ([]byte, error) {
	return []byte(`"way"`), nil
}

// xmlNameJSONTypeRel is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeRel xml.Name

func (x xmlNameJSONTypeRel) MarshalJSON() ([]byte, error) {
	return []byte(`"relation"`), nil
}

// xmlNameJSONTypeCS is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeCS xml.Name

func (x xmlNameJSONTypeCS) MarshalJSON() ([]byte, error) {
	return []byte(`"changeset"`), nil
}

// xmlNameJSONTypeUser is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeUser xml.Name

func (x xmlNameJSONTypeUser) MarshalJSON() ([]byte, error) {
	return []byte(`"user"`), nil
}

// xmlNameJSONTypeNote is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeNote xml.Name

func (x xmlNameJSONTypeNote) MarshalJSON() ([]byte, error) {
	return []byte(`"note"`), nil
}
