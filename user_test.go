package osm

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

func TestUser_UnmarshalXML(t *testing.T) {
	rawXML := []byte(`
	<user id="91499" display_name="pm" account_created="2009-01-13T19:49:59Z">
	  <description>mapper</description>
	  <img href="image url"/>
	  <changesets count="2638"/>
	  <traces count="1"/>
	  <blocks>
	    <received count="5" active="6"/>
	  </blocks>
	  <home lat="37.793" lon="-122.2712" zoom="3"/>
	  <languages>
	    <lang>en-UK</lang>
	    <lang>en</lang>
	  </languages>
	  <messages>
	    <received count="15" unread="3"/>
	    <sent count="7"/>
	  </messages>
	</user>`)
	u := &User{}

	err := xml.Unmarshal(rawXML, &u)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if v := u.ID; v != 91499 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Name; v != "pm" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Description; v != "mapper" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Img.Href; v != "image url" {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Changesets.Count; v != 2638 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Traces.Count; v != 1 {
		t.Errorf("incorrect value: %v", v)
	}

	// home
	if v := u.Home.Lat; v != 37.793 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Home.Lon; v != -122.2712 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Home.Zoom; v != 3 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Languages; !reflect.DeepEqual(v, []string{"en-UK", "en"}) {
		t.Errorf("incorrect value: %v", v)
	}

	// blocks
	if v := u.Blocks.Received.Count; v != 5 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Blocks.Received.Active; v != 6 {
		t.Errorf("incorrect value: %v", v)
	}

	// messages
	if v := u.Messages.Received.Count; v != 15 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Messages.Received.Unread; v != 3 {
		t.Errorf("incorrect value: %v", v)
	}

	if v := u.Messages.Sent.Count; v != 7 {
		t.Errorf("incorrect value: %v", v)
	}

	// created
	if v := u.CreatedAt; !v.Equal(time.Date(2009, 1, 13, 19, 49, 59, 0, time.UTC)) {
		t.Errorf("incorrect value: %v", v)
	}

	// should marshal correctly.
	data, err := xml.Marshal(u)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	nu := &User{}
	err = xml.Unmarshal(data, &nu)
	if err != nil {
		t.Fatalf("xml unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(nu, u) {
		t.Errorf("incorrect marshal")
		t.Log(nu)
		t.Log(u)
	}
}

func TestUser_ObjectID(t *testing.T) {
	u := User{ID: 123}
	id := u.ObjectID()

	if v := id.Type(); v != TypeUser {
		t.Errorf("incorrect type: %v", v)
	}

	if v := id.Ref(); v != 123 {
		t.Errorf("incorrect ref: %v", 123)
	}
}

func TestUser_MarshalJSON(t *testing.T) {
	u := User{
		ID:   123,
		Name: "user",
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !bytes.Equal(data, []byte(`{"type":"user","id":123,"name":"user","img":{"href":""},"changesets":{"count":0},"traces":{"count":0},"home":{"lat":0,"lon":0,"zoom":0},"languages":null,"blocks":{"received":{"count":0,"active":0}},"messages":{"received":{"count":0,"unread":0},"sent":{"count":0}},"created_at":"0001-01-01T00:00:00Z"}`)) {
		t.Errorf("incorrect json: %v", string(data))
	}
}
