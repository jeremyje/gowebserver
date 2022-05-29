# Go Web Server

A simple, convenient, reliable, well tested HTTP/HTTPS web server to host static files.
It can host a local directory or contents of a zip file.

```bash
# Download (linux amd64, see Downloads for other builds)
curl -o gowebserver -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.4.0/server-amd64; chmod +x gowebserver

# Host the current directory.
./gowebserver

# Host your home directory.
./gowebserver --path=${HOME}

# Host a zip file from the internet.
./gowebserver --path=https://github.com/jeremyje/gowebserver/archive/v2.4.0.zip
```

## Features

* Zero-config required, hosts on port 80 or 8080 based on root and supports Cloud9's $PORT variable.
* HTTP and HTTPs serving
* Automatic HTTPs certificate generation
* Optional configuration by flags or YAML config file.
* Host local or HTTP served static files from:
  * Local directory (current directory is default)
  * ZIP archive
  * Tarball archive (.tar, .tar.bz2, .tar.gz, .tar.lz4, .tar.xz)
  * 7-zip
  * RAR
  * Git repository (HTTPS, SSH)
* Metrics export to Prometheus.
* Prebuild binaries for all major OSes.

## Downloads

|    OS    | Arch  | Link
|----------|-------|-------------------------------------------------------------------------------------------
|Linux     | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.4.0/server-amd64
|Linux     | arm   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.4.0/server-arm
|Linux     | arm64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.4.0/server-arm64
|Linux     | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.4.0/server-386
|Windows   | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.4.0/server-amd64.exe
|Windows   | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.4.0/server-386.exe
|macOS     | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.4.0/server-amd64-darwin

## Build

![example workflow](https://github.com/jeremyje/gowebserver/actions/workflows/deploy.yml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/jeremyje/gowebserver)](https://goreportcard.com/report/github.com/jeremyje/gowebserver) [![Go Reference](https://pkg.go.dev/badge/github.com/jeremyje/gowebserver.svg)](https://pkg.go.dev/github.com/jeremyje/gowebserver) [![codebeat badge](https://codebeat.co/badges/de86a882-9038-4994-afe2-fea7d93f63cb)](https://codebeat.co/projects/github-com-jeremyje-gowebserver-master) [![codecov](https://codecov.io/gh/jeremyje/gowebserver/branch/master/graph/badge.svg)](https://codecov.io/gh/jeremyje/gowebserver) [![Total alerts](https://img.shields.io/lgtm/alerts/g/jeremyje/gowebserver.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/jeremyje/gowebserver/alerts/)

Install [Go 1.18 or newer](https://golang.org/dl/).

```bash
git clone git@github.com:jeremyje/gowebserver.git
make
```

## Test

```bash
make test
make bench
```

## Sample

Sample code for embedding a HTTP/HTTPS server in your application.

```go
package main

import (
  "github.com/jeremyje/gowebserver/pkg/gowebserver"
  "go.uber.org/zap"
)

func main() {
  logger, err := zap.NewProduction()
  if err != nil {
    zap.S().Fatal(err)
  }
  if err == nil {
    zap.ReplaceGlobals(logger)
  }
  defer logger.Sync()
  httpServer, err := gowebserver.New(&gowebserver.Config{
    Serve: []gowebserver.Serve{{Source: ".", Endpoint: "/"}},
  })
  if err != nil {
    zap.S().Fatal(err)
  }

  termCh := make(chan error)
  httpServer.Serve(termCh)
}

```
