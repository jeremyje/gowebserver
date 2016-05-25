Go Web Server
=============

A simple web server to host static files on a filesystem via HTTP and HTTPS with no configuration. 

Example command to host your home directory.

```
./gowebserver --directory=${HOME}
```

Downloads
---------

    OS    | Arch  | Link
----------|-------|-------------------------------------------------------------------------------------------
Linux     | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.3/server-amd64
Linux     | arm   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.3/server-arm
Linux     | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.3/server-386
Windows   | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.3/server-amd64.exe
Windows   | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/v1.3/server-386.exe


Build
-----

Install [Go 1.5+](https://golang.org/dl/).

```
make deps
make gowebserver
```
