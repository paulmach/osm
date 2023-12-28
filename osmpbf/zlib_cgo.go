//go:build cgo
// +build cgo

package osmpbf

import (
	deflate "github.com/4kills/go-libdeflate/v2"
)

func decompress(in []byte, out []byte) ([]byte, error) {
	_, buf, err := deflate.DecompressZlib(in, out)
	return buf, err
}
