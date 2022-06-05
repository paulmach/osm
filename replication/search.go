package replication

import (
	"context"
	"time"
)

// the valid minimum state number on planet.osm.org
const (
	minMinute = 1 // up to 2012-09-12T08:15:45Z
	minHour   = 1 // up to 2013-07-14T12:00:00Z
	minDay    = 1 // up to 2012-09-13T00:00:00Z

	// There are changes before this, but no state.
	minChangeset = 2007990 // 2016-09-07 10:45:02.148547780 Z
)

type stater struct {
	Min     uint64
	Current func(context.Context) (*State, error)
	State   func(context.Context, uint64) (*State, error)
}

// MinuteStateAt will return the replication state/sequence number that contains
// data for the given timestamp. This would be the first replication state written
// after the timestamp. If the timestamp is after all current replication state
// the most recent will be returned. The caller can check for this case using
// state.Before(givenTimestamp).
//
// This call can do 20+ requests to the binary search the replication states.
// Use sparingly or use a mirror.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func MinuteStateAt(ctx context.Context, timestamp time.Time) (MinuteSeqNum, *State, error) {
	return DefaultDatasource.MinuteStateAt(ctx, timestamp)
}

// MinuteStateAt will return the replication state/sequence number that contains
// data for the given timestamp. This would be the first replication state written
// after the timestamp. If the timestamp is after all current replication state
// the most recent will be returned. The caller can check for this case using
// state.Before(givenTimestamp).
//
// This call can do 20+ requests to the binary search the replication states.
// Use sparingly or use a mirror.
func (ds *Datasource) MinuteStateAt(ctx context.Context, timestamp time.Time) (MinuteSeqNum, *State, error) {
	s := &stater{
		Min: minMinute,
		Current: func(ctx context.Context) (*State, error) {
			_, s, err := ds.CurrentMinuteState(ctx)
			return s, err
		},
		State: func(ctx context.Context, n uint64) (*State, error) {
			return ds.MinuteState(ctx, MinuteSeqNum(n))
		},
	}
	state, err := searchTimestamp(ctx, s, timestamp)
	if err != nil {
		return 0, nil, err
	}

	return MinuteSeqNum(state.SeqNum), state, nil
}

// HourStateAt will return the replication state/sequence number that contains
// data for the given timestamp. This would be the first replication state written
// after the timestamp. If the timestamp is after all current replication state
// the most recent will be returned. The caller can check for this case using
// state.Before(givenTimestamp).
//
// This call can do 20+ requests to the binary search the replication states.
// Use sparingly or use a mirror.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func HourStateAt(ctx context.Context, timestamp time.Time) (HourSeqNum, *State, error) {
	return DefaultDatasource.HourStateAt(ctx, timestamp)
}

// HourStateAt will return the replication state/sequence number that contains
// data for the given timestamp. This would be the first replication state written
// after the timestamp. If the timestamp is after all current replication state
// the most recent will be returned. The caller can check for this case using
// state.Before(givenTimestamp).
//
// This call can do 20+ requests to the binary search the replication states.
// Use sparingly or use a mirror.
func (ds *Datasource) HourStateAt(ctx context.Context, timestamp time.Time) (HourSeqNum, *State, error) {
	s := &stater{
		Min: minHour,
		Current: func(ctx context.Context) (*State, error) {
			_, s, err := ds.CurrentHourState(ctx)
			return s, err
		},
		State: func(ctx context.Context, n uint64) (*State, error) {
			return ds.HourState(ctx, HourSeqNum(n))
		},
	}
	state, err := searchTimestamp(ctx, s, timestamp)
	if err != nil {
		return 0, nil, err
	}

	return HourSeqNum(state.SeqNum), state, nil
}

// DayStateAt will return the replication state/sequence number that contains
// data for the given timestamp. This would be the first replication state written
// after the timestamp. If the timestamp is after all current replication state
// the most recent will be returned. The caller can check for this case using
// state.Before(givenTimestamp).
//
// This call can do 20+ requests to the binary search the replication states.
// Use sparingly or use a mirror.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func DayStateAt(ctx context.Context, timestamp time.Time) (DaySeqNum, *State, error) {
	return DefaultDatasource.DayStateAt(ctx, timestamp)
}

// DayStateAt will return the replication state/sequence number that contains
// data for the given timestamp. This would be the first replication state written
// after the timestamp. If the timestamp is after all current replication state
// the most recent will be returned. The caller can check for this case using
// state.Before(givenTimestamp).
//
// This call can do 20+ requests to the binary search the replication states.
// Use sparingly or use a mirror.
func (ds *Datasource) DayStateAt(ctx context.Context, timestamp time.Time) (DaySeqNum, *State, error) {
	s := &stater{
		Min: minDay,
		Current: func(ctx context.Context) (*State, error) {
			_, s, err := ds.CurrentDayState(ctx)
			return s, err
		},
		State: func(ctx context.Context, n uint64) (*State, error) {
			return ds.DayState(ctx, DaySeqNum(n))
		},
	}
	state, err := searchTimestamp(ctx, s, timestamp)
	if err != nil {
		return 0, nil, err
	}

	return DaySeqNum(state.SeqNum), state, nil
}

