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

// MinuteSeqID indicates the sequence of the minutely diff replication found here:
// http://planet.osm.org/replication/minute/
type MinuteSeqID uint

// HourSeqID indicates the sequence of the hourly diff replication found here:
// http://planet.osm.org/replication/hour/
type HourSeqID uint

// DaySeqID indicates the sequence of the daily diff replication found here:
// http://planet.osm.org/replication/day/
type DaySeqID uint

// MinuteState returns the current state of the minutely replication.
func MinuteState(ctx context.Context) (MinuteSeqID, time.Time, error) {
	id, t, err := fetchIntervalState(ctx, "minute")
	return MinuteSeqID(id), t, err
}

// HourState returns the current state of the hourly replication.
func HourState(ctx context.Context) (HourSeqID, time.Time, error) {
	id, t, err := fetchIntervalState(ctx, "hour")
	return HourSeqID(id), t, err
}

// DayState returns the current state of the daily replication.
func DayState(ctx context.Context) (DaySeqID, time.Time, error) {
	id, t, err := fetchIntervalState(ctx, "day")
	return DaySeqID(id), t, err
}

func fetchIntervalState(ctx context.Context, interval string) (int, time.Time, error) {
	resp, err := ctxhttp.Get(
		ctx,
		httpClient,
		fmt.Sprintf("%s/replication/%s/state.txt", planetHost, interval),
	)
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

	return decodeIntervalState(data)
}

func decodeIntervalState(data []byte) (int, time.Time, error) {
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
		timestamp time.Time
		number    int
		err       error
	)

	for _, l := range bytes.Split(data, []byte("\n")) {
		parts := bytes.Split(l, []byte("="))

		if bytes.Equal(parts[0], []byte("sequenceNumber")) {
			number, err = strconv.Atoi(string(bytes.TrimSpace(parts[1])))
			if err != nil {
				return 0, time.Time{}, err
			}
		}

		if bytes.Equal(parts[0], []byte("timestamp")) {
			timeString := string(bytes.TrimSpace(parts[1]))
			timestamp, err = time.Parse(
				"2006-01-02T15\\:04\\:05Z",
				timeString)
			if err != nil {
				return 0, time.Time{}, err
			}
		}
	}

	return number, timestamp, nil
}

// Minute returns the change diff for a given minute.
func Minute(ctx context.Context, id MinuteSeqID) (*osm.Change, error) {
	return fetchIntervalData(ctx, minuteURL(id))
}

// Hour returns the change diff for a given hour.
func Hour(ctx context.Context, id HourSeqID) (*osm.Change, error) {
	return fetchIntervalData(ctx, hourURL(id))
}

// Day returns the change diff for a given day.
func Day(ctx context.Context, id DaySeqID) (*osm.Change, error) {
	return fetchIntervalData(ctx, dayURL(id))
}

func fetchIntervalData(ctx context.Context, url string) (*osm.Change, error) {
	resp, err := ctxhttp.Get(ctx, httpClient, url)
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

func minuteURL(id MinuteSeqID) string {
	return intervalURL("minute", int(id))
}

func hourURL(id HourSeqID) string {
	return intervalURL("hour", int(id))
}

func dayURL(id DaySeqID) string {
	return intervalURL("day", int(id))
}

func intervalURL(interval string, id int) string {
	return fmt.Sprintf("%s/replication/%s/%03d/%03d/%03d.osc.gz",
		planetHost,
		interval,
		id/1000000,
		(id%1000000)/1000,
		(id % 1000))
}
