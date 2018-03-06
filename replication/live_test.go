package replication

import (
	"context"
	"os"
	"testing"
)

func liveOnly(t testing.TB) {
	if os.Getenv("LIVE_TEST") != "true" {
		t.Skipf("skipping live test, set LIVE_TEST=true to enable")
	}
}

func TestCurrentState(t *testing.T) {
	liveOnly(t)
	ctx := context.Background()

	_, _, err := CurrentMinuteState(ctx)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	_, _, err = CurrentHourState(ctx)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	_, _, err = CurrentDayState(ctx)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
}

func TestDownloadChanges(t *testing.T) {
	liveOnly(t)
	ctx := context.Background()

	_, err := Minute(ctx, 10)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	_, err = Hour(ctx, 10)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	_, err = Day(ctx, 1)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
}

func TestCurrentChangesetState(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	_, _, err := CurrentChangesetState(ctx)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
}

func TestChangesets(t *testing.T) {
	liveOnly(t)

	ctx := context.Background()
	sets, err := Changesets(ctx, 100)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if l := len(sets); l != 12 {
		t.Errorf("incorrect number of changesets: %v", l)
	}
}
