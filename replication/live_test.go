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

	_, _, err := CurrentState(ctx)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
}

func TestDownloadChanges(t *testing.T) {
	liveOnly(t)
	ctx := context.Background()

	_, err := GetDiff(ctx, 10)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
}
