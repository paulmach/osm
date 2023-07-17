package osm

import (
	"encoding/json"
	"encoding/xml"
)

// CustomJSONMarshaler can be set to have the code use a different
// json marshaler than the default in the standard library.
// One use case in enabling `github.com/json-iterator/go`
// with something like this:
//
//	import (
//	  jsoniter "github.com/json-iterator/go"
//	  "github.com/onXmaps/osm"
//	)
//
//	var c = jsoniter.Config{
//	  EscapeHTML:              true,
//	  SortMapKeys:             false,
//	  MarshalFloatWith6Digits: true,
//	}.Froze()
//
//	osm.CustomJSONMarshaler = c
//	osm.CustomJSONUnmarshaler = c
//
// Note that any errors encountered during marshaling will be different.
var CustomJSONMarshaler interface {
	Marshal(v interface{}) ([]byte, error)
}

// CustomJSONUnmarshaler can be set to have the code use a different
// json unmarshaler than the default in the standard library.
// One use case in enabling `github.com/json-iterator/go`
// with something like this:
//
//	import (
//	  jsoniter "github.com/json-iterator/go"
//	  "github.com/onXmaps/osm"
//	)
//
//	var c = jsoniter.Config{
//	  EscapeHTML:              true,
//	  SortMapKeys:             false,
//	  MarshalFloatWith6Digits: true,
//	}.Froze()
//
//	osm.CustomJSONMarshaler = c
//	osm.CustomJSONUnmarshaler = c
//
// Note that any errors encountered during unmarshaling will be different.
var CustomJSONUnmarshaler interface {
	Unmarshal(data []byte, v interface{}) error
}

func marshalJSON(v interface{}) ([]byte, error) {
	if CustomJSONMarshaler == nil {
		return json.Marshal(v)
	}

	return CustomJSONMarshaler.Marshal(v)
}

func unmarshalJSON(data []byte, v interface{}) error {
	if CustomJSONUnmarshaler == nil {
		return json.Unmarshal(data, v)
	}

	return CustomJSONUnmarshaler.Unmarshal(data, v)
}

type nocopyRawMessage []byte

func (m *nocopyRawMessage) UnmarshalJSON(data []byte) error {
	*m = data
	return nil
}

// xmlNameJSONTypeNode is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeNode xml.Name

func (x xmlNameJSONTypeNode) MarshalJSON() ([]byte, error) {
	return []byte(`"node"`), nil
}

func (x xmlNameJSONTypeNode) UnmarshalJSON(data []byte) error {
	return nil
}

// xmlNameJSONTypeWay is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeWay xml.Name

func (x xmlNameJSONTypeWay) MarshalJSON() ([]byte, error) {
	return []byte(`"way"`), nil
}

func (x xmlNameJSONTypeWay) UnmarshalJSON(data []byte) error {
	return nil
}

// xmlNameJSONTypeRel is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeRel xml.Name

func (x xmlNameJSONTypeRel) MarshalJSON() ([]byte, error) {
	return []byte(`"relation"`), nil
}

func (x xmlNameJSONTypeRel) UnmarshalJSON(data []byte) error {
	return nil
}

// xmlNameJSONTypeCS is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeCS xml.Name

func (x xmlNameJSONTypeCS) MarshalJSON() ([]byte, error) {
	return []byte(`"changeset"`), nil
}

func (x xmlNameJSONTypeCS) UnmarshalJSON(data []byte) error {
	return nil
}

// xmlNameJSONTypeUser is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeUser xml.Name

func (x xmlNameJSONTypeUser) MarshalJSON() ([]byte, error) {
	return []byte(`"user"`), nil
}

func (x xmlNameJSONTypeUser) UnmarshalJSON(data []byte) error {
	return nil
}

// xmlNameJSONTypeNote is kind of a hack to encode the proper json
// object type attribute for this struct type.
type xmlNameJSONTypeNote xml.Name

func (x xmlNameJSONTypeNote) MarshalJSON() ([]byte, error) {
	return []byte(`"note"`), nil
}

func (x xmlNameJSONTypeNote) UnmarshalJSON(data []byte) error {
	return nil
}
