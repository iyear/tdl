module github.com/iyear/tdl/extension

go 1.21
toolchain go1.23.4

replace github.com/iyear/tdl/core => ../core

require (
	github.com/go-faster/errors v0.7.1
	github.com/gotd/td v0.116.0
	github.com/iyear/tdl/core v0.18.3
	go.uber.org/zap v1.27.0
)

require (
	github.com/beevik/ntp v1.3.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/coder/websocket v1.8.12 // indirect
	github.com/go-faster/jx v1.1.0 // indirect
	github.com/go-faster/xor v1.0.0 // indirect
	github.com/gotd/contrib v0.20.0 // indirect
	github.com/gotd/ige v0.2.2 // indirect
	github.com/gotd/neo v0.1.5 // indirect
	github.com/iyear/connectproxy v0.1.1 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.30.0 // indirect
	golang.org/x/net v0.32.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	rsc.io/qr v0.2.0 // indirect
)
