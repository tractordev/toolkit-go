package daemon_test

import (
	"context"
	"testing"
	"time"

	"tractor.dev/toolkit-go/engine"
	"tractor.dev/toolkit-go/engine/daemon"
)

type initService struct {
	used bool
}

func (s *initService) InitializeDaemon() error {
	s.used = true
	return nil
}

type termService struct {
	used bool
}

func (s *termService) TerminateDaemon(ctx context.Context) error {
	s.used = true
	return nil
}

type simpleService struct {
	used bool
}

func (s *simpleService) Serve(ctx context.Context) {
	s.used = true
	return
}

func fatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestDaemon(t *testing.T) {
	s1 := new(initService)
	s2 := new(simpleService)
	s3 := new(termService)

	r, _ := engine.New()
	fatal(t, r.Add(s1, s2, s3))

	d := &daemon.Framework{}
	r.Assemble(d)

	if len(d.Initializers) != 1 {
		t.Fatal("initializer not registered")
	}
	if len(d.Terminators) != 1 {
		t.Fatal("terminator not registered")
	}
	if len(d.Services) != 1 {
		t.Fatal("service not registered")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	fatal(t, d.Run(ctx))

	if !s1.used {
		t.Fatal("init not used")
	}
	if !s2.used {
		t.Fatal("service not used")
	}
	if !s3.used {
		t.Fatal("terminator not used")
	}
}
