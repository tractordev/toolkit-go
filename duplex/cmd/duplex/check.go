package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"tractor.dev/toolkit/duplex/codec"
	"tractor.dev/toolkit/duplex/fn"
	"tractor.dev/toolkit/duplex/interop"
	"tractor.dev/toolkit/duplex/mux"
	"tractor.dev/toolkit/duplex/pipe/quic"
	"tractor.dev/toolkit/duplex/rpc"
	"tractor.dev/toolkit/engine/cli"
)

var vals = []any{
	100,
	true,
	"hello",
	map[string]any{"foo": "bar"},
	[]any{1, 2, 3},
}

var checkCmd = &cli.Command{
	Usage: "check",
	Short: "check interop",
	Run: func(ctx context.Context, args []string) {
		log.SetOutput(os.Stderr)

		var c codec.Codec = codec.CBORCodec{}
		if os.Getenv("QTALK_CODEC") == "json" {
			log.Println("* Using JSON codec")
			c = codec.JSONCodec{}
		}

		var cmd *exec.Cmd
		var sess mux.Session
		var err error
		var u *url.URL

		if len(args) == 0 {
			// self check
			path, err := os.Executable()
			fatal(err)
			cmd = exec.Command(path, "interop")
		} else {
			u, err = url.Parse(args[0])
			if err != nil || u.Scheme == "" {
				// check against subprocess
				path, err := exec.LookPath("sh")
				fatal(err)
				cmd = exec.Command(path, "-c", args[0])
			}
		}

		if cmd != nil {
			cmd.Stderr = os.Stderr
			wc, err := cmd.StdinPipe()
			if err != nil {
				fatal(err)
			}
			rc, err := cmd.StdoutPipe()
			if err != nil {
				fatal(err)
			}
			sess, err = mux.DialIO(wc, rc)
			if err != nil {
				fatal(err)
			}
			if err := cmd.Start(); err != nil {
				fatal(err)
			}
			defer func() {
				cmd.Process.Signal(os.Interrupt)
				cmd.Wait()
			}()
		} else {
			switch u.Scheme {
			case "udp":
				// check against remote quic endpoint
				sess, err = quic.Dial(u.Host, false)
				fatal(err)
			case "tcp":
				sess, err = mux.DialTCP(u.Host)
				fatal(err)
			default:
				fatal(errors.New("unsupported protocol"))
			}
		}

		defer sess.Close()

		srv := rpc.Server{
			Handler: fn.HandlerFrom(interop.CallbackService{}),
			Codec:   c,
		}
		go srv.Respond(sess, nil)

		caller := rpc.NewClient(sess, c)
		var ret any

		// Error check
		_, err = caller.Call(ctx, "Error", "test", nil)
		if err == nil {
			log.Fatal("expected error")
		}
		fmt.Println("Error:", strings.TrimPrefix(err.Error(), "remote: "))
		_, err = caller.Call(ctx, "BadSelector", "test", nil)
		if err == nil {
			log.Fatal("expected error")
		}
		fmt.Println("Error:", strings.TrimPrefix(err.Error(), "remote: "))

		// Unary check
		for _, v := range vals {
			_, err = caller.Call(ctx, "Unary", v, &ret)
			fatal(err)
			fmt.Println("Unary:", v, ret)
		}

		// Stream check
		resp, err := caller.Call(ctx, "Stream", nil, nil)
		fatal(err)
		go func() {
			for _, v := range vals {
				fatal(resp.Send(v))
			}
			fatal(resp.Channel.CloseWrite())
		}()
		for {
			err = resp.Receive(&ret)
			if err != nil {
				if err != io.EOF {
					log.Println(err)
				}
				break
			}
			fmt.Println("Stream:", ret)
		}

		// Bytes check
		// 1 byte, 1kb, 1mb
		for _, v := range []int{1, 1024, 1 << 20} {
			data := make([]byte, v)
			rand.Read(data)
			resp, err = caller.Call(ctx, "Bytes", nil, nil)
			fatal(err)
			var buf bytes.Buffer
			go func() {
				io.Copy(resp.Channel, bytes.NewBuffer(data))
				resp.Channel.CloseWrite()
			}()
			io.Copy(&buf, resp.Channel)
			if buf.Len() != len(data) {
				log.Fatal("byte stream buffer does not match")
			}
			fmt.Println("Bytes:", buf.Len())
		}
	},
}
