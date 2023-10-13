package main

import (
	"context"
	"log"
	"os"

	"tractor.dev/toolkit-go/engine/cli"
)

func main() {
	root := &cli.Command{
		Usage: "duplex",
		Long:  `duplex is a utility for working with the duplex protocol stack`,
	}

	root.AddCommand(callCmd)
	root.AddCommand(interopCmd)
	root.AddCommand(checkCmd)
	root.AddCommand(benchCmd)

	if err := cli.Execute(context.Background(), root, os.Args[1:]); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
