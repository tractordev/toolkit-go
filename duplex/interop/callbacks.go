package interop

import (
	"io"
	"log"

	"tractor.dev/toolkit-go/duplex/rpc"
)

type CallbackService struct{}

func (s CallbackService) UnaryCallback(resp rpc.Responder, call *rpc.Call) {
	var params any
	if err := call.Receive(&params); err != nil {
		log.Println(err)
		return
	}
	if err := resp.Return(params); err != nil {
		log.Println(err)
	}
}

func (s CallbackService) StreamCallback(resp rpc.Responder, call *rpc.Call) {
	var v any
	if err := call.Receive(&v); err != nil {
		log.Println(err)
		return
	}
	ch, err := resp.Continue(v)
	if err != nil {
		log.Println(err)
		return
	}
	defer ch.Close()
	for err == nil {
		err = call.Receive(&v)
		if err == nil {
			err = resp.Send(v)
		}
	}
	if err != nil && err != io.EOF {
		log.Println(err)
	}
}

func (s CallbackService) BytesCallback(resp rpc.Responder, call *rpc.Call) {
	var params any
	if err := call.Receive(&params); err != nil {
		log.Println(err)
		return
	}
	ch, err := resp.Continue(params)
	if err != nil {
		log.Println(err)
		return
	}
	defer ch.Close()
	if _, err := io.Copy(ch, call); err != nil {
		log.Println(err)
	}
}