// ChangesetStateAt will return the replication state/sequence number that contains
// data for the given timestamp. This would be the first replication state written
// after the timestamp. If the timestamp is after all current replication state
// the most recent will be returned. The caller can check for this case using
// state.Before(givenTimestamp).
//
// This call can do 20+ requests to the binary search the replication states.
// Use sparingly or use a mirror.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func ChangesetStateAt(ctx context.Context, timestamp time.Time) (ChangesetSeqNum, *State, error) {
	return DefaultDatasource.ChangesetStateAt(ctx, timestamp)
}

// ChangesetStateAt will return the replication state/sequence number that contains
// data for the given timestamp. This would be the first replication state written
// after the timestamp. If the timestamp is after all current replication state
// the most recent will be returned. The caller can check for this case using
// state.Before(givenTimestamp).
//
// This call can do 20+ requests to the binary search the replication states.
// Use sparingly or use a mirror.
func (ds *Datasource) ChangesetStateAt(ctx context.Context, timestamp time.Time) (ChangesetSeqNum, *State, error) {
	s := &stater{
		Min: minDay,
		Current: func(ctx context.Context) (*State, error) {
			_, s, err := ds.CurrentChangesetState(ctx)
			return s, err
		},
		State: func(ctx context.Context, n uint64) (*State, error) {
			return ds.ChangesetState(ctx, ChangesetSeqNum(n))
		},
	}
	state, err := searchTimestamp(ctx, s, timestamp)
	if err != nil {
		return 0, nil, err
	}

	return ChangesetSeqNum(state.SeqNum), state, nil
}

func searchTimestamp(ctx context.Context, s *stater, timestamp time.Time) (*State, error) {
	// get the current timestamp from the server
	upper, err := s.Current(ctx)
	if NotFound(err) {
		return nil, err // current state not found?
	} else if err != nil {
		return nil, err
	}

	if timestamp.After(upper.Timestamp) {
		return upper, nil // given time is in the future or something
	}

	lower, err := s.State(ctx, s.Min)
	if err != nil && !NotFound(err) {
		return nil, err
	}

	if lower == nil {
		// now we need to find a lower bound state manually.
		// This can have edge cases if there are missing sequence numbers.
		var err error
		lower, upper, err = findBound(ctx, s, upper, timestamp)
		if err != nil {
			return nil, err
		}
	}

	if lower.SeqNum+1 >= upper.SeqNum {
		return lower, nil // edge case if there are only one or two sequence numbers
	}

	return findInRange(ctx, s, lower, upper, timestamp)
}

func findBound(ctx context.Context, s *stater, upper *State, timestamp time.Time) (*State, *State, error) {
	var (
		lowerID uint64 = 1
		lower   *State
		err     error
	)

	// we need to find the lower bound
	for lower == nil {
		lower, err = s.State(ctx, lowerID)

		if err != nil && !NotFound(err) {
			return nil, nil, err
		}

		if lower != nil && lower.Timestamp.After(timestamp) {
			if lower.SeqNum+1 >= upper.SeqNum {
				return lower, upper, nil // edge case if there are only two sequence numbers
			}

			// in our search for lower we found a new upper bound
			upper = lower
			lower = nil
			lowerID = 1
		}

		if lower != nil {
			break
		}

		// no lower yet, so try a higher id (binary search wise)
		newID := (lowerID + upper.SeqNum) / 2
		if newID <= lowerID {
			// nothing suitable found, so upper is probably the best we can do
			return upper, upper, nil
		}
		lowerID = newID
	}

	return lower, upper, nil
}

func findInRange(ctx context.Context, s *stater, lower, upper *State, timestamp time.Time) (*State, error) {
	// we do a binary search through the range to find the sequence number
	for lower.SeqNum+1 < upper.SeqNum {
		// could do better here
		splitID := (lower.SeqNum + upper.SeqNum) / 2

		split, err := s.State(ctx, splitID)
		if err != nil && !NotFound(err) {
			return nil, err
		}

		if split == nil {
			// file missing, search the next towards lower
			sID := splitID - 1

			for split == nil && lower.SeqNum < splitID {
				split, err = s.State(ctx, sID)
				if err != nil && !NotFound(err) {
					return nil, err
				}

				sID--
			}
		}

		if split == nil {
			// still missing? search the next towards upper
			sID := splitID + 1

			for split == nil && splitID < upper.SeqNum {
				split, err = s.State(ctx, sID)
				if err != nil && !NotFound(err) {
					return nil, err
				}

				sID++
			}
		}

		if split == nil {
			// still nothing
			return lower, nil
		}

		// set the new boundary
		if timestamp.After(split.Timestamp) {
			lower = split
		} else {
			upper = split
		}
	}

	// timestamp is now between lower and upper, we want to return the upper.
	return upper, nil
}
