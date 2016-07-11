prefix = /usr
bindir = $(prefix)/bin
sharedir = $(prefix)/share
mandir = $(sharedir)/man
man1dir = $(mandir)/man1
GO := @GO15VENDOREXPERIMENT=1 go
export PATH := $(PATH):/usr/local/go/bin:/usr/go/bin
BINARY_NAME=gowebserver

build: gowebserver
all: gowebserver extended-platforms main-platforms

main-platforms: gowebserver-linux-386 gowebserver-linux-amd64 gowebserver-linux-arm gowebserver-windows-386 gowebserver-windows-amd64
extended-platforms: gowebserver-linux-arm64 gowebserver-darwin-amd64 gowebserver-netbsd-amd64 gowebserver-openbsd-amd64 gowebserver-freebsd-amd64 gowebserver-dragonfly-amd64

dist: main-platforms
	@mkdir -p release/
	@mv gowebserver-* release/
	@mv release/gowebserver-linux-386 release/server-386
	@mv release/gowebserver-linux-arm release/server-arm
	@mv release/gowebserver-linux-amd64 release/server-amd64
	@mv release/gowebserver-windows-386.exe release/server-386.exe
	@mv release/gowebserver-windows-amd64.exe release/server-amd64.exe

gowebserver-%: GOOS = $(shell echo $@ | sed 's/.*-\(.*\)-.*/\1/')
gowebserver-%: GOARCH = $(shell echo $@ | sed 's/.*-\(.*\)/\1/')
gowebserver-%: BINARY_SUFFIX = ${GOOS}-${GOARCH}
gowebserver-%:
	@cp "$(BINARY_NAME).go" "$(BINARY_NAME)-$(BINARY_SUFFIX).go"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) GO15VENDOREXPERIMENT=1 go build "$(BINARY_NAME)-$(BINARY_SUFFIX).go"
	@rm "$(BINARY_NAME)-$(BINARY_SUFFIX).go"

gowebserver:
	$(GO) build gowebserver.go

lint:
	$(GO) fmt gowebserver.go
	$(GO) vet gowebserver.go

clean:
	@rm -f gowebserver gowebserver-* cert.pem rsa.pem
	@rm -rf release/

check: test

test:
	$(GO) test github.com/jeremyje/gowebserver/cert
	$(GO) test github.com/jeremyje/gowebserver/config
	$(GO) test github.com/jeremyje/gowebserver/server

package:
	@cd packaging
	@snapcraft
	@cd ..

install: all
	@install gowebserver $(DESTDIR)$(bindir)
	@install -m 0644 gowebserver.1 $(DESTDIR)$(man1dir)

.PHONY : all main-platforms extended-platforms dist build lint clean check test package install 
