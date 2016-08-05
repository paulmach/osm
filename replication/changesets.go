package replication

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/paulmach/go.osm"
	"github.com/paulmach/go.osm/osmutil"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

const planetHost = "http://planet.osm.org"

var httpClient = &http.Client{
	Timeout: 5 * time.Minute,
}

// ChangesetSeqID indicates the sequence of the changeset replication found here:
// http://planet.osm.org/replication/changesets/
type ChangesetSeqID uint

// ChangesetState returns the current state of the changeset replication.
func ChangesetState(ctx context.Context) (ChangesetSeqID, time.Time, error) {
	resp, err := ctxhttp.Get(ctx, httpClient, planetHost+"/replication/changesets/state.yaml")
	if err != nil {
		return 0, time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, time.Time{}, fmt.Errorf("incorrect status code: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, time.Time{}, err
	}

	return decodeChangesetState(data)
}

func decodeChangesetState(data []byte) (ChangesetSeqID, time.Time, error) {
	// example
	// ---
	// last_run: 2016-07-02 22:46:01.422137422 Z
	// sequence: 1912325

	lines := bytes.Split(data, []byte("\n"))

	parts := bytes.Split(lines[1], []byte(":"))
	timeString := string(bytes.TrimSpace(bytes.Join(parts[1:], []byte(":"))))
	t, err := time.Parse(
		"2006-01-02 15:04:05.999999999 Z",
		timeString)
	if err != nil {
		return 0, time.Time{}, err
	}

	parts = bytes.Split(lines[2], []byte(":"))
	id, err := strconv.Atoi(string(bytes.TrimSpace(parts[1])))
	if err != nil {
		return 0, time.Time{}, err
	}

	return ChangesetSeqID(id), t, nil
}

// Changesets returns the complete list of changesets in for the given replication sequence.
func Changesets(ctx context.Context, id ChangesetSeqID) (osm.Changesets, error) {
	r, err := changesetReader(ctx, id)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	var changesets []*osm.Changeset
	scanner := osmutil.NewChangesetScanner(ctx, gzReader)
	for scanner.Scan() {
		cs := scanner.Changeset()
		changesets = append(changesets, cs)
	}

	if scanner.Err() != nil {
		return changesets, err
	}

	return changesets, nil
}

// changesetReader will return a ReadCloser with the data from the changeset.
// It will be gzip compressed, so the caller must decompress.
// It is the caller's responsibility to call Close on the Reader when done.
func changesetReader(ctx context.Context, id ChangesetSeqID) (io.ReadCloser, error) {
	resp, err := ctxhttp.Get(ctx, httpClient, changesetURL(id))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("incorrect status code: %v", resp.StatusCode)
	}

	return resp.Body, nil
}

func changesetURL(id ChangesetSeqID) string {
	return fmt.Sprintf("%s/replication/changesets/%03d/%03d/%03d.osm.gz",
		planetHost,
		id/1000000,
		(id%1000000)/1000,
		(id % 1000))
}
