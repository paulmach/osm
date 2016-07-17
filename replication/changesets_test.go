package replication

import (
	"testing"
	"time"
)

func TestDecodeChangesetState(t *testing.T) {
	data := []byte(`---
last_run: 2016-07-02 22:46:01.422137422 Z
sequence: 1912325
`)

	id, updated, err := decodeChangesetState(data)
	if id != 1912325 {
		t.Errorf("incorrect id, got %v", id)
	}

	if !updated.Equal(time.Date(2016, 7, 2, 22, 46, 1, 422137422, time.UTC)) {
		t.Errorf("incorrect time, got %v", updated)
	}

	if err != nil {
		t.Errorf("got error: %v", err)
	}
}
