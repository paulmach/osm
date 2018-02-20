package osm

import (
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

func TestNoteUnmarshalXML(t *testing.T) {
	rawXML := []byte(`
<note lon="0.0088488" lat="51.5438971">
  <id>1302953</id>
  <url>note url</url>
  <comment_url>comment url</comment_url>
  <close_url>close url</close_url>
  <reopen_url>reopen url</reopen_url>
  <date_created>2018-02-17 17:34:48 UTC</date_created>
  <status>closed</status>
  <date_closed>2018-02-17 22:16:03 UTC</date_closed>
  <comments>
    <comment>
      <date>2018-02-17 17:34:48 UTC</date>
      <uid>251221</uid>
      <user>spiregrain</user>
      <user_url>user url</user_url>
      <action>opened</action>
      <text>comment text</text>
	  <html>comment html</html>
    </comment>
    <comment>
      <date>2018-02-17 22:16:03 UTC</date>
      <uid>251221</uid>
      <user>spiregrain</user>
      <user_url>https://api.openstreetmap.org/user/spiregrain</user_url>
      <action>closed</action>
      <text/>
      <html><p></p></html>
    </comment>
  </comments>
</note>`)
	n := &Note{}

	err := xml.Unmarshal(rawXML, &n)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if v := n.ID; v != 1302953 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Lat; v != 51.5438971 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Lon; v != 0.0088488 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.URL; v != "note url" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.CommentURL; v != "comment url" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.CloseURL; v != "close url" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.ReopenURL; v != "reopen url" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.DateCreated; !v.Equal(time.Date(2018, 2, 17, 17, 34, 48, 0, time.UTC)) {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.DateClosed; !v.Equal(time.Date(2018, 2, 17, 22, 16, 3, 0, time.UTC)) {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Status; v != NoteClosed {
		t.Errorf("incorrect value: %v", v)
	}

	// comments
	if v := len(n.Comments); v != 2 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Comments[0].Date; !v.Equal(time.Date(2018, 2, 17, 17, 34, 48, 0, time.UTC)) {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Comments[0].UserID; v != 251221 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Comments[0].User; v != "spiregrain" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Comments[0].UserURL; v != "user url" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Comments[0].Action; v != NoteCommentOpened {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Comments[0].Text; v != "comment text" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := n.Comments[0].HTML; v != "comment html" {
		t.Errorf("incorrect value: %v", v)
	}

	// should marshal correctly.
	data, err := xml.Marshal(n)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	nn := &Note{}
	err = xml.Unmarshal(data, &nn)
	if err != nil {
		t.Fatalf("xml unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(nn, n) {
		t.Errorf("incorrect marshal")
		t.Log(nn)
		t.Log(n)
	}
}
