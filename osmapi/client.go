package osmapi

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

const host = "http://api.openstreetmap.org/api/0.6"

var httpClient = &http.Client{
	Timeout: 5 * time.Minute,
}

func getFromAPI(ctx context.Context, url string, item interface{}) error {
	resp, err := ctxhttp.Get(ctx, httpClient, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound{URL: url}
	}

	if resp.StatusCode == http.StatusGone {
		return ErrGone{URL: url}
	}

	if resp.StatusCode != http.StatusOK {
		return ErrUnexpectedStatusCode{
			Code: resp.StatusCode,
			URL:  url,
		}
	}

	return xml.NewDecoder(resp.Body).Decode(item)
}

// ErrNotFound means 404 from the api.
type ErrNotFound struct {
	URL string
}

// Error returns an error message with the url causing the problem.
func (e ErrNotFound) Error() string {
	return fmt.Sprintf("osmapi: not found at %s", e.URL)
}

// ErrGone is returned for deleted elements that get 410 from the api.
type ErrGone struct {
	URL string
}

// Error returns an error message with the url causing the problem.
func (e ErrGone) Error() string {
	return fmt.Sprintf("osmapi: gone at %s", e.URL)
}

// ErrUnexpectedStatusCode is return for a non 200 or 404 status code.
type ErrUnexpectedStatusCode struct {
	Code int
	URL  string
}

// Error returns an error message with some information.
func (e ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf("osmapi: unexpected status code of %d for url %s", e.Code, e.URL)
}
