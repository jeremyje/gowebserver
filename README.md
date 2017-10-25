Go Web Server
=============

A simple, convenient, reliable, well tested HTTP/HTTPS web server to host static files.
It can host a local directory or contents of a zip file.

```
# Host the directory you're currently in.
gowebserver

# Host your home directory.
gowebserver --directory=${HOME}

# Host a zip file from the internet.
gowebserver --directory=https://github.com/jeremyje/gowebserver/archive/v1.7.zip
```

Features
--------
 * Zero-config required, hosts on port 80 or 8080 based on root and supports Cloud9's $PORT variable.
 * HTTP and HTTPs serving
 * Automatic HTTPs certificate generation
 * Optional configuration by flags or YAML config file.
 * Host static files from:
   * Local directory (current directory is default)
   * ZIP archive (local or from HTTP/HTTPS)
   * Tarball (.tar, .tar.gz, .tar.bz2) archive
 * Metrics export to Prometheus.
 * Prebuild binaries for all major OSes.
 * Ubuntu snappy packaging for Raspberry Pi and other IoT devices.


Downloads
---------

|    OS    | Arch  | Link
|----------|-------|-------------------------------------------------------------------------------------------
|Linux     | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.6/server-amd64
|Linux     | arm   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.6/server-arm
|Linux     | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.6/server-386
|Windows   | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.6/server-amd64.exe
|Windows   | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.6/server-386.exe


Build
-----

Status: [![Build Status](https://secure.travis-ci.org/jeremyje/gowebserver.png)](http://travis-ci.org/jeremyje/gowebserver) [![Go Report Card](https://goreportcard.com/badge/github.com/jeremyje/gowebserver)](https://goreportcard.com/report/github.com/jeremyje/gowebserver) [![GoDoc](https://godoc.org/github.com/jeremyje/gowebserver?status.svg)](https://godoc.org/github.com/jeremyje/gowebserver)

Install [Go 1.6+](https://golang.org/dl/).

```
git clone git@github.com:jeremyje/gowebserver.git --recursive
make

OR

go build gowebserver
```

Test
----

```
make test
make bench
```

Bazel
-----
Add the following to your WORKSPACE file.

```
go_repository(
    name = "com_github_jeremyje_gowebserver",
    importpath = "github.com/jeremyje/gowebserver",
    commit = "d93a4056c74c2948bec5b805078019aa807c50a9",
)
```