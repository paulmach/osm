package osmxml_test

import (
	"context"
	"fmt"
	"os"

	"github.com/nextmv-io/osm"
	"github.com/nextmv-io/osm/osmxml"
)

func ExampleScanner() {
	scanner := osmxml.New(context.Background(), os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Object().(*osm.Changeset))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
