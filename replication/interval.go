package replication

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/paulmach/osm"
)

var _ SeqNum = MinuteSeqNum(0)
var _ SeqNum = HourSeqNum(0)
var _ SeqNum = DaySeqNum(0)

// MinuteSeqStart is the beginning of valid minutely sequence data.
// The few before look to be way more than a minute.
// A quick looks says about 75, 57, 17 for 1, 2, 3 respectively.
const MinuteSeqStart = MinuteSeqNum(4)

// HourSeqStart is the beginning of valid hour sequence data.
// Without deep inspection it looks like 1-10 are from July 2013.
const HourSeqStart = HourSeqNum(11)

// DaySeqStart is the beginning of valid day sequence data.
const DaySeqStart = DaySeqNum(1)

// State returns information about the current replication state.
type State struct {
	SeqNum        uint64    `json:"seq_num"`
	Timestamp     time.Time `json:"timestamp"`
	TxnMax        int       `json:"txn_max,omitempty"`
	TxnMaxQueried int       `json:"txn_max_queries,omitempty"`
}

// SeqNum is an interface type that includes MinuteSeqNum,
// HourSeqNum and DaySeqNum. This is an experiment to implement
// a sum type, a type that can be one of several things only.
type SeqNum interface {
	fmt.Stringer
	Dir() string
	Uint64() uint64
	private()
}

func (n MinuteSeqNum) private() {}
func (n HourSeqNum) private()   {}
func (n DaySeqNum) private()    {}

var _ = SeqNum(MinuteSeqNum(0)).private // for the linters

// MinuteSeqNum indicates the sequence of the minutely diff replication found here:
// http://planet.osm.org/replication/minute
type MinuteSeqNum uint64

// String returns 'minute/%d'.
func (n MinuteSeqNum) String() string {
	return fmt.Sprintf("minute/%d", n)
}

// Dir returns the directory of this data on planet osm.
func (n MinuteSeqNum) Dir() string {
	return "minute"
}

// Uint64 returns the seq num as a uint64 type.
func (n MinuteSeqNum) Uint64() uint64 {
	return uint64(n)
}

// HourSeqNum indicates the sequence of the hourly diff replication found here:
// http://planet.osm.org/replication/hour
type HourSeqNum uint64

// String returns 'hour/%d'.
func (n HourSeqNum) String() string {
	return fmt.Sprintf("hour/%d", n)
}

// Dir returns the directory of this data on planet osm.
func (n HourSeqNum) Dir() string {
	return "hour"
}

// Uint64 returns the seq num as a uint64 type.
func (n HourSeqNum) Uint64() uint64 {
	return uint64(n)
}

// DaySeqNum indicates the sequence of the daily diff replication found here:
// http://planet.osm.org/replication/day
type DaySeqNum uint64

// String returns 'day/%d'.
func (n DaySeqNum) String() string {
	return fmt.Sprintf("day/%d", n)
}

// Dir returns the directory of this data on planet osm.
func (n DaySeqNum) Dir() string {
	return "day"
}

// Uint64 returns the seq num as a uint64 type.
func (n DaySeqNum) Uint64() uint64 {
	return uint64(n)
}

// CurrentMinuteState returns the current state of the minutely replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func CurrentMinuteState(ctx context.Context) (MinuteSeqNum, *State, error) {
	return DefaultDatasource.CurrentMinuteState(ctx)
}

// CurrentMinuteState returns the current state of the minutely replication.
func (ds *Datasource) CurrentMinuteState(ctx context.Context) (MinuteSeqNum, *State, error) {
	s, err := ds.MinuteState(ctx, 0)
	if err != nil {
		return 0, nil, err
	}

	return MinuteSeqNum(s.SeqNum), s, err
}

// MinuteState returns the state of the given minutely replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func MinuteState(ctx context.Context, n MinuteSeqNum) (*State, error) {
	return DefaultDatasource.MinuteState(ctx, n)
}

