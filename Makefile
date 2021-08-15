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

include golang.mk
include docker.mk

prefix = /usr
bindir = $(prefix)/bin
sharedir = $(prefix)/share
mandir = $(sharedir)/man
man1dir = $(mandir)/man1

RM = rm
ZIP = zip
TAR = tar
ECHO = @echo

BASE_VERSION = 0.0.0-dev
SHORT_SHA = $(shell git rev-parse --short=7 HEAD | tr -d [:punct:])
VERSION_SUFFIX = $(SHORT_SHA)
VERSION = $(BASE_VERSION)-$(VERSION_SUFFIX)
BUILD_DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
TAG := $(VERSION)

SOURCE_DIRS=$(shell go list ./... | grep -v '/vendor/')
export PATH := $(PWD)/bin/toolchain:$(PATH):/root/go/bin:/usr/lib/go-1.9/bin:/usr/local/go/bin:/usr/go/bin
BINARY_NAME=gowebserver
MAN_PAGE_NAME=${BINARY_NAME}.1
REPOSITORY_ROOT := $(patsubst %/,%,$(dir $(abspath Makefile)))

REGISTRY = docker.io/jeremyje
GOWEBSERVER_IMAGE = $(REGISTRY)/gowebserver

NICHE_PLATFORMS = freebsd openbsd netbsd darwin

LINUX_PLATFORMS = linux_386 linux_amd64 linux_arm_v5 linux_arm_v6 linux_arm_v7 linux_arm64 linux_s390x linux_ppc64le linux_riscv64 linux_mips64le linux_mips linux_mipsle linux_mips64
LINUX_NICHE_PLATFORMS = 
WINDOWS_PLATFORMS = windows_386 windows_amd64
ALL_PLATFORMS = $(LINUX_PLATFORMS) $(LINUX_NICHE_PLATFORMS) $(WINDOWS_PLATFORMS) $(foreach niche,$(NICHE_PLATFORMS),$(niche)_amd64 $(niche)_arm64)
ASSETS = embedded/bindata_assetfs.go
ALL_APPS = gowebserver
ALL_ASSETS = $(ASSETS) testing/testassets.go

ALL_BINARIES = $(foreach app,$(ALL_APPS),$(foreach platform,$(ALL_PLATFORMS),bin/go/$(platform)/$(app)$(if $(findstring windows_,$(platform)),.exe,)))
TEST_ASSETS = testing/testassets.zip testing/testassets.tar.gz testing/testassets.tar.bz2 testing/testassets.tar embedded/bindata_assetfs.go
WINDOWS_VERSIONS = 1809 1903 1909 2004 20H2
BUILDX_BUILDER = buildx-builder
LINUX_CPU_PLATFORMS = amd64 arm64 ppc64le s390x arm/v5 arm/v6 arm/v7
space := $(null) #
comma := ,

bin/go/%: CGO_ENABLED=0
bin/go/%: $(ASSETS)

.SECONDEXPANSION:

all: $(ALL_BINARIES)
assets: $(ASSETS)

dist: bin/release.tar.gz

bin/release.tar.gz: $(ALL_BINARIES)
	mkdir -p bin/
	cd bin/go/; $(TAR) -zcf ../release.tar.gz *

lint: $(ALL_ASSETS)
	$(GO) fmt ${SOURCE_DIRS}
	$(GO) vet ${SOURCE_DIRS}

