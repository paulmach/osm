*.proto files were downloaded from https://github.com/scrosby/OSM-binary/tree/master/src and changed in following ways:

* To eliminate continuous conversions from `[]byte` to `string`, this

```protobuf
message StringTable {
   repeated bytes s = 1;
}
```

was changed to

```protobuf
message StringTable {
   repeated string s = 1;
}
```

This changes is expected to be fully compatible with all PBF files.
