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
Linux     | amd64 | curl -O https://storage.googleapis.com/gowebserver/pub/linux-amd64/server
Linux     | 386   | curl -O https://storage.googleapis.com/gowebserver/pub/linux-386/server
Linux     | arm   | curl -O https://storage.googleapis.com/gowebserver/pub/linux-arm/server
Linux     | arm64 | curl -O https://storage.googleapis.com/gowebserver/pub/linux-arm64/server
Windows   | amd64 | curl -O https://storage.googleapis.com/gowebserver/pub/windows-amd64/server.exe
Windows   | 386   | curl -O https://storage.googleapis.com/gowebserver/pub/windows-386/server.exe
Darwin    | amd64 | curl -O https://storage.googleapis.com/gowebserver/pub/darwin-amd64/server
NetBSD    | amd64 | curl -O https://storage.googleapis.com/gowebserver/pub/netbsd-amd64/server
OpenBSD   | amd64 | curl -O https://storage.googleapis.com/gowebserver/pub/openbsd-amd64/server
FreeBSD   | amd64 | curl -O https://storage.googleapis.com/gowebserver/pub/freebsd-amd64/server
Dragonfly | amd64 | curl -O https://storage.googleapis.com/gowebserver/pub/dragonfly-amd64/server