// MinuteState returns the state of the given minutely replication.
func (ds *Datasource) MinuteState(ctx context.Context, n MinuteSeqNum) (*State, error) {
	return ds.fetchState(ctx, n)
}

// CurrentHourState returns the current state of the hourly replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func CurrentHourState(ctx context.Context) (HourSeqNum, *State, error) {
	return DefaultDatasource.CurrentHourState(ctx)
}

// CurrentHourState returns the current state of the hourly replication.
func (ds *Datasource) CurrentHourState(ctx context.Context) (HourSeqNum, *State, error) {
	s, err := ds.HourState(ctx, 0)
	if err != nil {
		return 0, nil, err
	}

	return HourSeqNum(s.SeqNum), s, err
}

// HourState returns the state of the given hourly replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func HourState(ctx context.Context, n HourSeqNum) (*State, error) {
	return DefaultDatasource.HourState(ctx, n)
}

// HourState returns the state of the given hourly replication.
func (ds *Datasource) HourState(ctx context.Context, n HourSeqNum) (*State, error) {
	return ds.fetchState(ctx, n)
}

// CurrentDayState returns the current state of the daily replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func CurrentDayState(ctx context.Context) (DaySeqNum, *State, error) {
	return DefaultDatasource.CurrentDayState(ctx)
}

// CurrentDayState returns the current state of the daily replication.
func (ds *Datasource) CurrentDayState(ctx context.Context) (DaySeqNum, *State, error) {
	s, err := ds.DayState(ctx, 0)
	if err != nil {
		return 0, nil, err
	}

	return DaySeqNum(s.SeqNum), s, err
}

// DayState returns the state of the given daily replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func DayState(ctx context.Context, n DaySeqNum) (*State, error) {
	return DefaultDatasource.DayState(ctx, n)
}

// DayState returns the state of the given daily replication.
func (ds *Datasource) DayState(ctx context.Context, n DaySeqNum) (*State, error) {
	return ds.fetchState(ctx, n)
}

func (ds *Datasource) fetchState(ctx context.Context, n SeqNum) (*State, error) {
	var url string
	if n.Uint64() != 0 {
		url = ds.baseSeqURL(n) + ".state.txt"
	} else {
		url = fmt.Sprintf("%s/replication/%s/state.txt", ds.baseURL(), n.Dir())
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ds.client().Do(req.WithContext(ctx))
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

	data, err := io.ReadAll(resp.Body)
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

// Minute returns the change diff for a given minute.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Minute(ctx context.Context, n MinuteSeqNum) (*osm.Change, error) {
	return DefaultDatasource.Minute(ctx, n)
}

// Minute returns the change diff for a given minute.
func (ds *Datasource) Minute(ctx context.Context, n MinuteSeqNum) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.changeURL(n))
}

// Hour returns the change diff for a given hour.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Hour(ctx context.Context, n HourSeqNum) (*osm.Change, error) {
	return DefaultDatasource.Hour(ctx, n)
}

// Hour returns the change diff for a given hour.
func (ds *Datasource) Hour(ctx context.Context, n HourSeqNum) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.changeURL(n))
}

// Day returns the change diff for a given day.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Day(ctx context.Context, n DaySeqNum) (*osm.Change, error) {
	return DefaultDatasource.Day(ctx, n)
}

// Day returns the change diff for a given day.
func (ds *Datasource) Day(ctx context.Context, n DaySeqNum) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.changeURL(n))
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

func (ds *Datasource) changeURL(n SeqNum) string {
	return ds.baseSeqURL(n) + ".osc.gz"
}

func (ds *Datasource) baseSeqURL(sn SeqNum) string {
	n := sn.Uint64()
	return fmt.Sprintf("%s/replication/%s/%03d/%03d/%03d",
		ds.baseURL(),
		sn.Dir(),
		n/1000000,
		(n%1000000)/1000,
		n%1000)
}
