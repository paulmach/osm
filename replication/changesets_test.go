package replication

import (
	"bytes"
	"compress/gzip"
	"context"
	"testing"
	"time"
)

func TestDecodeChangesetState(t *testing.T) {
	data := []byte(`---
last_run: 2016-07-02 22:46:01.422137422 Z
sequence: 1912325
`)

	state, err := decodeChangesetState(data)
	if v := ChangesetSeqNum(state.SeqNum); v != 1912325 {
		t.Errorf("incorrect sequence number, got %v", v)
	}

	if !state.Timestamp.Equal(time.Date(2016, 7, 2, 22, 46, 1, 422137422, time.UTC)) {
		t.Errorf("incorrect time, got %v", state.Timestamp)
	}

	if err != nil {
		t.Errorf("got error: %v", err)
	}
}

func TestChangesetDecoder(t *testing.T) {
	ctx := context.Background()

	buf := bytes.NewBuffer(nil)
	w := gzip.NewWriter(buf)
	w.Write([]byte(`
<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6" generator="replicate_changesets.rb">
  <changeset id="41976776" created_at="2016-09-07T11:11:04Z" closed_at="2016-09-07T11:11:19Z" open="false" num_changes="3" user="ابو عمار ياسر" uid="4537049" min_lat="15.3098203" max_lat="15.3316814" min_lon="44.2181132" max_lon="44.2335361" comments_count="0">
    <tag k="created_by" v="MAPS.ME android 6.3.7-Google"/>
    <tag k="comment" v="Created a clinic, a government office, and a supermarket shop"/>
    <tag k="bundle_id" v="com.mapswithme.maps.pro"/>
  </changeset>
  <changeset id="41976777" created_at="2016-09-07T11:11:17Z" closed_at="2016-09-07T11:11:18Z" open="false" num_changes="45" user="blubber75" uid="4522866" min_lat="49.7560493" max_lat="49.756761" min_lon="8.6103231" max_lon="8.6120139" comments_count="0">
    <tag k="comment" v="gebäude"/>
    <tag k="locale" v="de"/>
    <tag k="host" v="https://www.openstreetmap.org/id"/>
    <tag k="imagery_used" v="Bing"/>
    <tag k="created_by" v="iD 1.9.7"/>
  </changeset>
</osm>`))
	w.Close()

	changesets, err := changesetDecoder(ctx, buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if len(changesets) != 2 {
		t.Errorf("incorrect number of changes: %d", len(changesets))
	}
}

func TestBaseChangesetURL(t *testing.T) {
	url := DefaultDatasource.baseChangesetURL(123456789)
	if url != "https://planet.osm.org/replication/changesets/123/456/789" {
		t.Errorf("incorrect url, got %v", url)
	}
}
