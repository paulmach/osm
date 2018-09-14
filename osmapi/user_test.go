package osmapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUser_urls(t *testing.T) {
	ctx := context.Background()

	url := ""
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url = r.URL.String()
		w.Write([]byte(`<osm></osm>`))
	}))
	defer ts.Close()

	DefaultDatasource.BaseURL = ts.URL

	t.Run("user", func(t *testing.T) {
		User(ctx, 1)
		if !strings.Contains(url, "user/1") {
			t.Errorf("incorrect path: %v", url)
		}
	})
}
