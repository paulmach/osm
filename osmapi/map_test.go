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

func TestMap_urls(t *testing.T) {
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

	t.Run("map", func(t *testing.T) {
		bound := &osm.Bounds{
			MinLon: 1, MinLat: 2,
			MaxLon: 3, MaxLat: 4,
		}

		Map(ctx, bound)
		if !strings.Contains(url, "map?bbox=1.000000,2.000000,3.000000,4.000000") {
			t.Errorf("incorrect path: %v", url)
		}

		Map(ctx, bound, At(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
		if !strings.Contains(url, "map?bbox=1.000000,2.000000,3.000000,4.000000&at=2016-01-01T00:00:00Z") {
			t.Errorf("incorrect path: %v", url)
		}
	})
}
