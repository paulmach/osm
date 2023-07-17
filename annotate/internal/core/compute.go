package core

import (
	"context"
	"fmt"
	"time"

	"github.com/onXmaps/osm"
	"github.com/onXmaps/osm/annotate/shared"
)

// A Datasourcer is something that acts like a datasource allowing us to
// fetch children as needed.
type Datasourcer interface {
	Get(ctx context.Context, id osm.FeatureID) (ChildList, error)
	NotFound(err error) bool
}

// childLoc references a location of a child in the parents + children.
type childLoc struct {
	Parent int
	Index  int
}

type childLocs []childLoc

// Options allow for passing som parameters to the matching process.
type Options struct {
	Threshold             time.Duration
	IgnoreInconsistency   bool
	IgnoreMissingChildren bool
	ChildFilter           func(osm.FeatureID) bool
}

// Compute does two things: first it computes the exact version of
// the children in each parent. Then it returns a set of updates
// for each version of the parent.
func Compute(
	ctx context.Context,
	parents []Parent,
	histories Datasourcer,
	opts *Options,
) ([]osm.Updates, error) {
	if opts == nil {
		opts = &Options{}
	}

	results := make([]osm.Updates, len(parents))
	for fid, locations := range mapChildLocs(parents, opts.ChildFilter) {
		child, err := histories.Get(ctx, fid)
		if err != nil {
			if !histories.NotFound(err) {
				return nil, err
			}

			if opts.IgnoreMissingChildren {
				continue
			}

			return nil, &NoHistoryError{ChildID: fid}
		}

		for _, locs := range locations.GroupByParent() {
			// figure out the parent and the next parent
			parentIndex := locs[0].Parent
			parent := parents[parentIndex]
			if !parent.Visible() {
				continue
			}

			var nextParent Parent
			if parentIndex < len(parents)-1 {
				nextParent = parents[parentIndex+1]
			}

			// get the current child
			c := child.FindVisible(
				parent.ChangesetID(),
				timeThresholdParent(parent, 0),
				opts.Threshold,
			)
			if c == nil && !opts.IgnoreInconsistency {
				return nil, &NoVisibleChildError{
					ChildID:   fid,
					Timestamp: timeThresholdParent(parent, 0)}
			}

			// straight up set this child on major version
			for _, cl := range locs {
				parent.SetChild(cl.Index, c)
			}

			// nextVersionIndex figures out what version of this child
			// is present in the next parent version
			nextVersion := nextVersionIndex(c, child, nextParent, opts)

			start := 0
			if c != nil {
				start = c.VersionIndex + 1
			} else {
				// current child is not defined, is next child
				next := child.VersionBefore(timeThresholdParent(parent, 0))
				if next == nil {
					start = 0
				} else {
					start = next.VersionIndex + 1
				}
			}

			var updates osm.Updates
			for k := start; k < nextVersion; k++ {
				if child[k].Visible {
					// It's possible for this child to be present at multiple locations in the parent
					for _, cl := range locs {
						u := child[k].Update()
						u.Index = cl.Index
						updates = append(updates, u)
					}
				} else {
					// A child has become not-visible between parent version.
					// This is a data inconsistency that can happen in old data
					// i.e. pre element versioning.
					//
					// see node 321452894, changed 7 times in
					// the same changeset, version 5 was a delete. (also node 65172196)
					if !opts.IgnoreInconsistency {
						return nil, fmt.Errorf("%v: %v: child deleted between parent versions",
							parent.ID(), fid)
					}
				}
			}

			// we have what we need for this parent version.
			results[parentIndex] = append(results[parentIndex], updates...)
		}
	}

	for _, r := range results {
		r.SortByIndex()
	}

	return results, nil
}

func nextVersionIndex(current *shared.Child, child ChildList, nextParent Parent, opts *Options) int {
	if nextParent == nil {
		// No next parent version, so we need to include all
		// future versions of this child.
		return child[len(child)-1].VersionIndex + 1
	}

	next := child.FindVisible(
		nextParent.ChangesetID(),
		timeThresholdParent(nextParent, 0),
		opts.Threshold,
	)

	if next != nil {
		// if the child was updated enough before the next parent
		// include it in the minor versions.
		if timeThreshold(next, 0).Before(timeThresholdParent(nextParent, -opts.Threshold)) {
			return next.VersionIndex + 1
		}

		return next.VersionIndex
	}

	// child is one of:
	// - not in the next parent version,
	// - next parent is deleted,
	// - data inconsistency and not visible for the next parent
	// so we need to know what was the last available before the next parent.

	// this timestamp will help to create updates,
	// we want to make sure it is:
	// - 1 threshold before the next parent
	// - not before the current child timestamp
	ts := timeThresholdParent(nextParent, -opts.Threshold)

	if current != nil && !ts.After(timeThreshold(current, 0)) { // before and equal still matches
		// visible in current but child and next parent are
		// within the same threshold, no updates.
		// i.e. next version is same as current version
		return 0 // no updates.
	}

	// current child and next parent are far apart.
	next = child.VersionBefore(ts)
	if next == nil {
		// missing at current and next parent.
		return 0 // no updates.
	}

	// visble or not, we want to want to include it.
	// novisible versions of this child will be filtered out below.
	return next.VersionIndex + 1
}

// mapChildLocs builds a cache of a where a child is in a set of parents.
func mapChildLocs(parents []Parent, filter func(osm.FeatureID) bool) map[osm.FeatureID]childLocs {
	result := make(map[osm.FeatureID]childLocs)
	for i, p := range parents {
		refs, annotated := p.Refs()
		for j, fid := range refs {
			if annotated[j] && filter != nil && !filter(fid) {
				continue
			}

			if result[fid] == nil {
				result[fid] = make([]childLoc, 0, len(parents))
			}

			result[fid] = append(result[fid], childLoc{Parent: i, Index: j})
		}
	}

	return result
}

func (locs childLocs) GroupByParent() []childLocs {
	var result []childLocs

	for len(locs) > 0 {
		p := locs[0].Parent
		end := 0

		for end < len(locs) && locs[end].Parent == p {
			end++
		}

		result = append(result, locs[:end])
		locs = locs[end:]
	}

	return result
}
