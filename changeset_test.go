package osm

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

func TestChangesets(t *testing.T) {
	data := []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6" generator="replicate_changesets.rb" copyright="OpenStreetMap and contributors" attribution="http://www.openstreetmap.org/copyright" license="http://opendatacommons.org/licenses/odbl/1-0/">
  <changeset id="36947117" created_at="2016-02-01T21:57:17Z" closed_at="2016-02-01T23:05:55Z" open="true" num_changes="86" user="padvinder" uid="978786" min_lat="52.7016394" max_lat="52.7236643" min_lon="5.1545597" max_lon="5.2532961" comments_count="5">
    <tag k="build" v="2.4-16-g0c126d0"/>
    <tag k="created_by" v="Potlatch 2"/>
    <tag k="version" v="2.4"/>
  </changeset>
  <changeset id="36947173" created_at="2016-02-01T22:00:56Z" closed_at="2016-02-01T23:05:06Z" open="false" num_changes="9" user="florijn11" uid="1319603" min_lat="51.5871887" max_lat="51.6032569" min_lon="5.3214071" max_lon="5.33106" comments_count="0">
    <tag k="version" v="2.4"/>
    <tag k="build" v="2.4-16-g0c126d0"/>
    <tag k="comment" v="Fietsdoorsteek aangepast"/>
    <tag k="created_by" v="Potlatch 2"/>
  </changeset>
</osm>`)

	cs := &OSM{}
	err := xml.Unmarshal(data, &cs)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if l := len(cs.Changesets); l != 2 {
		t.Fatalf("incorrect number of changesets, got %v", l)
	}

	c := cs.Changesets[0]
	if v := c.ID; v != 36947117 {
		t.Errorf("incorrect id, got %v", v)
	}

	if v := c.CreatedAt; v != time.Date(2016, time.February, 1, 21, 57, 17, 0, time.UTC) {
		t.Errorf("incorrect created at, got %v", v)
	}

	if v := c.ClosedAt; v != time.Date(2016, time.February, 1, 23, 05, 55, 0, time.UTC) {
		t.Errorf("incorrect closed at, got %v", v)
	}

	if v := c.ChangesCount; v != 86 {
		t.Errorf("incorrect changes count, got %v", v)
	}

	if v := c.User; v != "padvinder" {
		t.Errorf("incorrect user, got %v", v)
	}

	if v := c.UserID; v != 978786 {
		t.Errorf("incorrect user id, got %v", v)
	}

	if v := c.MinLat; v != 52.7016394 {
		t.Errorf("incorrect min lat, got %v", v)
	}

	if v := c.MaxLat; v != 52.7236643 {
		t.Errorf("incorrect max lat, got %v", v)
	}

	if v := c.MinLon; v != 5.1545597 {
		t.Errorf("incorrect min lon, got %v", v)
	}

	if v := c.MaxLon; v != 5.2532961 {
		t.Errorf("incorrect max on, got %v", v)
	}

	if v := c.CommentsCount; v != 5 {
		t.Errorf("incorrect comment count, got %v", v)
	}
}

func TestChangeset(t *testing.T) {
	data := []byte(`
<changeset id="38162206" user="grah735" uid="2744209" created_at="2016-03-30T09:25:31Z" closed_at="2016-03-30T09:25:36Z" open="false" min_lat="44.5540891" min_lon="33.5261473" max_lat="44.5614501" max_lon="33.5302043" comments_count="0">
  <tag k="comment" v="двойная сплошная"/>
  <tag k="locale" v="ru"/>
  <tag k="host" v="https://www.openstreetmap.org/id"/>
  <tag k="imagery_used" v="Bing"/>
  <tag k="created_by" v="iD 1.9.2"/>
</changeset>`)
	c := loadChange(t, "testdata/changeset_38162206.osc")

	cs1 := &Changeset{}
	err := xml.Unmarshal(data, cs1)
	if err != nil {
		t.Fatalf("unable to unmarshal changeset: %v", err)
	}

	cs1.XMLName = xml.Name{}
	cs1.Change = c

	data, err = cs1.Marshal()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	cs2, err := UnmarshalChangeset(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(cs1, cs2) {
		t.Errorf("changesets are not equal")
		t.Logf("%+v", cs1)
		t.Logf("%+v", cs2)
	}

	// empty change
	cs3 := &Changeset{
		Change: &Change{},
	}
	data, err = cs3.Marshal()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if l := len(data); l != 0 {
		t.Errorf("empty should be empty, got %v", l)
	}
}

func TestChangesetOpen(t *testing.T) {
	data := []byte(`
<changeset id="40309372" user="Bahntech" uid="3619264" created_at="2016-06-26T21:26:41Z" open="true" min_lat="51.484563" min_lon="12.0995042" max_lat="51.484563" max_lon="12.0995042" comments_count="0">
	<tag k="comment" v="updated fire hydrant details with OsmHydrant"/>
	<tag k="created_by" v="OsmHydrant / http://yapafo.net v0.3"/>
</changeset>`)

	var c Changeset
	err := xml.Unmarshal(data, &c)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !c.ClosedAt.IsZero() {
		t.Errorf("closed at should be zero for open changesets")
	}
}

func TestChangesetTags(t *testing.T) {
	data := []byte(`
<changeset id="123123">
  <tag k="comment" v="changeset comment"/>
  <tag k="created_by" v="iD 1.8.3"/>
  <tag k="locale" v="en-US"/>
  <tag k="host" v="http://id.org"/>
  <tag k="imagery_used" v="Bing"/>
  <tag k="source" v="some data"/>
  <tag k="bot" v="yes"/>
</changeset>`)

	var c Changeset
	err := xml.Unmarshal(data, &c)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if v := c.Comment(); v != "changeset comment" {
		t.Errorf("incorrect comment, got %v", v)
	}

	if v := c.CreatedBy(); v != "iD 1.8.3" {
		t.Errorf("incorrect created by, got %v", v)
	}

	if v := c.Locale(); v != "en-US" {
		t.Errorf("incorrect locale, got %v", v)
	}

	if v := c.Host(); v != "http://id.org" {
		t.Errorf("incorrect host, got %v", v)
	}

	if v := c.ImageryUsed(); v != "Bing" {
		t.Errorf("incorrect imagery used, got %v", v)
	}

	if v := c.Source(); v != "some data" {
		t.Errorf("incorrect source, got %v", v)
	}

	if v := c.Bot(); v != true {
		t.Errorf("incorrect bot, got %v", v)
	}
}

func TestChangesetBound(t *testing.T) {
	data := []byte(`
<changeset id="36947173" created_at="2016-02-01T22:00:56Z" closed_at="2016-02-01T23:05:06Z" open="false" num_changes="9" user="florijn11" uid="1319603" min_lat="51.5871887" max_lat="51.6032569" min_lon="5.3214071" max_lon="5.33106" comments_count="0">
    <tag k="version" v="2.4"/>
    <tag k="build" v="2.4-16-g0c126d0"/>
    <tag k="comment" v="Fietsdoorsteek aangepast"/>
    <tag k="created_by" v="Potlatch 2"/>
</changeset>`)

	var c Changeset
	err := xml.Unmarshal(data, &c)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
}

func TestChangesetComment(t *testing.T) {
	data := []byte(`
<changeset id="40303151" user="Glen Bundrick" uid="4173877" created_at="2016-06-26T15:37:47Z" closed_at="2016-06-26T15:37:48Z" open="false" min_lat="34.6591676" min_lon="-81.8789825" max_lat="34.6594167" max_lon="-81.8788142" comments_count="3">
  <tag k="comment" v="Recent Doublewide addition"/>
  <tag k="locale" v="en-US"/>
  <tag k="host" v="https://www.openstreetmap.org/id"/>
  <tag k="imagery_used" v="Bing"/>
  <tag k="created_by" v="iD 1.9.6"/>
  <discussion>
    <comment date="2016-06-26T17:22:27Z" uid="5359" user="user_5359">
      <text>Welcome to OSM!</text>
    </comment>
    <comment date="2016-06-26T20:56:11Z" uid="4173877" user="Glen Bundrick">
      <text>OK New to this and learning
</text>
    </comment>
    <comment date="2016-06-26T21:37:40Z" uid="5359" user="user_5359">
      <text>No problem. If you need help, you are welcome!</text>
    </comment>
  </discussion>
</changeset>`)

	var c Changeset
	err := xml.Unmarshal(data, &c)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if l := len(c.Discussion.Comments); l != 3 {
		t.Errorf("incorrect number of comments, got %v", l)
	}

	com := c.Discussion.Comments[0]
	if v := com.CreatedAt; v != time.Date(2016, 6, 26, 17, 22, 27, 0, time.UTC) {
		t.Errorf("incorrect created at, got %v", v)
	}

	if v := com.User; v != "user_5359" {
		t.Errorf("incorrect user, got %v", v)
	}

	if v := com.UserID; v != 5359 {
		t.Errorf("incorrect username, got %v", v)
	}

	if v := com.Text; v != "Welcome to OSM!" {
		t.Errorf("incorrect text, got %v", v)
	}
}

func TestChangesetMarshalXML(t *testing.T) {
	cs := Changeset{
		ID: 123,
	}

	data, err := xml.Marshal(cs)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	expected := `<changeset id="123" user="" uid="0" created_at="0001-01-01T00:00:00Z" closed_at="0001-01-01T00:00:00Z" open="false" min_lat="0" max_lat="0" min_lon="0" max_lon="0"></changeset>`
	if !bytes.Equal(data, []byte(expected)) {
		t.Errorf("incorrect marshal, got: %s", string(data))
	}

	// changeset with discussion
	cs.Discussion = ChangesetDiscussion{
		Comments: []*ChangesetComment{
			&ChangesetComment{Text: "foo"},
		},
	}

	data, err = xml.Marshal(cs)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	expected = `<changeset id="123" user="" uid="0" created_at="0001-01-01T00:00:00Z" closed_at="0001-01-01T00:00:00Z" open="false" min_lat="0" max_lat="0" min_lon="0" max_lon="0"><discussion><comment user="" uid="0" date="0001-01-01T00:00:00Z"><text>foo</text></comment></discussion></changeset>`
	if !bytes.Equal(data, []byte(expected)) {
		t.Errorf("incorrect marshal, got: %s", string(data))
	}
}
