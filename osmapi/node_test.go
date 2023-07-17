package osmapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onXmaps/osm"
)

func TestNode_urls(t *testing.T) {
	ctx := context.Background()

	url := ""
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url = r.URL.String()
		w.Write([]byte(`<osm></osm>`))
	}))
	defer ts.Close()

	DefaultDatasource.BaseURL = ts.URL
	defer func() {
		DefaultDatasource.BaseURL = BaseURL
	}()

	t.Run("node", func(t *testing.T) {
		Node(ctx, 1)
		if !strings.Contains(url, "node/1") {
			t.Errorf("incorrect path: %v", url)
		}

		Node(ctx, 1, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "node/1?at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("nodes", func(t *testing.T) {
		Nodes(ctx, []osm.NodeID{1, 2})
		if !strings.Contains(url, "nodes?nodes=1,2") {
			t.Errorf("incorrect path: %v", url)
		}

		Nodes(ctx, []osm.NodeID{1, 2}, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "nodes?nodes=1,2&at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("node version", func(t *testing.T) {
		NodeVersion(ctx, 1, 2)
		if !strings.Contains(url, "node/1/2") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("node history", func(t *testing.T) {
		NodeHistory(ctx, 1)
		if !strings.Contains(url, "node/1/history") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("node ways", func(t *testing.T) {
		NodeWays(ctx, 1)
		if !strings.Contains(url, "node/1/ways") {
			t.Errorf("incorrect path: %v", url)
		}

		NodeWays(ctx, 1, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "node/1/ways?at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("node relations", func(t *testing.T) {
		NodeRelations(ctx, 1)
		if !strings.Contains(url, "node/1/relations") {
			t.Errorf("incorrect path: %v", url)
		}

		NodeRelations(ctx, 1, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "node/1/relations?at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})
}
