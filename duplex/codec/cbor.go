package codec

import (
	"io"

	"github.com/fxamacker/cbor/v2"
)

// CBORCodec provides a codec API for a CBOR encoder and decoder.
type CBORCodec struct{}

// Encoder returns a CBOR encoder
func (c CBORCodec) Encoder(w io.Writer) Encoder {
	return cbor.NewEncoder(w)
}

// Decoder returns a CBOR decoder
func (c CBORCodec) Decoder(r io.Reader) Decoder {
	return cbor.NewDecoder(r)
}
