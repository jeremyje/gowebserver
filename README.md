Go Web Server
=============

A simple standalone web server that can quickly host static files on a file system.


Example command to host your home directory.

```
./server --port=8080 --secure_port=8443 --directory=${HOME}
```

Downloads
---------

    OS    | Arch  | Link
----------|-------|--------------------------------------------------------------------------------
Linux     | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.1/server-amd64
Linux     | arm   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.1/server-arm
Windows   | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.1/server-amd64.exe


Build
-----

Install [Go 1.5+](https://golang.org/dl/).

```
go build server.go
```
