package osmapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChangeset_urls(t *testing.T) {
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

	t.Run("changeset", func(t *testing.T) {
		Changeset(ctx, 1)
		if !strings.Contains(url, "changeset/1") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("changeset with discussion", func(t *testing.T) {
		ChangesetWithDiscussion(ctx, 1)
		if !strings.Contains(url, "changeset/1?include_discussion=true") {
			t.Errorf("incorrect path: %v", url)
		}
	})

	t.Run("changeset download", func(t *testing.T) {
		ChangesetDownload(ctx, 1)
		if !strings.Contains(url, "changeset/1/download") {
			t.Errorf("incorrect path: %v", url)
		}
	})
}
