package annotate

import "github.com/paulmach/go.osm/annotate/internal/core"

// Option is a parameter that can be used for annotationg.
type Option func(*core.Options) error

// IgnoreInconsistency will try to match children even if they are missing.
// This should be used when you want to gracefully handle the weird data in OSM.
//
// Nodes with unclear/inconsistent data will not be annotated. Causes include:
// - redacted data: In 2012 due to the license change data had to be removed.
//   This could be some nodes of a way. There exist ways for which some nodes have
//   just a single delete version, e.g. way 159081205, node 376130526
// - data pre element versioning: pre-2012(?) data versions were not kept, so
//   for old ways there many be no information about some nodes. For example,
//   a node may be updated after a way and there is no way to get the original
//   version of the node and way.
// - bad editors: sometimes a node is edited 7 times in a single changeset
//   and version 5 is a delete. See node 321452894, part of way 28831147.
func IgnoreInconsistency(yes bool) Option {
	return func(o *core.Options) error {
		o.IgnoreInconsistency = yes
		return nil
	}
}
