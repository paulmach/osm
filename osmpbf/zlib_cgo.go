//go:build cgo
// +build cgo

package osmpbf

import (
	"bytes"
	"io"

	"github.com/datadog/czlib"
)

func zlibReader(data []byte) (io.ReadCloser, error) {
	return czlib.NewReader(bytes.NewReader(data))
}

func zlibWriter(w io.Writer) io.WriteCloser {
	return czlib.NewWriter(w)
}
