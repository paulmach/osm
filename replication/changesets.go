package replication

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmxml"
)

// ChangesetSeqNum indicates the sequence of the changeset replication found here:
// https://planet.osm.org/replication/changesets/
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
	s, err := ds.fetchChangesetState(ctx, 0)
	if err != nil {
		return 0, nil, err
	}

	return ChangesetSeqNum(s.SeqNum), s, err
}

// ChangesetState returns the state for the given changeset replication.
// There are no state files before 2007990. In that case a 404 error is returned.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func ChangesetState(ctx context.Context, n ChangesetSeqNum) (*State, error) {
	return DefaultDatasource.ChangesetState(ctx, n)
}

// ChangesetState returns the state for the given changeset replication.
// There are no state files before 2007990. In that case a 404 error is returned.
func (ds *Datasource) ChangesetState(ctx context.Context, n ChangesetSeqNum) (*State, error) {
	return ds.fetchChangesetState(ctx, n)
}

func (ds *Datasource) fetchChangesetState(ctx context.Context, n ChangesetSeqNum) (*State, error) {
	var url string
	if n.Uint64() != 0 {
		url = ds.baseChangesetURL(n) + ".state.txt"
	} else {
		url = fmt.Sprintf("%s/replication/%s/state.yaml", ds.baseURL(), n.Dir())
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ds.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &UnexpectedStatusCodeError{
			Code: resp.StatusCode,
			URL:  url,
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s, err := decodeChangesetState(data)
	if err != nil {
		return nil, err
	}

	// starting at 2008004 the changeset sequence number in the state file is
	// one less than the name of the file. This is a consistent mistake.
	// The correctly paired state and data files have the same name. The number
	// in the state file is the one that is off.
	if n == 0 {
		s.SeqNum++
	} else {
		s.SeqNum = uint64(n)
	}

	return s, nil
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

	return changesetDecoder(ctx, r)
}

func changesetDecoder(ctx context.Context, r io.Reader) (osm.Changesets, error) {
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
	url := ds.baseChangesetURL(n) + ".osm.gz"
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

func (ds *Datasource) baseChangesetURL(cn ChangesetSeqNum) string {
	n := cn.Uint64()
	return fmt.Sprintf("%s/replication/%s/%03d/%03d/%03d",
		ds.baseURL(),
		cn.Dir(),
		n/1000000,
		(n%1000000)/1000,
		n%1000)
}
