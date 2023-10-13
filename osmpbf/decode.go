package osmpbf

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/nextmv-io/osm"
	"github.com/nextmv-io/osm/osmpbf/internal/osmpbf"
	"google.golang.org/protobuf/proto"
)

const (
	maxBlobHeaderSize = 64 * 1024
	maxBlobSize       = 32 * 1024 * 1024
)

var (
	parseCapabilities = map[string]bool{
		"OsmSchema-V0.6":        true,
		"DenseNodes":            true,
		"HistoricalInformation": true,
	}
)

// osm block data types
const (
	osmHeaderType = "OSMHeader"
	osmDataType   = "OSMData"
)

// Header contains the contents of the header in the pbf file.
type Header struct {
	Bounds               *osm.Bounds
	RequiredFeatures     []string
	OptionalFeatures     []string
	WritingProgram       string
	Source               string
	ReplicationTimestamp time.Time
	ReplicationSeqNum    uint64
	ReplicationBaseURL   string
}

// iPair is the group sent on the chan into the decoder
// goroutines that unzip and decode the pbf from the headerblock.
type iPair struct {
	Offset int64
	Blob   *osmpbf.Blob
	Err    error
}

// oPair is the group sent on the chan out of the decoder
// goroutines. It'll contain a list of all the objects.
type oPair struct {
	Offset  int64
	Objects []osm.Object
	Err     error
}

// A Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type decoder struct {
	scanner *Scanner

	header    *Header
	r         io.Reader
	bytesRead int64

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup

	// for data decoders
	inputs     []chan<- iPair
	outputs    []<-chan oPair
	serializer chan oPair

	pOffset int64
	cOffset int64
	cData   oPair
	cIndex  int
}

// newDecoder returns a new decoder that reads from r.
func newDecoder(ctx context.Context, s *Scanner, r io.Reader) *decoder {
	c, cancel := context.WithCancel(ctx)
	return &decoder{
		scanner: s,
		ctx:     c,
		cancel:  cancel,
		r:       r,
	}
}

func (dec *decoder) Close() error {
	dec.cancel()
	dec.wg.Wait()
	return nil
}

// Start decoding process using n goroutines.
func (dec *decoder) Start(n int) error {
	if n < 1 {
		n = 1
	}
	dec.serializer = make(chan oPair, n)

	sizeBuf := make([]byte, 4)
	headerBuf := make([]byte, maxBlobHeaderSize)
	blobBuf := make([]byte, maxBlobSize)

	// read OSMHeader
	// NOTE: if the first block is not a header, i.e. after a restart we need
	// to decode that block. It gets pushed on the first "input" below.
	blobHeader, blob, err := dec.readFileBlock(sizeBuf, headerBuf, blobBuf)
	if err != nil {
		return err
	}

	if blobHeader.GetType() == osmHeaderType {
		var err error
		dec.header, err = decodeOSMHeader(blob)
		if err != nil {
			return err
		}
	}

	dec.wg.Add(n + 2)

	//use roughly 10 chanel inputs
	numChanels := 10 / n

	// High level overview of the decoder:
	// The decoder supports parallel unzipping and protobuf decoding of all
	// the header blocks. On goroutine feeds the headerblocks round-robin into
	// the input channels. n goroutines read from the input channel, decode
	// the block and put the objects on their output channel. A third type of
	// goroutines round-robin reads the output channels and feads them into the
	// serializer channel to maintain the order of the objects in the file.

	// start data decoders
	for i := 0; i < n; i++ {
		input := make(chan iPair, numChanels)
		output := make(chan oPair, numChanels)

		dd := &dataDecoder{scanner: dec.scanner}

		go func() {
			defer close(output)
			defer dec.wg.Done()

			for p := range input {
				var out oPair
				if p.Err == nil {
					// send decoded objects or decoding error
					objects, err := dd.Decode(p.Blob)
					out = oPair{Offset: p.Offset, Objects: objects, Err: err}
				} else {
					out = oPair{Err: p.Err} // send input error as is
				}

				select {
				case output <- out:
				case <-dec.ctx.Done():
				}
			}
		}()

		dec.inputs = append(dec.inputs, input)
		dec.outputs = append(dec.outputs, output)
	}

	// start reading OSMData
	go func() {
		defer dec.wg.Done()
		defer func() {
			for _, input := range dec.inputs {
				close(input)
			}
		}()

		var (
			i   int
			err error
		)

		// On restart the first block may not be a header and will need to be
		// added to the first input.
		if blobHeader.GetType() != osmHeaderType {
			dec.inputs[0] <- iPair{Offset: 0, Blob: blob, Err: err}

			i = (i + 1) % n
		}

		for dec.ctx.Err() == nil || err == nil {
			input := dec.inputs[i]
			i = (i + 1) % n

			offset := dec.bytesRead
			blobHeader, blob, err = dec.readFileBlock(sizeBuf, headerBuf, blobBuf)
			if err == nil && blobHeader.GetType() != osmDataType {
				err = fmt.Errorf("unexpected fileblock of type %s", blobHeader.GetType())
			}

			pair := iPair{Offset: offset, Blob: blob}
			if err != nil {
				pair = iPair{Err: err}
			}

			select {
			case input <- pair:
			case <-dec.ctx.Done():
			}
		}
	}()

	go func() {
		defer dec.wg.Done()
		defer func() {
			close(dec.serializer)
			dec.cancel()
		}()

		for i := 0; ; i = (i + 1) % n {
			output := dec.outputs[i]

			var p oPair
			select {
			case p = <-output:
			case <-dec.ctx.Done():
				dec.cData.Err = dec.ctx.Err()
				return
			}

			select {
			case dec.serializer <- p:
			case <-dec.ctx.Done():
				dec.cData.Err = dec.ctx.Err()
				return
			}

			if p.Err != nil {
				return
			}
		}
	}()

	return nil
}

