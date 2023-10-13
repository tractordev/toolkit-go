module github.com/tractordev/toolkit-go/duplex/cmd/duplex

go 1.21.1

require (
	github.com/progrium/clon-go v0.0.0-20221124010328-fe21965c77cb
	tractor.dev/toolkit-go v0.0.0-00010101000000-000000000000
	tractor.dev/toolkit-go/duplex/x/quic v0.0.0-00010101000000-000000000000
)

require (
	github.com/fxamacker/cbor/v2 v2.5.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/quic-go/qtls-go1-20 v0.3.4 // indirect
	github.com/quic-go/quic-go v0.39.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/mock v0.3.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
)

replace (
	tractor.dev/toolkit-go => ../../..
	tractor.dev/toolkit-go/duplex/x/quic => ../../../duplex/x/quic
)
