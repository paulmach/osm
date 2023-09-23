//go:build !cgo
// +build !cgo

package osmpbf

import (
	"bytes"
	"compress/zlib"
	"fmt"
)

func decompress(in []byte, size int, data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}

	// using the bytes.Buffer allows for the preallocation of the necessary space.
	l := size + bytes.MinRead
	if cap(data) < int(l) {
		data = make([]byte, 0, l+l/10)
	} else {
		data = data[:0]
	}
	buf := bytes.NewBuffer(data)
	if _, err = buf.ReadFrom(r); err != nil {
		return nil, err
	}

	if buf.Len() != int(size) {
		return nil, fmt.Errorf("raw blob data size %d but expected %d", buf.Len(), size)
	}
	return buf.Bytes(), nil
}
