# gotell

> Golang OTEL library, an opinionated thin-wrapper over the OpenTelemetry SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/wwmoraes/gotell.svg)](https://pkg.go.dev/github.com/wwmoraes/gotell)
[![GitHub Issues](https://img.shields.io/github/issues/wwmoraes/gotell.svg)](https://github.com/wwmoraes/gotell/issues)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/wwmoraes/gotell.svg)](https://github.com/wwmoraes/gotell/pulls)
![Codecov](https://img.shields.io/codecov/c/github/wwmoraes/gotell)

![GitHub branch status](https://img.shields.io/github/checks-status/wwmoraes/gotell/main)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](/LICENSE)

---

## 📝 Table of Contents

- [About](#-about)
- [Getting Started](#-getting-started)
- [Usage](#-usage)
- [Built Using](#-built-using)
- [TODO](./TODO.md)
- [Contributing](./CONTRIBUTING.md)
- [Authors](#-authors)
- [Acknowledgments](#-acknowledgements)

## 🧐 About

Gotell is an opinionated thin wrapper on top of the OpenTelemetry Go SDK. It
applies semantic convention metrics and attributes for you so you don't have to.

The name gotell comes from `go` + `otel` (OpenTelemetry) + `library`. 😉

## 🏁 Getting Started

Run `go get github.com/wwmoraes/gotell` or add the repository to your imports
then run `go mod tidy`.

## 🔧 Running the tests

Run `task test` 😄

## 🎈 Usage

```go
package main

import (
  "context"
  "errors"
  "fmt"
  "net"
  "net/http"
  "os"
  "os/signal"
  "time"

  telemetry "github.com/wwmoraes/gotell"
  "go.opentelemetry.io/otel/attribute"
  "go.opentelemetry.io/otel/sdk/resource"
)

const (
  NAMESPACE = "github.com/wwmoraes/gotell"
  NAME      = "http-server"
)

var version = "0.0.0"

func main() {
  ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
  defer cancel()

  hostname, err := os.Hostname()
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  err = telemetry.Initialize(ctx, resource.NewSchemaless(
    attribute.String("service.name", NAME),
    attribute.String("service.namespace", NAMESPACE),
    attribute.String("service.version", version),
    attribute.String("host.id", hostname),
  ))
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  log := telemetry.Logr(ctx)

  mux := http.NewServeMux()

  mux.Handle("/", telemetry.WithInstrumentationMiddleware(http.HandlerFunc(helloHandler)))

  listener, err := net.Listen("tcp4", "127.0.0.1:0")
  if !errors.Is(err, http.ErrServerClosed) {
    log.Error(err, "failed to create listener")
    os.Exit(1)
  }

  server := http.Server{
    Handler:           mux,
    ReadTimeout:       time.Minute / 2,
    ReadHeaderTimeout: time.Minute / 2,
  }

  log.Info("serving HTTP", "address", listener.Addr().String())

  err = server.Serve(listener)
  if !errors.Is(err, http.ErrServerClosed) {
    log.Error(err, "server stopped")
  }
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
  ctx, span := telemetry.Start(r.Context())
  defer span.End()

  log := telemetry.Logr(ctx)

  log.Info("hello handler triggered")

  name := r.URL.Query().Get("name")
  if name == "" {
    name = "stranger"
  }

  span.SetAttributes(attribute.String("name", name))

  fmt.Fprintf(w, "hello %s!", name)
}
```

## 🔧 Built Using

- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/) - Core
telemetry engine
- [Go](https://go.dev) - Programming language

## 🧑‍💻 Authors

- [@wwmoraes](https://github.com/wwmoraes) - Idea & Initial work

## 🎉 Acknowledgements

- OpenTelemetry <https://opentelemetry.io>
