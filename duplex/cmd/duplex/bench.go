package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"tractor.dev/toolkit-go/duplex/codec"
	"tractor.dev/toolkit-go/duplex/fn"
	"tractor.dev/toolkit-go/duplex/interop"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/duplex/pipe/quic"
	"tractor.dev/toolkit-go/duplex/rpc"
	"tractor.dev/toolkit-go/engine/cli"
)

var benchCmd = &cli.Command{
	Usage: "bench",
	Short: "interop benchmark",
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

		if len(args) == 0 {
			// self check
			path, err := os.Executable()
			fatal(err)
			cmd = exec.Command(path, "interop")
		} else if !strings.HasPrefix(args[0], "udp://") {
			// check against subprocess
			path, err := exec.LookPath("sh")
			fatal(err)
			cmd = exec.Command(path, "-c", args[0])
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
			// check against remote quic endpoint
			sess, err = quic.Dial(strings.TrimPrefix(args[0], "udp://"), false)
			fatal(err)
		}

		defer sess.Close()

		srv := rpc.Server{
			Handler: fn.HandlerFrom(interop.CallbackService{}),
			Codec:   c,
		}
		go srv.Respond(sess, nil)

		caller := rpc.NewClient(sess, c)
		//var ret any
		//var err error

		// Bytes check
		// 1mb
		mb := 1 << 20
		for _, v := range []int{mb * 256, mb * 512, mb * 1024} {
			data := make([]byte, v)
			rand.Read(data)
			start := time.Now()
			resp, err := caller.Call(ctx, "Bytes", nil, nil)
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
			diff := time.Now().Sub(start)
			fmt.Println("Bytes:", buf.Len()/mb, "MB", "RTT:", diff, "Thru:", int(float64(buf.Len())/diff.Seconds()/(1024*1024)), "MB/s")
		}
	},
}