clean:
	$(RM) -f ${BINARY_NAME} ${BINARY_NAME}-* cert.pem rsa.pem release.tar.gz testing/*.zip testing/*.tar* testing/testassets.go *.tar.bz2 *.snap
	$(RM) -rf parts/ prime/ snap/.snapcraft/ stage/ *.snap
	$(RM) -f embedded/bindata_assetfs.go
	$(RM) -rf upload/
	$(RM) -rf toolchain/
	$(RM) -rf bin/

check: test

testing/testassets.zip:
	$(ZIP) -qr9 testing/testassets.zip testing/*

testing/testassets.tar.gz:
	cd testing/testassets/; GZIP=-9 $(TAR) czf ../testassets.tar.gz *
	
testing/testassets.tar.bz2:
	cd testing/testassets/; BZIP=-9 $(TAR) cjf ../testassets.tar.bz2 *
	
testing/testassets.tar:
	cd testing/testassets/; $(TAR) cf ../testassets.tar *

testing/testassets.go: $(TEST_ASSETS)
	$(ECHO) "package testing" > testing/testassets.go
	$(ECHO) "const zipAssets=\"$(shell base64 -w0 testing/testassets.zip)\"" >> testing/testassets.go
	$(ECHO) "const tarAssets=\"$(shell base64 -w0 testing/testassets.tar)\"" >> testing/testassets.go
	$(ECHO) "const tarGzAssets=\"$(shell base64 -w0 testing/testassets.tar.gz)\"" >> testing/testassets.go
	$(ECHO) "const tarBzip2Assets=\"$(shell base64 -w0 testing/testassets.tar.bz2)\"" >> testing/testassets.go
	gofmt -s -w ./testing/

test: testing/testassets.go
	$(GO) test -race ${SOURCE_DIRS}

test-10: testing/testassets.go
	$(GO) test -race ${SOURCE_DIRS} -count 10

coverage: testing/testassets.go
	$(GO) test -cover ${SOURCE_DIRS}

coverage.txt: testing/testassets.go
	for sfile in ${SOURCE_DIRS} ; do \
		go test -race "$$sfile" -coverprofile=package.coverage -covermode=atomic; \
		if [ -f package.coverage ]; then \
			cat package.coverage >> coverage.txt; \
			$(RM) package.coverage; \
		fi; \
	done

bench: benchmark
benchmark: testing/testassets.go
	$(GO) test -benchmem -bench=. ${SOURCE_DIRS}

test-all: test test-10 benchmark coverage

package-legacy:
	snapcraft

package:
	LC_ALL=C.UTF-8 LANG=C.UTF-8 snapcraft

run: clean $(ALL_ASSETS) lint
	$(GO) run cmd/gowebserver/gowebserver.go -http.port 8181

install: gowebserver
	mkdir -p $(DESTDIR)$(bindir) $(DESTDIR)$(man1dir)
	install ${BINARY_NAME} $(DESTDIR)$(bindir)
	install -m 0644 ${MAN_PAGE_NAME} $(DESTDIR)$(man1dir)

deps:
	$(GO) mod tidy
	$(GO) mod download

ensure-builder:
	-$(DOCKER) buildx create --name $(BUILDX_BUILDER)

# https://github.com/docker-library/official-images#architectures-other-than-amd64
images: linux-images windows-images
	-$(DOCKER) manifest rm $(GOWEBSERVER_IMAGE):$(TAG)
	$(DOCKER) manifest create $(GOWEBSERVER_IMAGE):$(TAG) $(foreach winver,$(WINDOWS_VERSIONS),$(GOWEBSERVER_IMAGE):$(TAG)-windows_amd64-$(winver)) $(foreach platform,$(LINUX_PLATFORMS),$(GOWEBSERVER_IMAGE):$(TAG)-$(platform))

	for winver in $(WINDOWS_VERSIONS) ; do \
		windows_version=`$(DOCKER) manifest inspect mcr.microsoft.com/windows/nanoserver:$${winver} | jq -r '.manifests[0].platform["os.version"]'`; \
		$(DOCKER) manifest annotate --os-version $${windows_version} $(GOWEBSERVER_IMAGE):$(TAG) $(GOWEBSERVER_IMAGE):$(TAG)-windows_amd64-$${winver} ; \
	done
	$(DOCKER) manifest push $(GOWEBSERVER_IMAGE):$(TAG)

linux-images: $(foreach platform,$(LINUX_PLATFORMS),linux-image-$(platform))

linux-image-%: bin/go/%/gowebserver ensure-builder
	$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform $(subst _,/,$*) --build-arg BINARY_PATH=$< -f cmd/gowebserver/Dockerfile -t $(GOWEBSERVER_IMAGE):$(TAG)-$* . $(DOCKER_PUSH)

windows-images: $(foreach winver,$(WINDOWS_VERSIONS),windows-image-$(winver))

windows-image-%: bin/go/windows_amd64/gowebserver.exe ensure-builder
	$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform windows/amd64 -f cmd/gowebserver/Dockerfile.windows --build-arg WINDOWS_VERSION=$* -t $(GOWEBSERVER_IMAGE):$(TAG)-windows_amd64-$* . $(DOCKER_PUSH)

.PHONY : all assets dist lint clean check test test-10 coverage bench benchmark test-all package-legacy package install run deps
