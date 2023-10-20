package cli

import (
	"context"
	"io"
)

type Context struct {
	context.Context
	*iocontext
}

// ContextWithIO returns a child context with a ContextIO
// value added using the given Stdio equivalents.
func ContextWithIO(parent context.Context, in io.Reader, out io.Writer, err io.Writer) *Context {
	return &Context{
		Context: parent,
		iocontext: &iocontext{
			in:  in,
			out: out,
			err: err,
		},
	}
}

type iocontext struct {
	out, err io.Writer
	in       io.Reader
}

func (c *iocontext) Write(p []byte) (n int, err error) {
	return c.out.Write(p)
}

func (c *iocontext) Read(p []byte) (n int, err error) {
	return c.in.Read(p)
}

func (c *iocontext) Errout() io.Writer {
	return c.err
}
