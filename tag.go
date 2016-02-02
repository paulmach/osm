package osm

type Tag struct {
	Key   string `xml:"k,attr"`
	Value string `xml:"v,attr"`
}

type Tags []Tag

func (ts Tags) Find(k string) string {
	for _, t := range ts {
		if t.Key == k {
			return t.Value
		}
	}

	return ""
}
