// Package osmapi provides an interface to the OSM v0.6 API.
package osmapi

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/paulmach/osm"
)

// BaseURL defines the api host. This can be change to hit
// a dev server, for example, http://api06.dev.openstreetmap.org/api/0.6
const BaseURL = "http://api.openstreetmap.org/api/0.6"

// A RateLimiter is something that can wait until its next allowed request.
// This interface is met by `golang.org/x/time/rate.Limiter` and is meant
// to be used with it. For example:
//
//		// 10 qps
//		osmapi.DefaultDatasource.Limiter = rate.NewLimiter(10, 1)
type RateLimiter interface {
	Wait(context.Context) error
}

// Datasource defines context about the http client to use to make requests.
type Datasource struct {
	// If Limiter is non-nil. The datasource will wait/block until the request
	// is allowed by the rate limiter. To be a good citizen, it is recommended
	// to use this when making may concurrent requests against the prod osm api.
	// See the RateLimiter docs for more information.
	Limiter RateLimiter

	BaseURL string
	Client  *http.Client
}

// DefaultDatasource is the Datasource used by package level convenience functions.
var DefaultDatasource = &Datasource{
	BaseURL: BaseURL,
	Client: &http.Client{
		Timeout: 6 * time.Minute, // looks like the api server has a 5 min timeout.
	},
}

var _ osm.HistoryDatasourcer = &Datasource{}

// NewDatasource creates a Datasource using the given client.
func NewDatasource(client *http.Client) *Datasource {
	return &Datasource{
		Client: client,
	}
}

func (ds *Datasource) getFromAPI(ctx context.Context, url string, item interface{}) error {
	client := ds.Client
	if client == nil {
		client = DefaultDatasource.Client
	}

	if client == nil {
		client = http.DefaultClient
	}

	if ds.Limiter != nil {
		err := ds.Limiter.Wait(ctx)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req.WithContext(ctx))
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

	if resp.StatusCode == http.StatusRequestURITooLong {
		return &RequestURITooLongError{URL: url}
	}

	if resp.StatusCode != http.StatusOK {
		return &UnexpectedStatusCodeError{
			Code: resp.StatusCode,
			URL:  url,
		}
	}

	return xml.NewDecoder(resp.Body).Decode(item)
}

func (ds *Datasource) baseURL() string {
	if ds.BaseURL != "" {
		return ds.BaseURL
	}

	return BaseURL
}

// NotFound error will return true if the result is not found.
func (ds *Datasource) NotFound(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(*NotFoundError)
	return ok
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

// RequestURITooLongError is returned when requesting too many ids in
// a multi id request, ie. Nodes, Ways, Relations functions.
type RequestURITooLongError struct {
	URL string
}

// Error returns an error message with the url causing the problem.
func (e *RequestURITooLongError) Error() string {
	return fmt.Sprintf("osmapi: uri too long at %s", e.URL)
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
