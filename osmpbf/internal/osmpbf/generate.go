package osmpbf

// gogo protobuf seems to be about 10% faster
// https://github.com/gogo/protobuf
//go:generate protoc --gofast_out=. fileformat.proto osmformat.proto
