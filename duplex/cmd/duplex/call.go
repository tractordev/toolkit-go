package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/progrium/clon-go"
	"tractor.dev/toolkit/duplex/codec"
	"tractor.dev/toolkit/duplex/talk"
	"tractor.dev/toolkit/engine/cli"
)

var callCmd = &cli.Command{
	Usage: "call",
	Short: "call a remote function",
	Args:  cli.MinArgs(1),
	Run: func(ctx context.Context, args []string) {
		log.SetOutput(os.Stderr)
		u, err := url.Parse(args[0])
		if err != nil {
			log.Fatal(err)
		}

		var sargs any
		if len(args) > 1 {
			sargs, err = clon.Parse(args[1:])
			if err != nil {
				log.Fatal(err)
			}
		}

		var c codec.Codec = codec.CBORCodec{}
		if os.Getenv("QTALK_CODEC") == "json" {
			log.Println("* Using JSON codec")
			c = codec.JSONCodec{}
		}

		peer, err := talk.Dial(u.Scheme, u.Host, c)
		if err != nil {
			log.Fatal(err)
		}
		defer peer.Close()

		var ret any
		_, err = peer.Call(context.Background(), u.Path, sargs, &ret)
		if err != nil {
			log.Fatal(err)
		}

		b, err := json.MarshalIndent(ret, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	},
}
