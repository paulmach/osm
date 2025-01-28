// +build cgo

package osmpbf

import (
	"bytes"
	"io"

	"github.com/DataDog/czlib"
)

func zlibReader(data []byte) (io.ReadCloser, error) {
	return czlib.NewReader(bytes.NewReader(data))
}
