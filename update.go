package osm

import (
	"fmt"
	"sort"
	"time"

	"github.com/paulmach/osm/internal/osmpb"
)

// CommitInfoStart is the start time when we know committed at information.
// Any update.Timestamp >= this date is a committed at time. Anything before
// this date is the element timestamp.
var CommitInfoStart = time.Date(2012, 9, 12, 9, 30, 3, 0, time.UTC)

// An Update is a change to children of a way or relation.
// The child type, id, ref and/or role are the same as the child
// at the given index. Lon/Lat are only updated for nodes.
type Update struct {
	Index   int `xml:"index,attr" json:"index"`
	Version int `xml:"version,attr" json:"version"`

	// Timestamp is the committed at time if time > CommitInfoStart or the
	// element timestamp if before that date.
	Timestamp time.Time `xml:"timestamp,attr" json:"timestamp"`

	ChangesetID ChangesetID `xml:"changeset,attr,omitempty" json:"changeset,omitempty"`
	Lat         float64     `xml:"lat,attr,omitempty" json:"lat,omitempty"`
	Lon         float64     `xml:"lon,attr,omitempty" json:"lon,omitempty"`
	Reverse     bool        `xml:"reverse,attr,omitempty" json:"reverse,omitempty"`
}

// Updates are collections of updates.
type Updates []Update

// UpTo will return the subset of updates taking place upto and on
// the given time.
func (us Updates) UpTo(t time.Time) Updates {
	var result Updates

	for _, u := range us {
		if u.Timestamp.After(t) {
			continue
		}

		result = append(result, u)
	}

	return result
}

// UpdateIndexOutOfRangeError is return when appling an update to an object
// and the update index is out of range.
type UpdateIndexOutOfRangeError struct {
	Index int
}

var _ error = &UpdateIndexOutOfRangeError{}

// Error returns a string representation of the error.
func (e *UpdateIndexOutOfRangeError) Error() string {
	return fmt.Sprintf("osm: index %d is out of range", e.Index)
}

type updatesSortTS Updates

// SortByTimestamp will sort the updates by timestamp in ascending order.
func (us Updates) SortByTimestamp()    { sort.Sort(updatesSortTS(us)) }
func (us updatesSortTS) Len() int      { return len(us) }
func (us updatesSortTS) Swap(i, j int) { us[i], us[j] = us[j], us[i] }
func (us updatesSortTS) Less(i, j int) bool {
	return us[i].Timestamp.Before(us[j].Timestamp)
}

type updatesSortIndex Updates

// SortByIndex will sort the updates by index in ascending order.
func (us Updates) SortByIndex()           { sort.Sort(updatesSortIndex(us)) }
func (us updatesSortIndex) Len() int      { return len(us) }
func (us updatesSortIndex) Swap(i, j int) { us[i], us[j] = us[j], us[i] }
func (us updatesSortIndex) Less(i, j int) bool {
	if us[i].Index != us[j].Index {
		return us[i].Index < us[j].Index
	}

	return us[i].Timestamp.Before(us[j].Timestamp)
}

func marshalUpdates(updates Updates) *osmpb.DenseMembers {
	if len(updates) == 0 {
		return nil
	}

	l := len(updates)
	indexes := make([]int32, l)
	versions := make([]int32, l)
	timestamps := make([]int64, l)
	changesetIDs := make([]int64, l)
	lats := make([]int64, l)
	lons := make([]int64, l)
	revs := make([]int32, l)

	hasLoc := false
	hasRev := false
	for i, u := range updates {
		indexes[i] = int32(u.Index)
		versions[i] = int32(u.Version)
		timestamps[i] = timeToUnix(u.Timestamp)
		changesetIDs[i] = int64(u.ChangesetID)
		if u.Lat != 0 || u.Lon != 0 {
			hasLoc = true
			lats[i] = geoToInt64(u.Lat)
			lons[i] = geoToInt64(u.Lon)
		}

		if u.Reverse {
			hasRev = true
			revs[i] = 1
		}
	}

	result := &osmpb.DenseMembers{
		Indexes:      encodeInt32(indexes),
		Versions:     versions,
		ChangesetIds: encodeInt64(changesetIDs),
		Timestamps:   encodeInt64(timestamps),
	}

	if hasLoc {
		result.Lats = encodeInt64(lats)
		result.Lons = encodeInt64(lons)
	}

	if hasRev {
		result.Orientation = revs
	}

	return result
}

func unmarshalUpdates(encoded *osmpb.DenseMembers) Updates {
	if encoded == nil {
		return nil
	}

	result := make([]Update, len(encoded.Indexes))

	decodeInt32(encoded.Indexes)
	decodeInt64(encoded.ChangesetIds)
	decodeInt64(encoded.Timestamps)

	decodeInt64(encoded.Lats)
	decodeInt64(encoded.Lons)

	for i := range encoded.Indexes {
		result[i] = Update{
			Index:       int(encoded.Indexes[i]),
			Version:     int(encoded.Versions[i]),
			ChangesetID: ChangesetID(encoded.ChangesetIds[i]),
			Timestamp:   unixToTime(encoded.Timestamps[i]),
		}

		if len(encoded.Lats) > i {
			result[i].Lat = float64(encoded.Lats[i]) / locMultiple
			result[i].Lon = float64(encoded.Lons[i]) / locMultiple
		}

		if len(encoded.Orientation) > i && encoded.Orientation[i] > 0 {
			result[i].Reverse = true
		}
	}

	return result
}
