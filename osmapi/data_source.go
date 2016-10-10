package osmapi

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// BaseURL defines the api host. This can be change to hit
// a dev server, for example, http://api06.dev.openstreetmap.org/api/0.6
const BaseURL = "http://api.openstreetmap.org/api/0.6"

// DataSource defines context about the http client to use to make requests.
type DataSource struct {
	BaseURL string
	Client  *http.Client
}

// DefaultDataSource is the DataSource used by package level convenience functions.
var DefaultDataSource = &DataSource{
	Client: &http.Client{
		Timeout: 6 * time.Minute, // looks like the api server has a 5 min timeout.
	},
}

// NewDataSource creates a DataSource using the given client.
func NewDataSource(client *http.Client) *DataSource {
	return &DataSource{
		Client: client,
	}
}

func (ds *DataSource) getFromAPI(ctx context.Context, url string, item interface{}) error {
	client := ds.Client
	if client == nil {
		client = DefaultDataSource.Client
	}

	if client == nil {
		client = http.DefaultClient
	}

	resp, err := ctxhttp.Get(ctx, client, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{URL: url}
	}

	if resp.StatusCode == http.StatusForbidden {
		return &ForbiddenError{URL: url}
	}

	if resp.StatusCode == http.StatusGone {
		return &GoneError{URL: url}
	}

	if resp.StatusCode != http.StatusOK {
		return &UnexpectedStatusCodeError{
			Code: resp.StatusCode,
			URL:  url,
		}
	}

	return xml.NewDecoder(resp.Body).Decode(item)
}

func (ds *DataSource) baseURL() string {
	if ds.BaseURL != "" {
		return ds.BaseURL
	}

	return BaseURL
}

// NotFoundError means 404 from the api.
type NotFoundError struct {
	URL string
}

// Error returns an error message with the url causing the problem.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("osmapi: not found at %s", e.URL)
}

// ForbiddenError means 403 from the api.
// Returned whenever the version of the element is not available (due to redaction).
type ForbiddenError struct {
	URL string
}

// Error returns an error message with the url causing the problem.
func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("osmapi: forbidden at %s", e.URL)
}

// GoneError is returned for deleted elements that get 410 from the api.
type GoneError struct {
	URL string
}

// Error returns an error message with the url causing the problem.
func (e *GoneError) Error() string {
	return fmt.Sprintf("osmapi: gone at %s", e.URL)
}

// UnexpectedStatusCodeError is return for a non 200 or 404 status code.
type UnexpectedStatusCodeError struct {
	Code int
	URL  string
}

// Error returns an error message with some information.
func (e *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("osmapi: unexpected status code of %d for url %s", e.Code, e.URL)
}
