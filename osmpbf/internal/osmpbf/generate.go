package osmpbf

// gogo protobuf seems to be about 10% faster
// https://github.com/gogo/protobuf
//go:generate protoc --proto_path=../../../../../..:../../../../../gogo/protobuf/protobuf:. --gogofaster_out=. fileformat.proto osmformat.proto
