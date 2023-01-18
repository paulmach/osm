package osmapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onXmaps/osm"
)

func TestNote_urls(t *testing.T) {
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

	t.Run("note", func(t *testing.T) {
		Note(ctx, 1)
		if !strings.Contains(url, "notes/1") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("notes", func(t *testing.T) {
		bound := &osm.Bounds{
			MinLon: 1, MinLat: 2,
			MaxLon: 3, MaxLat: 4,
		}

		Notes(ctx, bound)
		if !strings.Contains(url, "notes?bbox=1.000000,2.000000,3.000000,4.000000") {
			t.Errorf("incorrect path: %v", url)
		}

		Notes(ctx, bound, Limit(1), MaxDaysClosed(4))
		if !strings.Contains(url, "notes?bbox=1.000000,2.000000,3.000000,4.000000&limit=1&closed=4") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("nodes search", func(t *testing.T) {
		NotesSearch(ctx, "asdf")
		if !strings.Contains(url, "notes/search?q=asdf") {
			t.Errorf("incorrect path: %v", url)
		}

		NotesSearch(ctx, "asdf", Limit(1), MaxDaysClosed(4))
		if !strings.Contains(url, "notes/search?q=asdf&limit=1&closed=4") {
			t.Errorf("incorrect path: %v", url)
		}
	})
}
