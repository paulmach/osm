package replication

import (
	"testing"
	"time"
)

func TestDecodeIntervalState(t *testing.T) {
	data := []byte(`#Sat Jul 16 06:28:03 UTC 2016
txnMaxQueried=836441250
sequenceNumber=2010594
timestamp=2016-07-16T06\:28\:02Z
txnReadyList=
txnMax=836441259
txnActiveList=836441203
`)

	state, err := decodeIntervalState(data, "minute")
	if v := MinuteSeqNum(state.SeqNum); v != 2010594 {
		t.Errorf("incorrect id, got %v", v)
	}

	if !state.Timestamp.Equal(time.Date(2016, 7, 16, 6, 28, 2, 0, time.UTC)) {
		t.Errorf("incorrect time, got %v", state.Timestamp)
	}

	if v := state.TxnMax; v != 836441259 {
		t.Errorf("incorrect txnMax, got %v", v)
	}

	if v := state.TxnMaxQueried; v != 836441250 {
		t.Errorf("incorrect txnMaxQueried, got %v", v)
	}

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	// to do some live testing
	// log.Println(CurrentMinuteState(context.Background()))
	// log.Println(MinuteState(context.Background(), 2139244))
	// log.Println(CurrentHourState(context.Background()))
	// log.Println(CurrentDayState(context.Background()))
	// log.Println(Minute(context.Background(), 2010617))
}
