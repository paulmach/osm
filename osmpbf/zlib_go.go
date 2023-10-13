//go:build !cgo
// +build !cgo

package osmpbf

import (
	"bytes"
	"compress/zlib"
	"io"
)

func zlibReader(data []byte) (io.ReadCloser, error) {
	return zlib.NewReader(bytes.NewReader(data))
}

func zlibWriter(w io.Writer) io.WriteCloser {
	return zlib.NewWriter(w)
}
