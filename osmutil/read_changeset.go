package osmutil

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/paulmach/go.osm"
	"golang.org/x/net/context"
)

// A ChangesetItem is returned by the ReadChangeset channel to enable to returning
// of error values.
type ChangesetItem struct {
	Changeset *osm.Changeset
	Err       error
}

// ReadChangesets reads from the xml.Decoders and returns ChangesetItems will allow
// for the returning of errors. Errors will occure if the underlying reader has an issue
// or if the data does not match osm xml changesets. On error the channel will be closed.
func ReadChangesets(ctx context.Context, decoder *xml.Decoder) <-chan *ChangesetItem {
	cc := make(chan *ChangesetItem)

	go func() {
		defer close(cc)

		for {
			if ctx.Err() != nil {
				return
			}

			t, err := decoder.Token()
			if err == io.EOF {
				return // we're done
			}

			if err != nil {
				cc <- &ChangesetItem{Err: err}
				return
			}

			se, ok := t.(xml.StartElement)
			if !ok {
				continue
			}

			if strings.ToLower(se.Name.Local) == "changeset" {
				c := &osm.Changeset{}
				err = decoder.DecodeElement(&c, &se)
				if err != nil {
					cc <- &ChangesetItem{Err: err}
					return
				}

				cc <- &ChangesetItem{Changeset: c}
			}
		}
	}()

	return cc
}
