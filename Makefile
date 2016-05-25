prefix = /usr/local
bindir = $(prefix)/bin
sharedir = $(prefix)/share
mandir = $(sharedir)/man
man1dir = $(mandir)/man1
export PATH := $(PATH):/usr/local/go/bin
BINARY_NAME=gowebserver

all: gowebserver

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
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build "$(BINARY_NAME)-$(BINARY_SUFFIX).go"
	@rm "$(BINARY_NAME)-$(BINARY_SUFFIX).go"

gowebserver:
	@go build gowebserver.go

lint:
	@go fmt gowebserver.go
	@go vet gowebserver.go

clean:
	@rm -f gowebserver gowebserver-* cert.pem rsa.pem
	@rm -rf release/

deps:
	@go get -u github.com/prometheus/client_golang/...

install: all
	@install gowebserver $(DESTDIR)$(bindir)
	@install -m 0644 gowebserver.1 $(DESTDIR)$(man1dir)

.PHONY : main-platforms extended-platforms dist lint deps
