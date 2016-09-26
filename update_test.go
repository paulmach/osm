package osm

import (
	"testing"
	"time"
)

func TestUpdatesUpTo(t *testing.T) {
	us := Updates{
		{Index: 1, Timestamp: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Index: 2, Timestamp: time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Index: 3, Timestamp: time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	if v := len(us.UpTo(time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC))); v != 0 {
		t.Errorf("incorrect number of updates, got %v", v)
	}

	u := us.UpTo(time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC))
	if v := len(u); v != 1 {
		t.Errorf("incorrect number of updates, got %v", v)
	}

	if v := u[0].Index; v != 1 {
		t.Errorf("incorrect value, got index: %v", v)
	}

	u = us.UpTo(time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC))
	if v := len(u); v != 2 {
		t.Errorf("incorrect number of updates, got %v", v)
	}

	if v := u[0].Index; v != 1 {
		t.Errorf("incorrect value, got index: %v", v)
	}

	if v := u[1].Index; v != 3 {
		t.Errorf("incorrect value, got index: %v", v)
	}

	if v := len(us.UpTo(time.Date(2013, 2, 1, 0, 0, 0, 0, time.UTC))); v != 2 {
		t.Errorf("incorrect number of updates, got %v", v)
	}

	if v := len(us.UpTo(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC))); v != 3 {
		t.Errorf("incorrect number of updates, got %v", v)
	}
}
