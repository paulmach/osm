package osmxml_test

import (
	"fmt"
	"os"

	"golang.org/x/net/context"

	"github.com/paulmach/go.osm/osmxml"
)

func ExampleChangesetScanner() {
	scanner := osmxml.New(context.Background(), os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Element().Changeset) // Println will add back the final '\n'
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
