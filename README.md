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
Linux     | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/1.4/server-amd64
Linux     | arm   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/1.4/server-arm
Linux     | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/1.4/server-386
Windows   | amd64 | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/1.4/server-amd64.exe
Windows   | 386   | curl -O -L https://github.com/jeremyje/gowebserver/releases/download/1.4/server-386.exe


Build
-----

Status: [![Build Status](https://secure.travis-ci.org/jeremyje/gowebserver.png)](http://travis-ci.org/jeremyje/gowebserver)

Install [Go 1.5+](https://golang.org/dl/).

```
git clone git@github.com:jeremyje/gowebserver.git --recursive
make
```
