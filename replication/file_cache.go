package replication

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/paulmach/osm"
)

// FileCache will will first check the file cache before fetching the data
// from the source.
type FileCache struct {
	Dir      string
	Fallback *Datasource
}

// NewFileCache will return a new file cache for replication data.
// Currently only implements the Minute, Hour and Day change data.
func NewFileCache(dir string, fallback *Datasource) *FileCache {
	return &FileCache{
		Dir:      dir,
		Fallback: fallback,
	}
}

// Minute will fetch minute replication changes from the file cache before going to the source.
func (fc *FileCache) Minute(ctx context.Context, n MinuteSeqNum) (*osm.Change, error) {
	return fc.getReplication(ctx, n)
}

// Hour will fetch hour replication changes from the file cache before going to the source.
func (fc *FileCache) Hour(ctx context.Context, n HourSeqNum) (*osm.Change, error) {
	return fc.getReplication(ctx, n)
}

// Day will fetch day replication changes from the file cache before going to the source.
func (fc *FileCache) Day(ctx context.Context, n DaySeqNum) (*osm.Change, error) {
	return fc.getReplication(ctx, n)
}

func (fc *FileCache) getReplication(ctx context.Context, sn SeqNum) (*osm.Change, error) {
	p, n := fc.fileName("replication", sn)
	filename := path.Join(p, n)

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return fc.fetchAndCacheReplication(ctx, sn)
	} else if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return osm.UnmarshalChange(data)
}

func (fc *FileCache) fetchAndCacheReplication(ctx context.Context, sn SeqNum) (*osm.Change, error) {
	var c *osm.Change
	var err error

	switch t := sn.(type) {
	case MinuteSeqNum:
		c, err = fc.Fallback.Minute(ctx, t)
	case HourSeqNum:
		c, err = fc.Fallback.Hour(ctx, t)
	case DaySeqNum:
		c, err = fc.Fallback.Day(ctx, t)
	default:
		panic(fmt.Sprintf("unsupported type %T", sn))
	}

	if err != nil {
		return nil, err
	}

	data, err := c.Marshal()
	if err != nil {
		return nil, err
	}

	p, n := fc.fileName("replication", sn)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		err := os.MkdirAll(p, 0755)
		if err != nil {
			return nil, err
		}
	}

	err = ioutil.WriteFile(path.Join(p, n), data, 0644)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (fc *FileCache) fileName(t string, sn SeqNum) (string, string) {
	n := sn.Uint64()

	n1 := fmt.Sprintf("%03d", n/1000000)
	n2 := fmt.Sprintf("%03d", (n%1000000)/1000)
	n3 := fmt.Sprintf("%03d", (n % 1000))

	return path.Join(fc.Dir, t, sn.Dir(), n1, n2, n3), fmt.Sprintf("%d.dat", n)
}
