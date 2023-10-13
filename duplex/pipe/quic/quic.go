package quic

import (
	"context"
	"crypto/tls"
	"io"
	"strings"

	"github.com/quic-go/quic-go"
	"tractor.dev/toolkit/duplex/mux"
)

const Protocol = "qtalk-quic"

func New(conn quic.Connection) mux.Session {
	return &session{conn}
}

var defaultTLSConfig = tls.Config{
	NextProtos: []string{Protocol},
}

type Config = quic.Config
type Connection = quic.Connection
type Listener = quic.Listener

func Listen(addr string, tlsConf *tls.Config, config *Config) (*Listener, error) {
	return quic.ListenAddr(addr, tlsConf, config)
}

func Dial(addr string, tlsVerify bool) (mux.Session, error) {
	cfg := defaultTLSConfig.Clone()
	cfg.InsecureSkipVerify = !tlsVerify
	conn, err := quic.DialAddr(context.Background(), addr, cfg, nil)
	if err != nil {
		return nil, err
	}
	return New(conn), nil
}

func init() {
	// TODO: figure out better way to deal with Dialers with arguments
	//talk.Dialers["quic"] = Dial
}

type session struct {
	conn quic.Connection
}

func (s *session) Close() error {
	return s.conn.CloseWithError(42, "close connection")
}

func (s *session) Accept() (mux.Channel, error) {
	stream, err := s.conn.AcceptStream(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "close connection") {
			return nil, io.EOF
		}
		return nil, err
	}
	header := make([]byte, 1)
	_, err = stream.Read(header)
	if err != nil {
		return nil, err
	}
	return &channel{stream}, nil
}

func (s *session) Open(ctx context.Context) (mux.Channel, error) {
	// TODO Make this wait for an acknowledgement from the remote that it has
	// accepted the connection. It writes some data in order to notify the remote
	// of the new stream immediately, but my initial attempt to send an
	// acknowledgement from the remote side lead to deadlocks in the tests.
	stream, err := s.conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	_, err = stream.Write([]byte("!"))
	if err != nil {
		return nil, err
	}
	return &channel{stream}, nil
}

func (s *session) Wait() error {
	<-s.conn.Context().Done()
	return s.conn.Context().Err()
}

type channel struct {
	stream quic.Stream
}

func (c *channel) ID() uint32 {
	return uint32(c.stream.StreamID())
}

func (c *channel) Read(p []byte) (int, error) {
	return c.stream.Read(p)
}

func (c *channel) Write(p []byte) (int, error) {
	return c.stream.Write(p)
}

func (c *channel) Close() error {
	c.stream.CancelRead(42)
	return c.CloseWrite()
}

func (c *channel) CloseWrite() error {
	// TODO this may need a lock to avoid concurrent call with Write
	return c.stream.Close()
}