// Next reads the next object from the input stream and returns either a
// Node, Way or Relation struct representing the underlying OpenStreetMap PBF
// data, or error encountered. The end of the input stream is reported by an io.EOF error.
func (dec *decoder) Next() (osm.Object, error) {
	for dec.cIndex >= len(dec.cData.Objects) {
		cd, ok := <-dec.serializer
		if !ok || cd.Err == io.EOF {
			if dec.cData.Err != nil {
				return nil, dec.cData.Err
			}
			return nil, io.EOF
		}

		dec.pOffset = dec.cOffset
		dec.cOffset = cd.Offset
		dec.cData = cd
		dec.cIndex = 0
	}

	v := dec.cData.Objects[dec.cIndex]
	dec.cIndex++
	return v, dec.cData.Err
}

func (dec *decoder) readFileBlock(sizeBuf, headerBuf, blobBuf []byte) (*osmpbf.BlobHeader, *osmpbf.Blob, error) {
	blobHeaderSize, err := dec.readBlobHeaderSize(sizeBuf)
	if err != nil {
		return nil, nil, err
	}

	headerBuf = headerBuf[:blobHeaderSize]
	blobHeader, err := dec.readBlobHeader(headerBuf)
	if err != nil {
		return nil, nil, err
	}

	blobBuf = blobBuf[:blobHeader.GetDatasize()]
	blob, err := dec.readBlob(blobBuf)
	if err != nil {
		return nil, nil, err
	}

	dec.bytesRead += 4 + int64(blobHeaderSize) + int64(blobHeader.GetDatasize())
	return blobHeader, blob, nil
}

func (dec *decoder) readBlobHeaderSize(buf []byte) (uint32, error) {
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return 0, err
	}

	size := binary.BigEndian.Uint32(buf)
	if size >= maxBlobHeaderSize {
		return 0, errors.New("blobHeader size >= 64Kb")
	}
	return size, nil
}

func (dec *decoder) readBlobHeader(buf []byte) (*osmpbf.BlobHeader, error) {
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blobHeader := &osmpbf.BlobHeader{}
	if err := proto.Unmarshal(buf, blobHeader); err != nil {
		return nil, err
	}

	if blobHeader.GetDatasize() >= maxBlobSize {
		return nil, errors.New("blob size >= 32Mb")
	}
	return blobHeader, nil
}

func (dec *decoder) readBlob(buf []byte) (*osmpbf.Blob, error) {
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blob := &osmpbf.Blob{}
	if err := proto.Unmarshal(buf, blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func getData(blob *osmpbf.Blob, data []byte) ([]byte, error) {
	switch {
	case blob.Raw != nil:
		return blob.GetRaw(), nil

	case blob.ZlibData != nil:
		r, err := zlibReader(blob.GetZlibData())
		if err != nil {
			return nil, err
		}

		// using the bytes.Buffer allows for the preallocation of the necessary space.
		l := blob.GetRawSize() + bytes.MinRead
		if cap(data) < int(l) {
			data = make([]byte, 0, l+l/10)
		} else {
			data = data[:0]
		}
		buf := bytes.NewBuffer(data)
		if _, err = buf.ReadFrom(r); err != nil {
			return nil, err
		}

		if buf.Len() != int(blob.GetRawSize()) {
			return nil, fmt.Errorf("raw blob data size %d but expected %d", buf.Len(), blob.GetRawSize())
		}

		return buf.Bytes(), nil
	default:
		return nil, errors.New("unknown blob data")
	}
}

func decodeOSMHeader(blob *osmpbf.Blob) (*Header, error) {
	data, err := getData(blob, nil)
	if err != nil {
		return nil, err
	}

	headerBlock := &osmpbf.HeaderBlock{}
	if err := proto.Unmarshal(data, headerBlock); err != nil {
		return nil, err
	}

	// Check we have the parse capabilities
	requiredFeatures := headerBlock.GetRequiredFeatures()
	for _, feature := range requiredFeatures {
		if !parseCapabilities[feature] {
			return nil, fmt.Errorf("parser does not have %s capability", feature)
		}
	}

	// read the header
	header := &Header{
		RequiredFeatures:   headerBlock.GetRequiredFeatures(),
		OptionalFeatures:   headerBlock.GetOptionalFeatures(),
		WritingProgram:     headerBlock.GetWritingprogram(),
		Source:             headerBlock.GetSource(),
		ReplicationBaseURL: headerBlock.GetOsmosisReplicationBaseUrl(),
		ReplicationSeqNum:  uint64(headerBlock.GetOsmosisReplicationSequenceNumber()),
	}

	// convert timestamp epoch seconds to golang time structure if it exists
	if headerBlock.OsmosisReplicationTimestamp != nil {
		header.ReplicationTimestamp = time.Unix(*headerBlock.OsmosisReplicationTimestamp, 0).UTC()
	}
	// read bounding box if it exists
	if headerBlock.Bbox != nil {
		// Units are always in nanodegree and do not obey granularity rules. See osmformat.proto
		header.Bounds = &osm.Bounds{
			MinLon: 1e-9 * float64(*headerBlock.Bbox.Left),
			MaxLon: 1e-9 * float64(*headerBlock.Bbox.Right),
			MinLat: 1e-9 * float64(*headerBlock.Bbox.Bottom),
			MaxLat: 1e-9 * float64(*headerBlock.Bbox.Top),
		}
	}

	return header, nil
}
