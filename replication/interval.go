package replication

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/paulmach/go.osm"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// State returns information about the current replication state.
type State struct {
	SeqNum        uint      `json:"seq_num"`
	Timestamp     time.Time `json:"timestamp"`
	TxnMax        int       `json:"txn_max,omitempty"`
	TxnMaxQueried int       `json:"txn_max_queries,omitempty"`
}

// MinuteSeqNum indicates the sequence of the minutely diff replication found here:
// http://planet.osm.org/replication/minute
type MinuteSeqNum uint

// HourSeqNum indicates the sequence of the hourly diff replication found here:
// http://planet.osm.org/replication/hour
type HourSeqNum uint

// DaySeqNum indicates the sequence of the daily diff replication found here:
// http://planet.osm.org/replication/day
type DaySeqNum uint

// CurrentMinuteState returns the current state of the minutely replication.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func CurrentMinuteState(ctx context.Context) (MinuteSeqNum, State, error) {
	return DefaultDataSource.CurrentMinuteState(ctx)
}

// CurrentMinuteState returns the current state of the minutely replication.
func (ds *DataSource) CurrentMinuteState(ctx context.Context) (MinuteSeqNum, State, error) {
	s, err := ds.MinuteState(ctx, 0)
	return MinuteSeqNum(s.SeqNum), s, err
}

// MinuteState returns the state of the given minutely replication.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func MinuteState(ctx context.Context, n MinuteSeqNum) (State, error) {
	return DefaultDataSource.MinuteState(ctx, n)
}

// MinuteState returns the state of the given minutely replication.
func (ds *DataSource) MinuteState(ctx context.Context, n MinuteSeqNum) (State, error) {
	return ds.fetchIntervalState(ctx, "minute", int(n))
}

// CurrentHourState returns the current state of the hourly replication.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func CurrentHourState(ctx context.Context) (HourSeqNum, State, error) {
	return DefaultDataSource.CurrentHourState(ctx)
}

// CurrentHourState returns the current state of the hourly replication.
func (ds *DataSource) CurrentHourState(ctx context.Context) (HourSeqNum, State, error) {
	s, err := ds.HourState(ctx, 0)
	return HourSeqNum(s.SeqNum), s, err
}

// HourState returns the state of the given hourly replication.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func HourState(ctx context.Context, n HourSeqNum) (State, error) {
	return DefaultDataSource.HourState(ctx, n)
}

// HourState returns the state of the given hourly replication.
func (ds *DataSource) HourState(ctx context.Context, n HourSeqNum) (State, error) {
	return ds.fetchIntervalState(ctx, "hour", int(n))
}

// CurrentDayState returns the current state of the daily replication.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func CurrentDayState(ctx context.Context) (DaySeqNum, State, error) {
	return DefaultDataSource.CurrentDayState(ctx)
}

// CurrentDayState returns the current state of the daily replication.
func (ds *DataSource) CurrentDayState(ctx context.Context) (DaySeqNum, State, error) {
	s, err := ds.DayState(ctx, 0)
	return DaySeqNum(s.SeqNum), s, err
}

// DayState returns the state of the given daily replication.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func DayState(ctx context.Context, n DaySeqNum) (State, error) {
	return DefaultDataSource.DayState(ctx, n)
}

// DayState returns the state of the given daily replication.
func (ds *DataSource) DayState(ctx context.Context, n DaySeqNum) (State, error) {
	return ds.fetchIntervalState(ctx, "day", int(n))
}

func (ds *DataSource) fetchIntervalState(ctx context.Context, interval string, n int) (State, error) {
	var url string
	if n != 0 {
		url = ds.baseIntervalURL(interval, n) + ".state.txt"
	} else {
		url = fmt.Sprintf("%s/replication/%s/state.txt", ds.baseURL(), interval)
	}

	resp, err := ctxhttp.Get(ctx, ds.client(), url)
	if err != nil {
		return State{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return State{}, fmt.Errorf("incorrect status code: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return State{}, err
	}

	return decodeIntervalState(data)
}

func decodeIntervalState(data []byte) (State, error) {
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
		state State
		n     int
		err   error
	)

	for _, l := range bytes.Split(data, []byte("\n")) {
		parts := bytes.Split(l, []byte("="))

		if bytes.Equal(parts[0], []byte("sequenceNumber")) {
			n, err = strconv.Atoi(string(bytes.TrimSpace(parts[1])))
			if err != nil {
				return State{}, err
			}
			state.SeqNum = uint(n)
		} else if bytes.Equal(parts[0], []byte("txnMax")) {
			state.TxnMax, err = strconv.Atoi(string(bytes.TrimSpace(parts[1])))
			if err != nil {
				return State{}, err
			}
		} else if bytes.Equal(parts[0], []byte("txnMaxQueried")) {
			state.TxnMaxQueried, err = strconv.Atoi(string(bytes.TrimSpace(parts[1])))
			if err != nil {
				return State{}, err
			}
		} else if bytes.Equal(parts[0], []byte("timestamp")) {
			timeString := string(bytes.TrimSpace(parts[1]))
			state.Timestamp, err = time.Parse(
				"2006-01-02T15\\:04\\:05Z",
				timeString)
			if err != nil {
				return State{}, err
			}
		}
	}

	return state, nil
}

// Minute returns the change diff for a given minute.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func Minute(ctx context.Context, n MinuteSeqNum) (*osm.Change, error) {
	return DefaultDataSource.Minute(ctx, n)
}

// Minute returns the change diff for a given minute.
func (ds *DataSource) Minute(ctx context.Context, n MinuteSeqNum) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.minuteURL(n))
}

// Hour returns the change diff for a given hour.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func Hour(ctx context.Context, n HourSeqNum) (*osm.Change, error) {
	return DefaultDataSource.Hour(ctx, n)
}

// Hour returns the change diff for a given hour.
func (ds *DataSource) Hour(ctx context.Context, n HourSeqNum) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.hourURL(n))
}

// Day returns the change diff for a given day.
// Delegates to the DefaultDataSource and uses its http.Client to make the request.
func Day(ctx context.Context, n DaySeqNum) (*osm.Change, error) {
	return DefaultDataSource.Day(ctx, n)
}

// Day returns the change diff for a given day.
func (ds *DataSource) Day(ctx context.Context, n DaySeqNum) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.dayURL(n))
}

func (ds *DataSource) fetchIntervalData(ctx context.Context, url string) (*osm.Change, error) {
	resp, err := ctxhttp.Get(ctx, ds.client(), url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("incorrect status code: %v", resp.StatusCode)
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

func (ds *DataSource) minuteURL(n MinuteSeqNum) string {
	return ds.dataURL("minute", int(n))
}

func (ds *DataSource) hourURL(n HourSeqNum) string {
	return ds.dataURL("hour", int(n))
}

func (ds *DataSource) dayURL(n DaySeqNum) string {
	return ds.dataURL("day", int(n))
}

func (ds *DataSource) dataURL(interval string, n int) string {
	return ds.baseIntervalURL(interval, n) + ".osc.gz"
}

func (ds *DataSource) baseIntervalURL(interval string, n int) string {
	return fmt.Sprintf("%s/replication/%s/%03d/%03d/%03d",
		ds.baseURL(),
		interval,
		n/1000000,
		(n%1000000)/1000,
		(n % 1000))
}
