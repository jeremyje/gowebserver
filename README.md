# Go Web Server

A simple, convenient, reliable, well tested HTTP/HTTPS web server to host static files.
It can host a local directory or contents of a zip file.

```bash
# Download (linux amd64, see Downloads for other builds)
curl -o gowebserver -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.5.4/server-amd64; chmod +x gowebserver

# Host the current directory.
./gowebserver

# Host your home directory.
./gowebserver --path=${HOME}

# Host a zip file from the internet.
./gowebserver --path=https://github.com/jeremyje/gowebserver/archive/v2.5.4.zip

# Install in your Kubernetes Cluster.
kubectl apply -f https://raw.githubusercontent.com/jeremyje/gowebserver/main/install/kubernetes.yaml
```

## Windows Service

```powershell
sc.exe create gowebserver start= delayed-auto binpath= "C:\apps\gowebserver.exe -configfile=C:\apps\gowebserver.yaml"
sc.exe failure gowebserver reset= 0 actions= restart/1000
sc.exe start gowebserver
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

|   OS   | Arch  | Link
|--------|-------|------------------------------------------------------------------------------------------------------------------------------------------------------
|Linux   | amd64 | `curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.5.4/server-amd64`
|Linux   | arm   | `curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.5.4/server-arm`
|Linux   | arm64 | `curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.5.4/server-arm64`
|Windows | amd64 | `(New-Object System.Net.WebClient).DownloadFile("https://github.com/jeremyje/gowebserver/releases/download/v2.5.4/server-amd64.exe", "server-amd64.exe")`
|macOS   | amd64 | `curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.5.4/server-amd64-darwin`
|macOS   | arm64 | `curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v2.5.4/server-arm64-darwin`

## Build

![example workflow](https://github.com/jeremyje/gowebserver/actions/workflows/deploy.yml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/jeremyje/gowebserver)](https://goreportcard.com/report/github.com/jeremyje/gowebserver) [![Go Reference](https://pkg.go.dev/badge/github.com/jeremyje/gowebserver.svg)](https://pkg.go.dev/github.com/jeremyje/gowebserver) [![codebeat badge](https://codebeat.co/badges/55274aa8-2846-40d2-96c1-f0c9175534ae)](https://codebeat.co/projects/github-com-jeremyje-gowebserver-main) [![codecov](https://codecov.io/gh/jeremyje/gowebserver/branch/main/graph/badge.svg)](https://codecov.io/gh/jeremyje/gowebserver) [![Total alerts](https://img.shields.io/lgtm/alerts/g/jeremyje/gowebserver.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/jeremyje/gowebserver/alerts/)

Install [Go 1.19 or newer](https://golang.org/dl/).

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
