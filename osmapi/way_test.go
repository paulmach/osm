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

func TestWay_urls(t *testing.T) {
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

	t.Run("way", func(t *testing.T) {
		Way(ctx, 1)
		if !strings.Contains(url, "way/1") {
			t.Errorf("incorrect path: %v", url)
		}

		Way(ctx, 1, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "way/1?at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("ways", func(t *testing.T) {
		Ways(ctx, []osm.WayID{1, 2})
		if !strings.Contains(url, "ways?ways=1,2") {
			t.Errorf("incorrect path: %v", url)
		}

		Ways(ctx, []osm.WayID{1, 2}, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "ways?ways=1,2&at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("way version", func(t *testing.T) {
		WayVersion(ctx, 1, 2)
		if !strings.Contains(url, "way/1/2") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("way history", func(t *testing.T) {
		WayHistory(ctx, 1)
		if !strings.Contains(url, "way/1/history") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("way relations", func(t *testing.T) {
		WayRelations(ctx, 1)
		if !strings.Contains(url, "way/1/relations") {
			t.Errorf("incorrect path: %v", url)
		}

		WayRelations(ctx, 1, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "way/1/relations?at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("way full", func(t *testing.T) {
		WayFull(ctx, 1)
		if !strings.Contains(url, "way/1/full") {
			t.Errorf("incorrect path: %v", url)
		}

		WayFull(ctx, 1, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "way/1/full?at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})
}
