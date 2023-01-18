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

func TestRelation_urls(t *testing.T) {
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

	t.Run("relation", func(t *testing.T) {
		Relation(ctx, 1)
		if !strings.Contains(url, "relation/1") {
			t.Errorf("incorrect path: %v", url)
		}

		Relation(ctx, 1, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "relation/1?at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("relations", func(t *testing.T) {
		Relations(ctx, []osm.RelationID{1, 2})
		if !strings.Contains(url, "relations?relations=1,2") {
			t.Errorf("incorrect path: %v", url)
		}

		Relations(ctx, []osm.RelationID{1, 2}, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "relations?relations=1,2&at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("relation version", func(t *testing.T) {
		RelationVersion(ctx, 1, 2)
		if !strings.Contains(url, "relation/1/2") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("relation history", func(t *testing.T) {
		RelationHistory(ctx, 1)
		if !strings.Contains(url, "relation/1/history") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("relation relations", func(t *testing.T) {
		RelationRelations(ctx, 1)
		if !strings.Contains(url, "relation/1/relations") {
			t.Errorf("incorrect path: %v", url)
		}

		RelationRelations(ctx, 1, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "relation/1/relations?at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("relation full", func(t *testing.T) {
		RelationFull(ctx, 1)
		if !strings.Contains(url, "relation/1/full") {
			t.Errorf("incorrect path: %v", url)
		}

		RelationFull(ctx, 1, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "relation/1/full?at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})
}
