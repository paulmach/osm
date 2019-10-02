package osm

import (
	"encoding/json"
	"sort"
)

// Polygon returns true if the way should be considered a closed polygon area.
// OpenStreetMap doesn't have an intrinsic area data type. The algorithm used
// here considers a set of heuristics to determine what is most likely an area.
// The heuristics can be found here,
// https://wiki.openstreetmap.org/wiki/Overpass_turbo/Polygon_Features
// and are used by osmtogeojson and overpass turbo.
func (w *Way) Polygon() bool {
	if len(w.Nodes) <= 3 {
		// need more than 3 nodes to be a polygon since first/last is repeated.
		return false
	}

	if w.Nodes[0].ID != w.Nodes[len(w.Nodes)-1].ID {
		// must be closed
		return false
	}

	if area := w.Tags.Find("area"); area == "no" {
		return false
	} else if area != "" {
		return true
	}

	for _, c := range polyConditions {
		v := w.Tags.Find(c.Key)
		if v == "" || v == "no" {
			continue
		}

		if c.Condition == conditionAll {
			return true
		} else if c.Condition == conditionWhitelist {
			index := sort.SearchStrings(c.Values, v)
			if index != len(c.Values) && c.Values[index] == v {
				return true
			}
		} else if c.Condition == conditionBlacklist {
			index := sort.SearchStrings(c.Values, v)
			if index == len(c.Values) || c.Values[index] != v {
				return true
			}
		}
	}

	return false
}

func init() {
	err := json.Unmarshal(polygonJSON, &polyConditions)
	if err != nil {
		// This must be valid json
		panic(err)
	}

	for _, p := range polyConditions {
		sort.StringSlice(p.Values).Sort()
	}
}

var polyConditions []polyCondition

type polyCondition struct {
	Key       string        `json:"key"`
	Condition conditionType `json:"polygon"`
	Values    []string      `json:"values"`
}
type conditionType string

var (
	conditionAll       conditionType = "all"
	conditionBlacklist conditionType = "blacklist"
	conditionWhitelist conditionType = "whitelist"
)

// polygonJSON holds advanced conditions for when an osm way is a polygon.
// Sourced from: https://wiki.openstreetmap.org/wiki/Overpass_turbo/Polygon_Features
// Also used by node lib: https://github.com/tyrasd/osmtogeojson
var polygonJSON = []byte(`
[
    {
        "key": "building",
        "polygon": "all"
    },
    {
        "key": "highway",
        "polygon": "whitelist",
        "values": [
            "services",
            "rest_area",
            "escape",
            "elevator"
        ]
    },
    {
        "key": "natural",
        "polygon": "blacklist",
        "values": [
            "coastline",
            "cliff",
            "ridge",
            "arete",
            "tree_row"
        ]
    },
    {
        "key": "landuse",
        "polygon": "all"
    },
    {
        "key": "waterway",
        "polygon": "whitelist",
        "values": [
            "riverbank",
            "dock",
            "boatyard",
            "dam"
        ]
    },
    {
        "key": "amenity",
        "polygon": "all"
    },
    {
        "key": "leisure",
        "polygon": "all"
    },
    {
        "key": "barrier",
        "polygon": "whitelist",
        "values": [
            "city_wall",
            "ditch",
            "hedge",
            "retaining_wall",
            "wall",
            "spikes"
        ]
    },
    {
        "key": "railway",
        "polygon": "whitelist",
        "values": [
            "station",
            "turntable",
            "roundhouse",
            "platform"
        ]
    },
    {
        "key": "boundary",
        "polygon": "all"
    },
    {
        "key": "man_made",
        "polygon": "blacklist",
        "values": [
            "cutline",
            "embankment",
            "pipeline"
        ]
    },
    {
        "key": "power",
        "polygon": "whitelist",
        "values": [
            "plant",
            "substation",
            "generator",
            "transformer"
        ]
    },
    {
        "key": "place",
        "polygon": "all"
    },
    {
        "key": "shop",
        "polygon": "all"
    },
    {
        "key": "aeroway",
        "polygon": "blacklist",
        "values": [
            "taxiway"
        ]
    },
    {
        "key": "tourism",
        "polygon": "all"
    },
    {
        "key": "historic",
        "polygon": "all"
    },
    {
        "key": "public_transport",
        "polygon": "all"
    },
    {
        "key": "office",
        "polygon": "all"
    },
    {
        "key": "building:part",
        "polygon": "all"
    },
    {
        "key": "military",
        "polygon": "all"
    },
    {
        "key": "ruins",
        "polygon": "all"
    },
    {
        "key": "area:highway",
        "polygon": "all"
    },
    {
        "key": "craft",
        "polygon": "all"
    },
    {
        "key": "golf",
        "polygon": "all"
    },
    {
        "key": "indoor",
        "polygon": "all"
    }
]`)

// Polygon returns true if the relation is of type multipolygon or boundary.
func (r *Relation) Polygon() bool {
	t := r.Tags.Find("type")
	return t == "multipolygon" || t == "boundary"
}
