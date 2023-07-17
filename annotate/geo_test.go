package annotate

import (
	"encoding/xml"
	"testing"

	"github.com/paulmach/orb"
	"github.com/onXmaps/osm"
)

func TestWayPointOnSurface(t *testing.T) {
	data := `
<way id="38238655">
	<nd lat="41.4176729" lon="-81.8752338"/>
	<nd lat="41.418435" lon="-81.874286"/>
	<nd lat="41.418526" lon="-81.873181"/>
	<nd lat="41.418548" lon="-81.868659"/>
	<nd lat="41.419093" lon="-81.856178"/>
	<nd lat="41.419033" lon="-81.85595"/>
</way>`

	var w *osm.Way
	if err := xml.Unmarshal([]byte(data), &w); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	sp := wayPointOnSurface(w)

	expected := orb.Point{w.Nodes[3].Lon, w.Nodes[3].Lat}
	if !sp.Equal(expected) {
		t.Errorf("incorrect centroid: %v", sp)
	}
}
