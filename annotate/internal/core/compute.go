package core

import (
	"context"
	"fmt"
	"time"

	"github.com/paulmach/osm"
)

// Datasourcer TODO
type Datasourcer interface {
	Get(ctx context.Context, id osm.FeatureID) (ChildList, error)
	NotFound(err error) bool
}

type parentChild struct {
	ChildID       osm.FeatureID
	ParentVersion int
}

// Options allow for passing som parameters to the matching process.
type Options struct {
	Threshold             time.Duration
	IgnoreInconsistency   bool
	IgnoreMissingChildren bool
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

	// elementMap is a reverse index of the specific version of a
	// child that is part of a given parent version. This is used to
	// determine the version of the member that is part of the base of
	// the next parent version. If there is gap, we know something changed
	// and need to insert a update there.
	elementMap, err := setupMajorChildren(ctx, parents, histories, opts)
	if err != nil {
		return nil, err
	}

	/////////////////////////////////////////////////////////////////
	// figure out if there are any child changes between parent versions.
	results := make([]osm.Updates, 0, len(parents))
	for i, parent := range parents {
		if !parent.Visible() {
			results = append(results, nil)
			continue
		}

		var (
			nextParent        Parent
			nextParentVersion int
		)

		if i < len(parents)-1 {
			nextParent = parents[i+1]
			nextParentVersion = nextParent.Version()
		}

		var updates osm.Updates
		for j, fid := range parent.Refs() {
			var nextVersion int
			c := elementMap[parentChild{
				ChildID:       fid,
				ParentVersion: parent.Version()}]

			child, err := histories.Get(ctx, fid)
			if err != nil {
				if histories.NotFound(err) && opts.IgnoreMissingChildren {
					continue
				}

				return nil, err
			}

			if nextParent == nil {
				// No next parent version, so we need to include all
				// future versions of this child.
				nextVersion = child[len(child)-1].VersionIndex() + 1
			} else {
				next := elementMap[parentChild{
					ChildID:       fid,
					ParentVersion: nextParentVersion}]
				if next == nil {
					// child is one of:
					// - not in the next parent version,
					// - next parent is deleted,
					// - data inconsistency and not visible for the next parent
					// so we need to know what was the last available before the next parent.

					// this timestamp will help to create updates,
					// we want to make sure it is:
					// - 1 threshold before the next parent
					// - not before the current child timestamp
					ts := timeThreshold(nextParent, -opts.Threshold)

					if c != nil && !ts.After(timeThreshold(c, 0)) { // before and equal still matches
						// visible in current but child and next parent are
						// within the same threshold, no updates.
						// i.e. next version is same as current version
						nextVersion = 0 // no updates.
					} else {
						// current child and next parent are far apart.
						next = child.VersionBefore(ts)
						if next == nil {
							// missing at current and next parent.
							nextVersion = 0 // no updates.
						} else {
							// visble or not, we want to want to include it.
							// novisible versions of this child will be filtered out below.
							nextVersion = next.VersionIndex() + 1
						}
					}
				} else {
					// if the child was updated enough before the next parent
					// include it in the minor versions.
					if timeThreshold(next, 0).Before(timeThreshold(nextParent, -opts.Threshold)) {
						nextVersion = next.VersionIndex() + 1
					} else {
						nextVersion = next.VersionIndex()
					}
				}
			}

			start := 0
			if c != nil {
				start = c.VersionIndex() + 1
			} else {
				// current child is not defined, is next child
				next := child.VersionBefore(timeThreshold(parent, 0))
				if next == nil {
					start = 0
				} else {
					start = next.VersionIndex() + 1
				}
			}

			for k := start; k < nextVersion; k++ {
				if child[k].Visible() {
					u := child[k].Update()
					u.Index = j
					updates = append(updates, u)
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
		}

		// we have what we need for this parent version.
		results = append(results, updates)
	}

	return results, nil
}

// setMajorChildren figures out the child versions that are active
// for each visible parent version.
func setupMajorChildren(
	ctx context.Context,
	parents []Parent,
	histories Datasourcer,
	opts *Options,
) (map[parentChild]Child, error) {
	elementMap := make(map[parentChild]Child)
	for _, p := range parents {
		if !p.Visible() {
			p.SetChildren(nil)
			continue
		}

		refs := p.Refs()
		cl := make(ChildList, len(refs))
		for j, ref := range refs {
			versions, err := histories.Get(ctx, ref)
			if err != nil {
				if !histories.NotFound(err) {
					return nil, err
				}

				if opts.IgnoreMissingChildren {
					continue
				}

				return nil, &NoHistoryError{ChildID: ref}
			}

			c := versions.FindVisible(
				p.ChangesetID(),
				timeThreshold(p, 0),
				opts.Threshold,
			)
			if c == nil && !opts.IgnoreInconsistency {
				return nil, &NoVisibleChildError{
					ChildID:   ref,
					Timestamp: timeThreshold(p, 0)}
			}

			cl[j] = c
			elementMap[parentChild{
				ChildID:       ref,
				ParentVersion: p.Version(),
			}] = c
		}

		p.SetChildren(cl)
	}

	return elementMap, nil
}
