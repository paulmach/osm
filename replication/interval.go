package replication

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/onXmaps/osm"
)

// State returns information about the current replication state.
type State struct {
	SeqNum        uint64    `json:"seq_num"`
	Timestamp     time.Time `json:"timestamp"`
	TxnMax        int       `json:"txn_max,omitempty"`
	TxnMaxQueried int       `json:"txn_max_queries,omitempty"`
}

// CurrentState returns the current state of the replication.
func (ds *Datasource) CurrentState(ctx context.Context) (uint64, *State, error) {
	s, err := ds.State(ctx, 0)
	if err != nil {
		return 0, nil, err
	}

	return s.SeqNum, s, err
}

// State returns the state of the given replication.
func (ds *Datasource) State(ctx context.Context, seqNum uint64) (*State, error) {
	return ds.fetchState(ctx, seqNum)
}

func (ds *Datasource) fetchState(ctx context.Context, seqNum uint64) (*State, error) {
	var url string
	if seqNum != 0 {
		url = ds.baseSeqURL(seqNum) + ".state.txt"
	} else {
		url = fmt.Sprintf("%s/state.txt", ds.baseURL())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ds.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, &UnexpectedStatusCodeError{
			Code: resp.StatusCode,
			URL:  url,
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return decodeIntervalState(data)
}

func decodeIntervalState(data []byte) (*State, error) {
	// example
	// ---
	// #Sat Jul 16 06:14:03 UTC 2016
	// txnMaxQueried=836439235
	// sequenceNumber=2010580
	// timestamp=2016-07-16T06\:14\:02Z
	// txnReadyList=
	// txnMax=836439235
	// txnActiveList=836439008

	var (
		n   int
		err error
	)

	state := &State{}
	for _, l := range bytes.Split(data, []byte("\n")) {
		parts := bytes.Split(l, []byte("="))

		if bytes.Equal(parts[0], []byte("sequenceNumber")) {
			n, err = strconv.Atoi(string(bytes.TrimSpace(parts[1])))
			if err != nil {
				return nil, err
			}

			state.SeqNum = uint64(n)
		} else if bytes.Equal(parts[0], []byte("txnMax")) {
			state.TxnMax, err = strconv.Atoi(string(bytes.TrimSpace(parts[1])))
			if err != nil {
				return nil, err
			}
		} else if bytes.Equal(parts[0], []byte("txnMaxQueried")) {
			state.TxnMaxQueried, err = strconv.Atoi(string(bytes.TrimSpace(parts[1])))
			if err != nil {
				return nil, err
			}
		} else if bytes.Equal(parts[0], []byte("timestamp")) {
			timeString := string(bytes.TrimSpace(parts[1]))
			state.Timestamp, err = decodeTime(timeString)
			if err != nil {
				return nil, err
			}
		}
	}

	return state, nil
}

func (ds *Datasource) GetDiff(ctx context.Context, seqNum uint64) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.changeURL(seqNum))
}

func (ds *Datasource) fetchIntervalData(ctx context.Context, url string) (*osm.Change, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ds.Client.Do(req)
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

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	change := &osm.Change{}
	err = xml.NewDecoder(gzReader).Decode(change)
	return change, err
}

func (ds *Datasource) changeURL(seqNum uint64) string {
	return ds.baseSeqURL(seqNum) + ".osc.gz"
}

func (ds *Datasource) baseSeqURL(seqNum uint64) string {
	return fmt.Sprintf("%s/%03d/%03d/%03d",
		ds.baseURL(),
		seqNum/1000000,
		(seqNum%1000000)/1000,
		seqNum%1000)
}
