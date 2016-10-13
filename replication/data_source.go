package replication

import (
	"fmt"
	"net/http"
	"time"
)

// BaseURL defines the planet server to hit.
const BaseURL = "http://planet.osm.org"

// DataSource defines context around replication data requests.
type DataSource struct {
	BaseURL string // will use package level BaseURL if empty
	Client  *http.Client
}

// DefaultDataSource is the DataSource used by the package level convenience functions.
var DefaultDataSource = &DataSource{
	Client: &http.Client{
		Timeout: 30 * time.Minute,
	},
}

// NewDataSource creates a DataSource using the given client.
func NewDataSource(client *http.Client) *DataSource {
	return &DataSource{
		Client: client,
	}
}

func (ds DataSource) baseURL() string {
	if ds.BaseURL != "" {
		return ds.BaseURL
	}

	return BaseURL
}

func (ds DataSource) client() *http.Client {
	if ds.Client != nil {
		return ds.Client
	}

	if DefaultDataSource.Client != nil {
		return DefaultDataSource.Client
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
