package main

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/hanwen/go-fuse/v2/fs"
	"tractor.dev/toolkit-go/engine/fs/fuser"
	"tractor.dev/toolkit-go/engine/fs/githubfs"
)

func main() {
	debug := flag.Bool("debug", false, "print debug data")

	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatal("Usage:\n  fusedev [-debug] REPO MOUNTPOINT")
	}

	repo := strings.Split(flag.Arg(0), "/")
	if len(repo) < 2 {
		log.Fatal("Usage:\n  fusedev [-debug] PATH MOUNTPOINT")
	}
	// osfs := workingpathfs.New(osfs.New(), flag.Arg(0))
	gfs := githubfs.New(repo[0], repo[1], os.Getenv("GH_TOKEN"))

	opts := &fs.Options{}
	//opts.Debug = *debug
	if !*debug {
		log.SetOutput(io.Discard)
	}

	server, err := fs.Mount(flag.Arg(1), &fuser.Node{FS: gfs}, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}

	server.Wait()
}
