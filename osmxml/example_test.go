package osmxml_test

import (
	"context"
	"fmt"
	"os"

	osm "github.com/paulmach/go.osm"
	"github.com/paulmach/go.osm/osmxml"
)

func ExampleScanner() {
	scanner := osmxml.New(context.Background(), os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Element().(*osm.Changeset))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
