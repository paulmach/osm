package osmpb

// gogo protobuf seems to be about 10% faster
// https://github.com/gogo/protobuf
//
// setup:
//    brew install protobuf
//    go get github.com/gogo/protobuf/...
//
//go:generate protoc --proto_path=$GOPATH/src/:$GOPATH/src/github.com/gogo/protobuf/protobuf:. --gogofaster_out=. osm.proto
