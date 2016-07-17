package replication

import (
	"testing"
	"time"
)

func TestDecodeIntervalState(t *testing.T) {
	data := []byte(`#Sat Jul 16 06:28:03 UTC 2016
txnMaxQueried=836441259
sequenceNumber=2010594
timestamp=2016-07-16T06\:28\:02Z
txnReadyList=
txnMax=836441259
txnActiveList=836441203
`)

	id, updated, err := decodeIntervalState(data)
	if id != 2010594 {
		t.Errorf("incorrect id, got %v", id)
	}

	if !updated.Equal(time.Date(2016, 7, 16, 6, 28, 2, 0, time.UTC)) {
		t.Errorf("incorrect time, got %v", updated)
	}

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	// to do some live testing
	// log.Println(MinuteState(context.Background()))
	// log.Println(HourState(context.Background()))
	// log.Println(DayState(context.Background()))
	// log.Println(Minute(context.Background(), 2010617))
}
