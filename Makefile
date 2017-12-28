prefix = /usr
bindir = $(prefix)/bin
sharedir = $(prefix)/share
mandir = $(sharedir)/man
man1dir = $(mandir)/man1
GO := @go
GOGET := go get -u
GOGETBUILD := go get -u
SOURCE_DIRS=$(shell go list ./... | grep -v '/vendor/')
export PATH := $(PATH):/usr/local/go/bin:/usr/go/bin
BINARY_NAME=gowebserver
MAN_PAGE_NAME=${BINARY_NAME}.1
BINARY_MAIN=gowebserver.go
GOAPP := $(shell command -v go 2> /dev/null)

build: gowebserver
all: gowebserver extended-platforms main-platforms

install-go:
ifndef GOAPP
	snap install --classic go
endif

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
	@tar -zcf release.tar.gz release/ 

gowebserver-%: GOOS = $(shell echo $@ | sed 's/.*-\(.*\)-.*/\1/')
gowebserver-%: GOARCH = $(shell echo $@ | sed 's/.*-\(.*\)/\1/')
gowebserver-%: BINARY_SUFFIX = ${GOOS}-${GOARCH}
gowebserver-%:
	@cp "$(BINARY_NAME).go" "$(BINARY_NAME)-$(BINARY_SUFFIX).go"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) GO15VENDOREXPERIMENT=1 go build "$(BINARY_NAME)-$(BINARY_SUFFIX).go"
	@rm "$(BINARY_NAME)-$(BINARY_SUFFIX).go"

gowebserver: embedded/bindata_assetfs.go
	$(GO) build ${BINARY_MAIN}

lint:
	$(GO) fmt ${SOURCE_DIRS}
	$(GO) vet ${SOURCE_DIRS}
	@golint ${SOURCE_DIRS}
	@gocyclo -over 10 .

clean:
	@rm -f ${BINARY_NAME} ${BINARY_NAME}-* cert.pem rsa.pem release.tar.gz testing/*.zip testing/*.tar* testing/testassets.go *.tar.bz2 *.snap
	@rm -rf release/
	@rm -rf parts/ prime/ snap/.snapcraft/ stage/ *.snap
	@rm -f embedded/bindata_assetfs.go
	@rm -rf upload/
	@bazel clean

check: test

testing/testassets.zip:
	@zip -qr9 testing/testassets.zip testing/*

testing/testassets.tar.gz:
	@cd testing/testassets/; GZIP=-9 tar czf ../testassets.tar.gz *
	
testing/testassets.tar.bz2:
	@cd testing/testassets/; BZIP=-9 tar cjf ../testassets.tar.bz2 *
	
testing/testassets.tar:
	@cd testing/testassets/; tar cf ../testassets.tar *

testing/testassets.go: testing
	@echo "package testing" > testing/testassets.go
	@echo "const zipAssets=\"$(shell base64 -w0 testing/testassets.zip)\"" >> testing/testassets.go
	@echo "const tarAssets=\"$(shell base64 -w0 testing/testassets.tar)\"" >> testing/testassets.go
	@echo "const tarGzAssets=\"$(shell base64 -w0 testing/testassets.tar.gz)\"" >> testing/testassets.go
	@echo "const tarBzip2Assets=\"$(shell base64 -w0 testing/testassets.tar.bz2)\"" >> testing/testassets.go
	@gofmt -s -w ./testing/

testing: testing/testassets.zip testing/testassets.tar.gz testing/testassets.tar.bz2 testing/testassets.tar embedded/bindata_assetfs.go

test: testing/testassets.go
	$(GO) test -race ${SOURCE_DIRS}
	
test-10: testing/testassets.go
	$(GO) test -race ${SOURCE_DIRS} -count 10

coverage: testing/testassets.go
	$(GO) test -cover ${SOURCE_DIRS}

bench: benchmark

benchmark: testing/testassets.go
	$(GO) test -benchmem -bench=. ${SOURCE_DIRS}

test-all: test test-10 benchmark coverage

package-legacy:
	@snapcraft

package:
	@LC_ALL=C.UTF-8 LANG=C.UTF-8 snapcraft

run: clean gowebserver lint
	@go run gowebserver.go

install: gowebserver
	@mkdir -p $(DESTDIR)$(bindir) $(DESTDIR)$(man1dir)
	@install ${BINARY_NAME} $(DESTDIR)$(bindir)
	@install -m 0644 ${MAN_PAGE_NAME} $(DESTDIR)$(man1dir)

deps:
	$(GOGET) gopkg.in/yaml.v2
	$(GOGET) github.com/prometheus/client_golang/prometheus
	$(GOGET) github.com/rs/cors
	$(GOGET) github.com/stretchr/testify/assert
	$(GOGET) gopkg.in/src-d/go-git.v4
	# Resources
	$(GOGETBUILD) github.com/jteeuwen/go-bindata/...
	$(GOGETBUILD) github.com/elazarl/go-bindata-assetfs/...

tools:
	$(GOGET) golang.org/x/tools/cmd/gorename
	$(GOGET) github.com/golang/lint/golint
	$(GOGET) github.com/nsf/gocode
	$(GOGET) github.com/rogpeppe/godef
	$(GOGET) github.com/lukehoban/go-outline
	$(GOGET) github.com/newhook/go-symbols
	$(GOGET) github.com/sqs/goreturns
	$(GOGET) github.com/fzipp/gocyclo

embedded/bindata_assetfs.go:
	@rm -f embedded/bindata_assetfs.go
	@cd embedded; go-bindata-assetfs -pkg embedded *

.PHONY : all main-platforms extended-platforms dist build lint clean check testdata testing test test-10 coverage bench benchmark test-all package-legacy package install run deps tools install-go
