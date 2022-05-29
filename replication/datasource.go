package replication

import (
	"fmt"
	"net/http"
	"time"
)

// BaseURL defines the planet server to hit.
const BaseURL = "http://planet.osm.org"

// Datasource defines context around replication data requests.
type Datasource struct {
	BaseURL string // will use package level BaseURL if empty
	Client  *http.Client
}

// DefaultDatasource is the Datasource used by the package level convenience functions.
var DefaultDatasource = &Datasource{
	Client: &http.Client{
		Timeout: 30 * time.Minute,
	},
}

// NewDatasource creates a Datasource using the given client.
func NewDatasource(client *http.Client) *Datasource {
	return &Datasource{
		Client: client,
	}
}

func (ds Datasource) baseURL() string {
	if ds.BaseURL != "" {
		return ds.BaseURL
	}

	return BaseURL
}

func (ds Datasource) client() *http.Client {
	if ds.Client != nil {
		return ds.Client
	}

	if DefaultDatasource.Client != nil {
		return DefaultDatasource.Client
	}

	return http.DefaultClient
}

// UnexpectedStatusCodeError is return for a non 200 or 404 status code.
type UnexpectedStatusCodeError struct {
	Code int
	URL  string
}

// Error returns an error message with some information.
func (e *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("replication: unexpected status code of %d for url %s", e.Code, e.URL)
}

// NotFound will return try if the error from one of the methods was due
// to the file not found on the remote host.
func NotFound(err error) bool {
	if e, ok := err.(*UnexpectedStatusCodeError); ok {
		return e.Code == http.StatusNotFound
	}

	return false
}

// timeFormats contains the set of different formats we've see the time in.
var timeFormats = []string{
	"2006-01-02 15:04:05.999999999 Z",
	"2006-01-02 15:04:05.999999999 +00:00",
	"2006-01-02T15\\:04\\:05Z",
}

func decodeTime(s string) (time.Time, error) {
	var (
		t   time.Time
		err error
	)
	for _, format := range timeFormats {
		t, err = time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}

	return t, err
}
