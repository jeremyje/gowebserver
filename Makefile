# Copyright 2019 Jeremy Edwards
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

prefix = /usr
bindir = $(prefix)/bin
sharedir = $(prefix)/share
mandir = $(sharedir)/man
man1dir = $(mandir)/man1
GO := @GO111MODULE=on go
GOGET := @GO111MODULE=on go get -u
GOGETBUILD := @GO111MODULE=on go get -u
SOURCE_DIRS=$(shell go list ./... | grep -v '/vendor/')
export PATH := $(PWD)/toolchain:$(PATH):/root/go/bin:/usr/lib/go-1.9/bin:/usr/local/go/bin:/usr/go/bin
BINARY_NAME=gowebserver
MAN_PAGE_NAME=${BINARY_NAME}.1
BINARY_MAIN=gowebserver.go
GOAPP := $(shell command -v go 2> /dev/null)
REPOSITORY_ROOT := $(patsubst %/,%,$(dir $(abspath Makefile)))
TOOLCHAIN_DIR = $(REPOSITORY_ROOT)/toolchain
GOBINDATA = $(TOOLCHAIN_DIR)/go-bindata
GOBINDATA_ASSETFS = $(TOOLCHAIN_DIR)/go-bindata-assetfs

build: gowebserver
all: gowebserver extended-platforms main-platforms

install-go:
ifndef GOAPP
	sudo apt update
	sudo apt install -y software-properties-common python-software-properties
	sudo add-apt-repository ppa:gophers/archive
	sudo apt update
	sudo apt install -y golang-1.9-go
endif

main-platforms: gowebserver-linux-386 gowebserver-linux-amd64 gowebserver-linux-arm gowebserver-windows-386.exe gowebserver-windows-amd64.exe gowebserver-linux-arm64 gowebserver-darwin-amd64
extended-platforms: gowebserver-netbsd-amd64 gowebserver-openbsd-amd64 gowebserver-freebsd-amd64 gowebserver-dragonfly-amd64

dist: release/
release/: all
	@mkdir -p release/
	@mv gowebserver-* release/
	@mv release/gowebserver-linux-386 release/server-386
	@mv release/gowebserver-linux-arm release/server-arm
	@mv release/gowebserver-linux-amd64 release/server-amd64
	@mv release/gowebserver-linux-arm64 release/server-arm64
	@mv release/gowebserver-darwin-amd64 release/server-darwin-amd64
	@mv release/gowebserver-windows-386.exe release/server-386.exe
	@mv release/gowebserver-windows-amd64.exe release/server-amd64.exe
	@tar -zcf release.tar.gz release/ 

gowebserver-%: GOOS = $(shell echo $@ | cut -d'-' -f2 | cut -d'.' -f1)
gowebserver-%: GOARCH = $(shell echo $@ | cut -d'-' -f3 | cut -d'.' -f1)
gowebserver-%: BINARY_SUFFIX = $(GOOS)-$(GOARCH)
gowebserver-%:
	@GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=on go build -o $@ gowebserver.go

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
	@rm -rf toolchain/

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
	
coverage.txt: testing/testassets.go
	@for sfile in ${SOURCE_DIRS} ; do \
		go test -race "$$sfile" -coverprofile=package.coverage -covermode=atomic; \
		if [ -f package.coverage ]; then \
			cat package.coverage >> coverage.txt; \
			rm package.coverage; \
		fi; \
	done
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

sync-deps:
	@go mod tidy

deps:
	@go mod download

toolchain/:
	mkdir -p $(TOOLCHAIN_DIR)
	GO111MODULE=on GOPROXY= go build -o $(GOBINDATA) github.com/kevinburke/go-bindata/go-bindata
	GO111MODULE=on GOPROXY= go build -o $(GOBINDATA_ASSETFS) github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs

tools:
	$(GOGET) golang.org/x/tools/cmd/gorename
	$(GOGET) golang.org/x/lint
	$(GOGET) github.com/nsf/gocode
	$(GOGET) github.com/rogpeppe/godef
	$(GOGET) github.com/lukehoban/go-outline
	$(GOGET) github.com/newhook/go-symbols
	$(GOGET) github.com/sqs/goreturns
	$(GOGET) github.com/fzipp/gocyclo

embedded/bindata_assetfs.go: toolchain/
	@rm -f embedded/bindata_assetfs.go
	@cd embedded; $(GOBINDATA_ASSETFS) -pkg embedded *

.PHONY : all main-platforms extended-platforms dist build lint clean check testdata testing test test-10 coverage bench benchmark test-all package-legacy package install run deps tools install-go
