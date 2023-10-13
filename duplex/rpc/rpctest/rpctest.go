package rpctest

import (
	"io"

	"tractor.dev/toolkit-go/duplex/codec"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/duplex/rpc"
)

// NewPair creates a Client and Server connected by in-memory pipes.
// The server Respond method is called in a goroutine. Only the client
// should need to be cleaned up with call to Close.
func NewPair(handler rpc.Handler, codec codec.Codec) (*rpc.Client, *rpc.Server) {
	ar, bw := io.Pipe()
	br, aw := io.Pipe()
	sessA, _ := mux.DialIO(aw, ar)
	sessB, _ := mux.DialIO(bw, br)

	srv := &rpc.Server{
		Codec:   codec,
		Handler: handler,
	}
	go srv.Respond(sessA, nil)

	return rpc.NewClient(sessB, codec), srv
}
