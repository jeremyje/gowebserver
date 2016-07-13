Go Web Server
=============

A simple, convenient, reliable, well tested HTTP/HTTPS web server to host static files.
It can host a local directory or contents of a zip file.

Features
--------
 * Zero-config required, hosts on port 80 or 8080 based on root and supports Cloud9's $PORT variable.
 * HTTP and HTTPs serving
 * Automatic HTTPs certificate generation
 * Optional configuration by flags or YAML config file.
 * Host static files from:
   * Local directory (current directory is default)
   * ZIP archive (local or from HTTP/HTTPS)
 * Metrics export to Prometheus.
 * Prebuild binaries for all major OSes.
 * Ubuntu snappy packaging for Raspberry Pi and other IoT devices.

Example command to host your home directory.

```
./gowebserver --directory=${HOME}
```

Downloads
---------

    OS    | Arch  | Link
----------|-------|-------------------------------------------------------------------------------------------
Linux     | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.5/server-amd64
Linux     | arm   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.5/server-arm
Linux     | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.5/server-386
Windows   | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.5/server-amd64.exe
Windows   | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.5/server-386.exe


Build
-----

Status: [![Build Status](https://secure.travis-ci.org/jeremyje/gowebserver.png)](http://travis-ci.org/jeremyje/gowebserver)

Install [Go 1.5+](https://golang.org/dl/).

```
git clone git@github.com:jeremyje/gowebserver.git --recursive
make
```

Test
----

```
make test
make bench
```
