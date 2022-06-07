package replication

import (
	"context"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

var baseTime = time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)

// buildState is helper to build "valid" state files.
func buildState(n int) *State {
	d := time.Duration(n)
	return &State{SeqNum: uint64(n), Timestamp: baseTime.Add(d * time.Hour)}
}

func TestSearchTimestamp(t *testing.T) {
	t.Skip()
	ctx := context.Background()

	ts := &stater{
		Min:     1,
		Current: func(ctx context.Context) (*State, error) { return buildState(50), nil },
		State: func(ctx context.Context, n uint64) (*State, error) {
			states := map[uint64]*State{
				5:  buildState(5),
				10: buildState(10),
				15: buildState(15),
				20: buildState(20),
				25: buildState(25),
				30: buildState(30),
			}

			s := states[n]
			if s == nil {
				return nil, &UnexpectedStatusCodeError{Code: http.StatusNotFound}
			}

			return s, nil
		},
	}

	cases := []struct {
		name     string
		time     time.Time
		expected uint64
	}{
		{
			name:     "before",
			time:     baseTime,
			expected: 5,
		},
		{
			name:     "after",
			time:     baseTime.Add(100 * time.Hour),
			expected: 50,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := searchTimestamp(ctx, ts, tc.time)
			if err != nil {
				t.Fatalf("error getting timestamp: %v", err)
			}

			if s.SeqNum != tc.expected {
				t.Errorf("incorrect seq number: %v != %v", s.SeqNum, tc.expected)
			}
		})
	}
}

func TestFindBound(t *testing.T) {
	ctx := context.Background()

	ts := &stater{
		Min:     1,
		Current: func(ctx context.Context) (*State, error) { return buildState(9), nil },
		State: func(ctx context.Context, n uint64) (*State, error) {
			states := map[uint64]*State{
				3: buildState(3),
				4: buildState(4),
				5: buildState(5),
				6: buildState(6),
				7: buildState(7),
				8: buildState(8),
			}

			s := states[n]
			if s == nil {
				return nil, &UnexpectedStatusCodeError{Code: http.StatusNotFound}
			}

			return s, nil
		},
	}

	cases := []struct {
		name  string
		time  time.Time
		lower uint64
		upper uint64
	}{
		{
			name:  "before",
			time:  baseTime,
			lower: 3,
			upper: 3,
		},
		{
			name:  "middle before",
			time:  baseTime.Add(6 * time.Hour).Add(30 * time.Minute),
			lower: 5,
			upper: 9,
		},
		{
			name:  "middle after",
			time:  baseTime.Add(4 * time.Hour).Add(30 * time.Minute),
			lower: 3,
			upper: 5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			upper, _ := ts.Current(ctx)
			lower, upper, err := findBound(ctx, ts, upper, tc.time)
			if err != nil {
				t.Fatalf("error getting timestamp: %v", err)
			}

			if lower.SeqNum != tc.lower {
				t.Errorf("incorrect lower seq number: %v != %v", lower.SeqNum, tc.lower)
			}

			if upper.SeqNum != tc.upper {
				t.Errorf("incorrect upper seq number: %v != %v", upper.SeqNum, tc.upper)
			}
		})
	}
}

func TestMinuteStateAt(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()

	base := time.Date(2012, 9, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC)
	diff := int64(now.Sub(base)/time.Second + 10)

	r := rand.New(rand.NewSource(42))

	for i := 0; i < 10; i++ {
		secs := r.Int63n(diff)
		timestamp := base.Add(time.Duration(secs) * time.Second)

		t.Logf("n: %d  timestamp: %v", i, timestamp)
		_, state, err := MinuteStateAt(ctx, timestamp)
		if err != nil {
			t.Fatalf("failed to get state: %v", err)
		}

		if timestamp.After(state.Timestamp) {
			_, current, err := CurrentMinuteState(ctx)
			if err != nil {
				t.Fatalf("could not get current state: %v", err)
			}

			if current.SeqNum != state.SeqNum {
				t.Logf("state: %+v", state)
				t.Fatalf("if timstamp is before, it must be the current timestamp")
			}

			continue
		}

		if state.SeqNum == minMinute {
			// timestamp was before the first state
			continue
		}

		// state's timestamp is after what we want
		// get previous to make sure it before what we want
		previous, err := MinuteState(ctx, MinuteSeqNum(state.SeqNum-1))
		if err != nil {
			t.Fatalf("could not get previous state: %v", err)
		}

		if !previous.Timestamp.Before(timestamp) {
			t.Logf("prev:  %+v", previous)
			t.Logf("state: %+v", state)
			t.Fatalf("previus state not before timestamp")
		}

		t.Logf("found: %+v", state)
	}
}

func TestChangesetStateAt(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()

	base := time.Date(2016, 9, 15, 0, 0, 0, 0, time.UTC)
	now := time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC)
	diff := int64(now.Sub(base)/time.Second + 10)

	r := rand.New(rand.NewSource(42))

	for i := 0; i < 10; i++ {
		secs := r.Int63n(diff)
		timestamp := base.Add(time.Duration(secs) * time.Second)

		t.Logf("n: %d  timestamp: %v", i, timestamp)
		_, state, err := ChangesetStateAt(ctx, timestamp)
		if err != nil {
			t.Fatalf("failed to get state: %v", err)
		}

		if timestamp.After(state.Timestamp) {
			_, current, err := CurrentChangesetState(ctx)
			if err != nil {
				t.Fatalf("could not get current state: %v", err)
			}

			if current.SeqNum != state.SeqNum {
				t.Logf("state: %+v", state)
				t.Fatalf("if timstamp is before, it must be the current timestamp")
			}

			continue
		}

		if state.SeqNum == minChangeset {
			// timestamp was before the first state
			continue
		}

		// state's timestamp is after what we want
		// get previous to make sure it before what we want
		previous, err := ChangesetState(ctx, ChangesetSeqNum(state.SeqNum-1))
		if err != nil {
			t.Fatalf("could not get previous state: %v", err)
		}

		if !previous.Timestamp.Before(timestamp) {
			t.Logf("prev:  %+v", previous)
			t.Logf("state: %+v", state)
			t.Fatalf("previus state not before timestamp")
		}

		t.Logf("found: %+v", state)
	}
}
