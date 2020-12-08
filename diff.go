package osm

import "encoding/xml"

// Diff represents a difference of osm data with old and new data.
type Diff struct {
	XMLName    xml.Name   `xml:"osm"`
	Actions    Actions    `xml:"action"`
	Changesets Changesets `xml:"changeset"`
}

// Actions is a set of diff actions.
type Actions []Action

// Action is an explicit create, modify or delete action with
// old and new data if applicable. Different properties of this
// struct will be populated depending on the action.
//	Create: da.OSM will contain the new element
//	Modify: da.Old and da.New will contain the old and new elements.
//	Delete: da.Old and da.New will contain the old and new elements.
type Action struct {
	Type ActionType `xml:"type,attr"`
	*OSM `xml:",omitempty"`
	Old  *OSM `xml:"old,omitempty"`
	New  *OSM `xml:"new,omitempty"`
}

// UnmarshalXML converts xml into a diff action.
func (a *Action) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		if attr.Name.Local == "type" {
			a.Type = ActionType(attr.Value)
			break
		}
	}

	for {
		token, err := d.Token()
		if err != nil {
			break
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		switch start.Name.Local {
		case "old":
			a.Old = &OSM{}
			if err := d.DecodeElement(a.Old, &start); err != nil {
				return err
			}
		case "new":
			a.New = &OSM{}
			if err := d.DecodeElement(a.New, &start); err != nil {
				return err
			}
		case "node":
			n := &Node{}
			if err := d.DecodeElement(&n, &start); err != nil {
				return err
			}
			a.OSM = &OSM{Nodes: Nodes{n}}
		case "way":
			w := &Way{}
			if err := d.DecodeElement(&w, &start); err != nil {
				return err
			}
			a.OSM = &OSM{Ways: Ways{w}}
		case "relation":
			r := &Relation{}
			if err := d.DecodeElement(&r, &start); err != nil {
				return err
			}
			a.OSM = &OSM{Relations: Relations{r}}
		}
	}

	return nil
}

// MarshalXML converts a diff action to xml creating the proper structures.
func (a Action) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "type"}, Value: string(a.Type)})
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if a.OSM != nil {
		if err := a.OSM.marshalInnerElementsXML(e); err != nil {
			return err
		}
	}

	if a.Old != nil {
		if err := marshalInnerChange(e, "old", a.Old); err != nil {
			return err
		}
	}

	if a.New != nil {
		if err := marshalInnerChange(e, "new", a.New); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

// ActionType is a strong type for the different diff actions.
type ActionType string

// The different types of diff actions.
const (
	ActionCreate ActionType = "create"
	ActionModify ActionType = "modify"
	ActionDelete ActionType = "delete"
)
