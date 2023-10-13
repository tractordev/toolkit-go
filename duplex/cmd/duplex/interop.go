package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math/big"
	"os"

	"tractor.dev/toolkit-go/duplex/codec"
	"tractor.dev/toolkit-go/duplex/fn"
	"tractor.dev/toolkit-go/duplex/interop"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/duplex/rpc"
	"tractor.dev/toolkit-go/duplex/x/quic"
	"tractor.dev/toolkit-go/engine/cli"
)

var interopCmd = &cli.Command{
	Usage: "interop",
	Short: "run interop service",
	Args:  cli.MaxArgs(1),
	Run: func(ctx context.Context, args []string) {
		log.SetOutput(os.Stderr)

		var c codec.Codec = codec.CBORCodec{}
		if os.Getenv("QTALK_CODEC") == "json" {
			c = codec.JSONCodec{}
		}

		if len(args) == 0 {
			// STDIO
			sess, err := mux.DialStdio()
			fatal(err)
			serve(sess, c)
			return
		}

		// QUIC
		log.Printf("* Listening on %s...\n", args[0])
		l, err := quic.Listen(args[0], generateTLSConfig(), nil)
		fatal(err)
		defer l.Close()

		for {
			conn, err := l.Accept(context.Background())
			fatal(err)
			go serve(quic.New(conn), c)
		}
	},
}

func serve(sess mux.Session, c codec.Codec) {
	srv := rpc.Server{
		Handler: fn.HandlerFrom(interop.InteropService{}),
		Codec:   c,
	}
	srv.Respond(sess, nil)
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		NextProtos:   []string{quic.Protocol},
		Certificates: []tls.Certificate{tlsCert},
	}
}
