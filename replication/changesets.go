package replication

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmxml"
)

// ChangesetSeqNum indicates the sequence of the changeset replication found here:
// http://planet.osm.org/replication/changesets/
type ChangesetSeqNum uint64

// String returns 'changeset/%d'.
func (n ChangesetSeqNum) String() string {
	return fmt.Sprintf("changeset/%d", n)
}

// Dir returns the directory of this data on planet osm.
func (n ChangesetSeqNum) Dir() string {
	return "changesets"
}

// Uint64 returns the seq num as a uint64 type.
func (n ChangesetSeqNum) Uint64() uint64 {
	return uint64(n)
}

// CurrentChangesetState returns the current state of the changeset replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func CurrentChangesetState(ctx context.Context) (ChangesetSeqNum, *State, error) {
	return DefaultDatasource.CurrentChangesetState(ctx)
}

// CurrentChangesetState returns the current state of the changeset replication.
func (ds *Datasource) CurrentChangesetState(ctx context.Context) (ChangesetSeqNum, *State, error) {
	url := ds.baseURL() + "/replication/changesets/state.yaml"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, err
	}

	resp, err := ds.client().Do(req.WithContext(ctx))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, nil, &UnexpectedStatusCodeError{
			Code: resp.StatusCode,
			URL:  url,
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	s, err := decodeChangesetState(data)
	return ChangesetSeqNum(s.SeqNum), s, err
}

func decodeChangesetState(data []byte) (*State, error) {
	// example
	// ---
	// last_run: 2016-07-02 22:46:01.422137422 +00:00  (or Z)
	// sequence: 1912325

	lines := bytes.Split(data, []byte("\n"))

	parts := bytes.Split(lines[1], []byte(":"))
	timeString := string(bytes.TrimSpace(bytes.Join(parts[1:], []byte(":"))))

	t, err := decodeTime(timeString)
	if err != nil {
		return nil, err
	}

	parts = bytes.Split(lines[2], []byte(":"))
	n, err := strconv.ParseUint(string(bytes.TrimSpace(parts[1])), 10, 64)
	if err != nil {
		return nil, err
	}

	return &State{
		SeqNum:    n,
		Timestamp: t,
	}, nil
}

// Changesets returns the complete list of changesets for the given replication sequence.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Changesets(ctx context.Context, n ChangesetSeqNum) (osm.Changesets, error) {
	return DefaultDatasource.Changesets(ctx, n)
}

// Changesets returns the complete list of changesets in for the given replication sequence.
func (ds *Datasource) Changesets(ctx context.Context, n ChangesetSeqNum) (osm.Changesets, error) {
	r, err := ds.changesetReader(ctx, n)
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
	scanner := osmxml.New(ctx, gzReader)
	for scanner.Scan() {
		o := scanner.Object()
		c, ok := o.(*osm.Changeset)
		if !ok {
			return nil, fmt.Errorf("osm replication: object not a changeset: %[1]T: %[1]v", o)
		}
		changesets = append(changesets, c)
	}

	return changesets, scanner.Err()
}

// changesetReader will return a ReadCloser with the data from the changeset.
// It will be gzip compressed, so the caller must decompress.
// It is the caller's responsibility to call Close on the Reader when done.
func (ds *Datasource) changesetReader(ctx context.Context, n ChangesetSeqNum) (io.ReadCloser, error) {
	url := ds.changesetURL(n)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ds.client().Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, &UnexpectedStatusCodeError{
			Code: resp.StatusCode,
			URL:  url,
		}
	}

	return resp.Body, nil
}

func (ds *Datasource) changesetURL(n ChangesetSeqNum) string {
	return fmt.Sprintf("%s/replication/changesets/%03d/%03d/%03d.osm.gz",
		ds.baseURL(),
		n/1000000,
		(n%1000000)/1000,
		n % 1000)
}
