//go:build !cgo
// +build !cgo

package osmpbf

import (
	"bytes"
	"compress/zlib"
)

func decompress(in []byte, out []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(out)
	if _, err = buf.ReadFrom(r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
