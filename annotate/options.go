package annotate

import (
	"time"

	"github.com/nextmv-io/osm"
	"github.com/nextmv-io/osm/annotate/internal/core"
)

// Option is a parameter that can be used for annotating.
type Option func(*core.Options) error

const defaultThreshold = 30 * time.Minute

// Threshold is used if the "committed at" time is unknown and deals with
// the flexibility of commit orders, e.g. nodes in the same commit as
// the way can have a timestamp after the way. Threshold defines the time
// range to "forward group" these changes.
// Default 30 minutes.
func Threshold(t time.Duration) Option {
	return func(o *core.Options) error {
		o.Threshold = t
		return nil
	}
}

// IgnoreInconsistency will try to match children even if they are missing.
// This should be used when you want to gracefully handle the weird data in OSM.
//
// Nodes with unclear/inconsistent data will not be annotated. Causes include:
//   - redacted data: In 2012 due to the license change data had to be removed.
//     This could be some nodes of a way. There exist ways for which some nodes have
//     just a single delete version, e.g. way 159081205, node 376130526
//   - data pre element versioning: pre-2012(?) data versions were not kept, so
//     for old ways there many be no information about some nodes. For example,
//     a node may be updated after a way and there is no way to get the original
//     version of the node and way.
//   - bad editors: sometimes a node is edited 7 times in a single changeset
//     and version 5 is a delete. See node 321452894, part of way 28831147.
func IgnoreInconsistency(yes bool) Option {
	return func(o *core.Options) error {
		o.IgnoreInconsistency = yes
		return nil
	}
}

// IgnoreMissingChildren will ignore children for which the datasource returns
// datasource.ErrNotFound. This can be useful for partial history extracts where
// there may be relations for which the way was not included, e.g. a relation has
// a way inside the extract bounds and other ways outside the bounds.
func IgnoreMissingChildren(yes bool) Option {
	return func(o *core.Options) error {
		o.IgnoreMissingChildren = yes
		return nil
	}
}

// ChildFilter allows for only a subset of children to be annotated on the parent.
// This can greatly improve update speed by only worrying about the children
// updated in the same batch. All unannotated children will be annotated regardless
// of the results of the filter function.
func ChildFilter(filter func(osm.FeatureID) bool) Option {
	return func(o *core.Options) error {
		o.ChildFilter = filter
		return nil
	}
}
