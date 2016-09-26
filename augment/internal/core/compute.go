package core

import (
	"fmt"
	"time"

	osm "github.com/paulmach/go.osm"
)

type parentChild struct {
	ChildID       ChildID
	ParentVersion int
}

// Compute does two things: first it computes the exact version of
// the children in each parent. Then it returns a set of updates
// for each version of the parent.
func Compute(
	parents []Parent,
	histories map[ChildID]ChildList,
	threshold time.Duration,
) ([]osm.Updates, error) {

	// elementMap is a reverse index of the specific version of a
	// child that is part of a given parent version. This is used to
	// determine the version of the member that is part of the base of
	// the next parent version. If there is gap, we know something changed
	// and need to insert a update there.
	elementMap, err := setupMajorChildren(parents, histories, threshold)
	if err != nil {
		return nil, err
	}

	/////////////////////////////////////////////////////////////////
	// figure out if there are any child changes between parent versions.
	results := make([]osm.Updates, 0, len(parents))
	for i, p := range parents {
		if !p.Visible() {
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
		for j, c := range p.Children() {
			var nextVersion int

			if nextParent == nil {
				// No next parent version, so we need to include all
				// future versions of this child.
				ns := histories[c.ID()]
				nextVersion = ns[len(ns)-1].VersionIndex() + 1
			} else {
				next := elementMap[parentChild{
					ChildID:       c.ID(),
					ParentVersion: nextParentVersion}]
				if next == nil {
					// child is not in the next parent version, or next parent is deleted,
					// so we need to know what was the last visible before it was removed
					// from the next parent.

					// this timestamp will help to create updates.
					// We want to make sure it is:
					// - 1 threshold before the next parent
					// - not before the current child timestamp
					ts := timeThreshold(nextParent, -threshold)
					if !ts.After(timeThreshold(c, 0)) { // before and equal still matches
						nextVersion = c.VersionIndex() // ie. no updates
					} else {
						next = histories[c.ID()].LastVisibleBefore(ts)
						if next == nil {
							// This a is a data inconsistency that should be looked at more closely.
							t, id := p.ID()
							return nil, fmt.Errorf("%v %d: %v: not visible at next parent timestamp %v",
								t, id, c.ID(), ts)
						}

						nextVersion = next.VersionIndex() + 1
					}
				} else {
					// if the child was updated enough before the next parent
					// include it in the minor versions.

					if timeThreshold(next, 0).Before(timeThreshold(nextParent, -threshold)) {
						nextVersion = next.VersionIndex() + 1
					} else {
						nextVersion = next.VersionIndex()
					}
				}
			}

			for k := c.VersionIndex() + 1; k < nextVersion; k++ {
				if histories[c.ID()][k].Visible() {
					u := histories[c.ID()][k].Update()
					u.Index = j
					updates = append(updates, u)
				} else {
					// A child has become not-visible between parent version.
					// This is a data inconsistency that needs to be looked into.
					t, id := p.ID()
					return nil, fmt.Errorf("%v %d: %v: child deleted between parent versions",
						t, id, c.ID())
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
	parents []Parent,
	histories map[ChildID]ChildList,
	threshold time.Duration,
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
			versions, ok := histories[ref]
			if !ok {
				return nil, &NoHistoryError{ChildID: ref}
			}

			c := versions.FindVisible(timeThreshold(p, 0), threshold)
			if c == nil {
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
