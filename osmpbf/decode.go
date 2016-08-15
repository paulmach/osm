package osmpbf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"

	"golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/paulmach/go.osm"
	"github.com/paulmach/go.osm/osmpbf/internal/osmpbf"
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

// iPair is the group sent on the chan into the decoder
// goroutines that unzip and decode the pbf from the headerblock.
type iPair struct {
	Blob *osmpbf.Blob
	Err  error
}

// oPair is the group sent on the chan out of the decoder
// goroutines. It'll contain a list of all the elements.
type oPair struct {
	Elements []osm.Element
	Err      error
}

// A Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type decoder struct {
	r io.Reader

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup

	// for data decoders
	inputs     []chan<- iPair
	outputs    []<-chan oPair
	serializer chan oPair

	cData  oPair
	cIndex int
}

// newDecoder returns a new decoder that reads from r.
func newDecoder(ctx context.Context, r io.Reader) *decoder {
	c, cancel := context.WithCancel(ctx)
	return &decoder{
		ctx:    c,
		cancel: cancel,
		r:      r,
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

	// read OSMHeader
	blobHeader, blob, err := dec.readFileBlock()
	if err != nil {
		return err
	}

	if blobHeader.GetType() != "OSMHeader" {
		return fmt.Errorf("unexpected first fileblock of type %s", blobHeader.GetType())
	}

	err = decodeOSMHeader(blob)
	if err != nil {
		return err
	}

	// High level overview of the decoder:
	// The decoder supports parallel unzipping and protobuf decoding of all
	// the header blocks. On goroutine feeds the headerblocks round-robin into
	// the input channels. n goroutines read from the input channel, decode
	// the block and put the elements on their output channel. A third type of
	// goroutines round-robin reads the output channels and feads them into the
	// serializer channel to maintain the order of the objects in the file.

	// start data decoders
	dec.wg.Add(n)
	for i := 0; i < n; i++ {
		input := make(chan iPair, n)
		output := make(chan oPair, n)
		go func() {
			defer close(output)
			defer dec.wg.Done()

			dd := &dataDecoder{}
			for p := range input {
				if dec.ctx.Err() != nil {
					return
				}

				var out oPair
				if p.Err == nil {
					// send decoded objects or decoding error
					objects, err := dd.Decode(p.Blob)
					out = oPair{objects, err}
				} else {
					// send input error as is
					out = oPair{nil, p.Err}
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
	dec.wg.Add(1)
	go func() {
		defer dec.wg.Done()
		defer func() {
			for _, input := range dec.inputs {
				close(input)
			}
		}()

		i := 0
		var err error
		for dec.ctx.Err() == nil || err == nil {
			input := dec.inputs[i]
			i = (i + 1) % n

			blobHeader, blob, err = dec.readFileBlock()
			if err == nil && blobHeader.GetType() != "OSMData" {
				err = fmt.Errorf("unexpected fileblock of type %s", blobHeader.GetType())
			}

			pair := iPair{Blob: blob, Err: nil}
			if err != nil {
				pair = iPair{Blob: nil, Err: err}
			}

			select {
			case input <- pair:
			case <-dec.ctx.Done():
			}
		}
	}()

	dec.wg.Add(1)
	go func() {
		defer dec.wg.Done()

		i := 0
		for {
			output := dec.outputs[i]
			i = (i + 1) % n

			var p oPair
			select {
			case <-dec.ctx.Done():
				p = oPair{Err: dec.ctx.Err()}
			case p = <-output:
			}

			if p.Elements != nil {
				dec.serializer <- p
			}

			if p.Err != nil {
				dec.serializer <- p
				close(dec.serializer)
				dec.cancel()
				return
			}
		}
	}()

	return nil
}

// Next reads the next element from the input stream and returns either a
// Node, Way or Relation struct representing the underlying OpenStreetMap PBF
// data, or error encountered. The end of the input stream is reported by an io.EOF error.
func (dec *decoder) Next() (osm.Element, error) {
	for dec.cIndex >= len(dec.cData.Elements) {
		cd, ok := <-dec.serializer
		if !ok {
			if dec.cData.Err != nil {
				return osm.Element{}, dec.cData.Err
			}
			return osm.Element{}, io.EOF
		}

		dec.cData = cd
		dec.cIndex = 0
	}

	v := dec.cData.Elements[dec.cIndex]
	dec.cIndex++
	return v, dec.cData.Err
}

func (dec *decoder) readFileBlock() (*osmpbf.BlobHeader, *osmpbf.Blob, error) {
	blobHeaderSize, err := dec.readBlobHeaderSize()
	if err != nil {
		return nil, nil, err
	}

	blobHeader, err := dec.readBlobHeader(blobHeaderSize)
	if err != nil {
		return nil, nil, err
	}

	blob, err := dec.readBlob(blobHeader)
	if err != nil {
		return nil, nil, err
	}

	return blobHeader, blob, err
}

func (dec *decoder) readBlobHeaderSize() (uint32, error) {
	buf := make([]byte, 4, 4)
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return 0, err
	}

	size := binary.BigEndian.Uint32(buf)
	if size >= maxBlobHeaderSize {
		return 0, errors.New("BlobHeader size >= 64Kb")
	}
	return size, nil
}

func (dec *decoder) readBlobHeader(size uint32) (*osmpbf.BlobHeader, error) {
	buf := make([]byte, size, size)
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blobHeader := &osmpbf.BlobHeader{}
	if err := proto.Unmarshal(buf, blobHeader); err != nil {
		return nil, err
	}

	if blobHeader.GetDatasize() >= maxBlobSize {
		return nil, errors.New("Blob size >= 32Mb")
	}
	return blobHeader, nil
}

func (dec *decoder) readBlob(blobHeader *osmpbf.BlobHeader) (*osmpbf.Blob, error) {
	buf := make([]byte, blobHeader.GetDatasize())
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blob := &osmpbf.Blob{}
	if err := proto.Unmarshal(buf, blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func getData(blob *osmpbf.Blob) ([]byte, error) {
	switch {
	case blob.Raw != nil:
		return blob.GetRaw(), nil

	case blob.ZlibData != nil:
		r, err := zlib.NewReader(bytes.NewReader(blob.GetZlibData()))
		if err != nil {
			return nil, err
		}

		// using the bytes.Buffer allows for the preallocation of the necessary space.
		buf := bytes.NewBuffer(make([]byte, 0, blob.GetRawSize()+bytes.MinRead))
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

func decodeOSMHeader(blob *osmpbf.Blob) error {
	data, err := getData(blob)
	if err != nil {
		return err
	}

	headerBlock := &osmpbf.HeaderBlock{}
	if err := proto.Unmarshal(data, headerBlock); err != nil {
		return err
	}

	// Check we have the parse capabilities
	requiredFeatures := headerBlock.GetRequiredFeatures()
	for _, feature := range requiredFeatures {
		if !parseCapabilities[feature] {
			return fmt.Errorf("parser does not have %s capability", feature)
		}
	}

	return nil
}
