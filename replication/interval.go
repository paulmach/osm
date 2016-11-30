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

	osm "github.com/paulmach/go.osm"
)

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
	private()
}

func (n MinuteSeqNum) private()    {}
func (n HourSeqNum) private()      {}
func (n DaySeqNum) private()       {}
func (n ChangesetSeqNum) private() {}

// MinuteSeqNum indicates the sequence of the minutely diff replication found here:
// http://planet.osm.org/replication/minute
type MinuteSeqNum uint64

// HourSeqNum indicates the sequence of the hourly diff replication found here:
// http://planet.osm.org/replication/hour
type HourSeqNum uint64

// DaySeqNum indicates the sequence of the daily diff replication found here:
// http://planet.osm.org/replication/day
type DaySeqNum uint64

// CurrentMinuteState returns the current state of the minutely replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func CurrentMinuteState(ctx context.Context) (MinuteSeqNum, State, error) {
	return DefaultDatasource.CurrentMinuteState(ctx)
}

// CurrentMinuteState returns the current state of the minutely replication.
func (ds *Datasource) CurrentMinuteState(ctx context.Context) (MinuteSeqNum, State, error) {
	s, err := ds.MinuteState(ctx, 0)
	if err != nil {
		return 0, State{}, err
	}

	return MinuteSeqNum(s.SeqNum), s, err
}

// MinuteState returns the state of the given minutely replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func MinuteState(ctx context.Context, n MinuteSeqNum) (State, error) {
	return DefaultDatasource.MinuteState(ctx, n)
}

// MinuteState returns the state of the given minutely replication.
func (ds *Datasource) MinuteState(ctx context.Context, n MinuteSeqNum) (State, error) {
	return ds.fetchIntervalState(ctx, "minute", int(n))
}

// CurrentHourState returns the current state of the hourly replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func CurrentHourState(ctx context.Context) (HourSeqNum, State, error) {
	return DefaultDatasource.CurrentHourState(ctx)
}

// CurrentHourState returns the current state of the hourly replication.
func (ds *Datasource) CurrentHourState(ctx context.Context) (HourSeqNum, State, error) {
	s, err := ds.HourState(ctx, 0)
	if err != nil {
		return 0, State{}, err
	}

	return HourSeqNum(s.SeqNum), s, err
}

// HourState returns the state of the given hourly replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func HourState(ctx context.Context, n HourSeqNum) (State, error) {
	return DefaultDatasource.HourState(ctx, n)
}

// HourState returns the state of the given hourly replication.
func (ds *Datasource) HourState(ctx context.Context, n HourSeqNum) (State, error) {
	return ds.fetchIntervalState(ctx, "hour", int(n))
}

// CurrentDayState returns the current state of the daily replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func CurrentDayState(ctx context.Context) (DaySeqNum, State, error) {
	return DefaultDatasource.CurrentDayState(ctx)
}

// CurrentDayState returns the current state of the daily replication.
func (ds *Datasource) CurrentDayState(ctx context.Context) (DaySeqNum, State, error) {
	s, err := ds.DayState(ctx, 0)
	if err != nil {
		return 0, State{}, err
	}

	return DaySeqNum(s.SeqNum), s, err
}

// DayState returns the state of the given daily replication.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func DayState(ctx context.Context, n DaySeqNum) (State, error) {
	return DefaultDatasource.DayState(ctx, n)
}

// DayState returns the state of the given daily replication.
func (ds *Datasource) DayState(ctx context.Context, n DaySeqNum) (State, error) {
	return ds.fetchIntervalState(ctx, "day", int(n))
}

func (ds *Datasource) fetchIntervalState(ctx context.Context, interval string, n int) (State, error) {
	var url string
	if n != 0 {
		url = ds.baseIntervalURL(interval, n) + ".state.txt"
	} else {
		url = fmt.Sprintf("%s/replication/%s/state.txt", ds.baseURL(), interval)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return State{}, err
	}

	resp, err := ds.Client.Do(req.WithContext(ctx))
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

	return decodeIntervalState(data, interval)
}

func decodeIntervalState(data []byte, interval string) (State, error) {
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

			switch interval {
			case "minute":
				state.SeqNum = uint64(n)
			case "hour":
				state.SeqNum = uint64(n)
			case "day":
				state.SeqNum = uint64(n)
			default:
				panic("unsupported interval")
			}
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
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Minute(ctx context.Context, n MinuteSeqNum) (*osm.Change, error) {
	return DefaultDatasource.Minute(ctx, n)
}

// Minute returns the change diff for a given minute.
func (ds *Datasource) Minute(ctx context.Context, n MinuteSeqNum) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.minuteURL(n))
}

// Hour returns the change diff for a given hour.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Hour(ctx context.Context, n HourSeqNum) (*osm.Change, error) {
	return DefaultDatasource.Hour(ctx, n)
}

// Hour returns the change diff for a given hour.
func (ds *Datasource) Hour(ctx context.Context, n HourSeqNum) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.hourURL(n))
}

// Day returns the change diff for a given day.
// Delegates to the DefaultDatasource and uses its http.Client to make the request.
func Day(ctx context.Context, n DaySeqNum) (*osm.Change, error) {
	return DefaultDatasource.Day(ctx, n)
}

// Day returns the change diff for a given day.
func (ds *Datasource) Day(ctx context.Context, n DaySeqNum) (*osm.Change, error) {
	return ds.fetchIntervalData(ctx, ds.dayURL(n))
}

func (ds *Datasource) fetchIntervalData(ctx context.Context, url string) (*osm.Change, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ds.Client.Do(req.WithContext(ctx))
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

func (ds *Datasource) minuteURL(n MinuteSeqNum) string {
	return ds.dataURL("minute", int(n))
}

func (ds *Datasource) hourURL(n HourSeqNum) string {
	return ds.dataURL("hour", int(n))
}

func (ds *Datasource) dayURL(n DaySeqNum) string {
	return ds.dataURL("day", int(n))
}

func (ds *Datasource) dataURL(interval string, n int) string {
	return ds.baseIntervalURL(interval, n) + ".osc.gz"
}

func (ds *Datasource) baseIntervalURL(interval string, n int) string {
	return fmt.Sprintf("%s/replication/%s/%03d/%03d/%03d",
		ds.baseURL(),
		interval,
		n/1000000,
		(n%1000000)/1000,
		(n % 1000))
}
