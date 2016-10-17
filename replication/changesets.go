package replication

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/paulmach/go.osm"
	"github.com/paulmach/go.osm/osmxml"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// ChangesetSeqNum indicates the sequence of the changeset replication found here:
// http://planet.osm.org/replication/changesets/
type ChangesetSeqNum uint

// CurrentChangesetState returns the current state of the changeset replication.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func CurrentChangesetState(ctx context.Context) (ChangesetSeqNum, State, error) {
	return DefaultDataSource.CurrentChangesetState(ctx)
}

// CurrentChangesetState returns the current state of the changeset replication.
func (ds *DataSource) CurrentChangesetState(ctx context.Context) (ChangesetSeqNum, State, error) {
	url := ds.baseURL() + "/replication/changesets/state.yaml"
	resp, err := ctxhttp.Get(ctx, ds.client(), url)
	if err != nil {
		return 0, State{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, State{}, &UnexpectedStatusCodeError{
			Code: resp.StatusCode,
			URL:  url,
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, State{}, err
	}

	s, err := decodeChangesetState(data)
	return ChangesetSeqNum(s.SeqNum), s, err
}

func decodeChangesetState(data []byte) (State, error) {
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
		return State{}, err
	}

	parts = bytes.Split(lines[2], []byte(":"))
	n, err := strconv.Atoi(string(bytes.TrimSpace(parts[1])))
	if err != nil {
		return State{}, err
	}

	return State{
		SeqNum:    uint(n),
		Timestamp: t,
	}, nil
}

// Changesets returns the complete list of changesets in for the given replication sequence.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func Changesets(ctx context.Context, n ChangesetSeqNum) (osm.Changesets, error) {
	return DefaultDataSource.Changesets(ctx, n)
}

// Changesets returns the complete list of changesets in for the given replication sequence.
func (ds *DataSource) Changesets(ctx context.Context, n ChangesetSeqNum) (osm.Changesets, error) {
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
		e := scanner.Element()
		changesets = append(changesets, e.Changeset)
	}

	return changesets, scanner.Err()
}

// changesetReader will return a ReadCloser with the data from the changeset.
// It will be gzip compressed, so the caller must decompress.
// It is the caller's responsibility to call Close on the Reader when done.
func (ds *DataSource) changesetReader(ctx context.Context, n ChangesetSeqNum) (io.ReadCloser, error) {
	url := ds.changesetURL(n)
	resp, err := ctxhttp.Get(ctx, ds.client(), url)
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

func (ds *DataSource) changesetURL(n ChangesetSeqNum) string {
	return fmt.Sprintf("%s/replication/changesets/%03d/%03d/%03d.osm.gz",
		ds.baseURL(),
		n/1000000,
		(n%1000000)/1000,
		(n % 1000))
}
