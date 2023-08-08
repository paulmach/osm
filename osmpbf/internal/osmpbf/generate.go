package osmpbf

//go:generate protoc  --proto_path=. --go_opt=module=github.com/paulmach/osmpbf/internal/osmpbf  --go_out=.  fileformat.proto osmformat.proto
