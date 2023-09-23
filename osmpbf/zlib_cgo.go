//go:build cgo
// +build cgo

package osmpbf

import (
	"fmt"

	deflate "github.com/4kills/go-libdeflate/v2"
)

func decompress(in []byte, size int, data []byte) ([]byte, error) {
	if cap(data) > (int)(size) {
		data = data[0:size]
	} else {
		data = nil
	}

	_, buf, err := deflate.DecompressZlib(in, data)
	if len(buf) != int(size) {
		return nil, fmt.Errorf("raw blob data size %d but expected %d", len(buf), size)
	}

	return buf, err
}
