package talk

import (
	"tractor.dev/toolkit/duplex/codec"
	"tractor.dev/toolkit/duplex/mux"
	"tractor.dev/toolkit/duplex/rpc"
)

// Peer is a mux session, RPC client and responder, all in one.
type Peer struct {
	mux.Session
	*rpc.Client
	*rpc.Server
	*rpc.RespondMux
	codec.Codec
}

// NewPeer returns a Peer based on a session and codec.
func NewPeer(session mux.Session, codec codec.Codec) *Peer {
	mux := rpc.NewRespondMux()
	return &Peer{
		Session:    session,
		Codec:      codec,
		Client:     rpc.NewClient(session, codec),
		Server:     &rpc.Server{Handler: mux, Codec: codec},
		RespondMux: mux,
	}
}

// Close will close the underlying session.
func (p *Peer) Close() error {
	return p.Client.Close()
}

// Respond lets the Peer respond to incoming channels like
// a server, using any registered handlers.
func (p *Peer) Respond() {
	p.Server.Respond(p.Session, nil)
}
