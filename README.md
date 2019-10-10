# Go Web Server

A simple, convenient, reliable, well tested HTTP/HTTPS web server to host static files.
It can host a local directory or contents of a zip file.

```
# Download (linux amd64, see Downloads for other builds)
curl -o gowebserver -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.9.0/server-amd64; chmod +x gowebserver

# Host the directory you're currently in.
./gowebserver

# Host your home directory.
./gowebserver --path=${HOME}

# Host a zip file from the internet.
./gowebserver --path=https://github.com/jeremyje/gowebserver/archive/v1.9.0.zip
```

## Features
 * Zero-config required, hosts on port 80 or 8080 based on root and supports Cloud9's $PORT variable.
 * HTTP and HTTPs serving
 * Automatic HTTPs certificate generation
 * Optional configuration by flags or YAML config file.
 * Host static files from:
   * Local directory (current directory is default)
   * ZIP archive (local or from HTTP/HTTPS)
   * Tarball (.tar, .tar.gz, .tar.bz2) archive
   * Git repository (HTTPS, SSH)
 * Metrics export to Prometheus.
 * Prebuild binaries for all major OSes.
 * Ubuntu snappy packaging for Raspberry Pi and other IoT devices.


## Downloads

|    OS    | Arch  | Link
|----------|-------|-------------------------------------------------------------------------------------------
|Linux     | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.9.0/server-amd64
|Linux     | arm   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.9.0/server-arm
|Linux     | arm64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.9.0/server-arm64
|Linux     | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.9.0/server-386
|Windows   | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.9.0/server-amd64.exe
|Windows   | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.9.0/server-386.exe
|macOS     | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.9.0/server-amd64-darwin


## Build

Status: [![Build Status](https://secure.travis-ci.org/jeremyje/gowebserver.png)](http://travis-ci.org/jeremyje/gowebserver) [![Go Report Card](https://goreportcard.com/badge/github.com/jeremyje/gowebserver)](https://goreportcard.com/report/github.com/jeremyje/gowebserver) [![GoDoc](https://godoc.org/github.com/jeremyje/gowebserver?status.svg)](https://godoc.org/github.com/jeremyje/gowebserver) [![Snap Status](https://build.snapcraft.io/badge/jeremyje/gowebserver.svg)](https://build.snapcraft.io/user/jeremyje/gowebserver) [![codebeat badge](https://codebeat.co/badges/de86a882-9038-4994-afe2-fea7d93f63cb)](https://codebeat.co/projects/github-com-jeremyje-gowebserver-master) [![codecov](https://codecov.io/gh/jeremyje/gowebserver/branch/master/graph/badge.svg)](https://codecov.io/gh/jeremyje/gowebserver)

Install [Go 1.9 or newer](https://golang.org/dl/).

```bash
git clone git@github.com:jeremyje/gowebserver.git --recursive
make

OR

go build gowebserver
```

Test
----

```bash
make test
make bench
```

Sample
------
Sample code for embedding a HTTP/HTTPS server in your application.

```go
import (
	"github.com/jeremyje/gowebserver/server"
	"github.com/jeremyje/gowebserver/cert"
)
func main() {
	certBuilder := cert.NewCertificateBuilder().
		SetRsa2048().
		SetValidDurationInDays(365)
	checkError(certBuilder.WriteCertificate("public.cert"))
	checkError(certBuilder.WritePrivateKey("private.key"))

	httpServer := server.NewWebServer().
		SetPorts(80, 443).
		SetMetricsEnabled(true).
		SetServePath("/", "/metrics").
		SetCertificateFile("public.cert").
		SetPrivateKey("private.key").
		SetVerbose(true)
	checkError(httpServer.SetPath("."))
	checkError(httpServer.SetUpload("./upload", "/upload.html"))
	httpServer.Serve()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
```
