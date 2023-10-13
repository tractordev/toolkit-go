package duplex

import (
	"context"
	"testing"

	"tractor.dev/toolkit-go/duplex/codec"
	"tractor.dev/toolkit-go/duplex/fn"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/duplex/talk"
)

func fatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func equal(t *testing.T, a, b any, message string) {
	if a != b {
		t.Log(message)
		t.Fail()
	}
}

func makeClientServer(service any) (client, server *talk.Peer, closer func()) {
	c, s := mux.Pair()
	client = talk.NewPeer(c, codec.CBORCodec{})
	server = talk.NewPeer(s, codec.CBORCodec{})
	server.Server.Handler = fn.HandlerFrom(service)
	go client.Respond()
	go server.Respond()
	closer = func() {
		client.Close()
		server.Close()
	}
	return
}

type TestService struct {
	T *testing.T
}

func (s *TestService) Hello() string {
	return "Hello"
}

func (s *TestService) NilAnyArg(a string, v any) {
	equal(s.T, v, nil, "nil any arg not nil")
}

func TestBasic(t *testing.T) {
	c, _, closer := makeClientServer(&TestService{T: t})
	defer closer()

	var ret string
	_, err := c.Call(context.Background(), "Hello", nil, &ret)
	fatal(t, err)

	equal(t, ret, "Hello", "return not Hello")
}

func TestNilAnyArg(t *testing.T) {
	c, _, closer := makeClientServer(&TestService{T: t})
	defer closer()

	_, err := c.Call(context.Background(), "NilAnyArg", fn.Args{"abc", nil}, nil)
	fatal(t, err)

}

// this test ensures the io.Pipe based mux.Pair buffers enough writes
// to allow a simultaneous call or channel open, since a lockstep pipe
// will deadlock unable to write the open confirm since the other end
// will also be trying to write an open confirm before reading next packet
func TestSimultaneousOpen(t *testing.T) {
	c, s := mux.Pair()
	client := talk.NewPeer(c, codec.CBORCodec{})
	client.Server.Handler = fn.HandlerFrom(&TestService{T: t})
	server := talk.NewPeer(s, codec.CBORCodec{})
	server.Server.Handler = fn.HandlerFrom(&TestService{T: t})
	go client.Respond()
	go server.Respond()

	cdone := make(chan bool)
	sdone := make(chan bool)

	go func() {
		var ret any
		_, err := client.Call(context.Background(), "Hello", nil, &ret)
		fatal(t, err)
		close(cdone)
	}()

	go func() {
		var ret any
		_, err := server.Call(context.Background(), "Hello", nil, &ret)
		fatal(t, err)
		close(sdone)
	}()

	<-sdone
	<-cdone

	client.Close()
	server.Close()
}
