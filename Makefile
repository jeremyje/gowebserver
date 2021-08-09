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

BASE_VERSION = 0.0.0-dev
SHORT_SHA = $(shell git rev-parse --short=7 HEAD | tr -d [:punct:])
VERSION_SUFFIX = $(SHORT_SHA)
VERSION = $(BASE_VERSION)-$(VERSION_SUFFIX)
BUILD_DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
TAG := $(VERSION)

SOURCE_DIRS=$(shell go list ./... | grep -v '/vendor/')
export PATH := $(PWD)/toolchain:$(PATH):/root/go/bin:/usr/lib/go-1.9/bin:/usr/local/go/bin:/usr/go/bin
BINARY_NAME=gowebserver
MAN_PAGE_NAME=${BINARY_NAME}.1
REPOSITORY_ROOT := $(patsubst %/,%,$(dir $(abspath Makefile)))

REGISTRY = gcr.io/jeremyje
GOWEBSERVER_IMAGE = $(REGISTRY)/gowebserver

NICHE_PLATFORMS = freebsd openbsd netbsd darwin

ALL_PLATFORMS = linux_386 linux_amd64 linux_arm_v5 linux_arm_v6 linux_arm_v7 linux_arm64 linux_riscv64 linux_ppc64le linux_mips64le linux_mips linux_mipsle linux_mips64 linux_s390x windows_386 windows_amd64 $(foreach niche,$(NICHE_PLATFORMS),$(niche)_amd64 $(niche)_arm64)
ASSETS = embedded/bindata_assetfs.go
ALL_APPS = gowebserver
ALL_ASSETS = $(ASSETS) testing/testassets.go

ALL_BINARIES = $(foreach app,$(ALL_APPS),$(foreach platform,$(ALL_PLATFORMS),bin/go/$(platform)/$(app)$(if $(findstring windows_,$(platform)),.exe,)))

bin/go/%: CGO_ENABLED=0
bin/go/%: $(ASSETS)

.SECONDEXPANSION:

bin/image-artifacts/%/gowebserver: bin/go/$$(subst /,_,%)/gowebserver
	mkdir -p $(dir $@)
	cp -f $< $@

bin/image-artifacts/%/gowebserver.exe: bin/go/$$(subst /,_,%)/gowebserver.exe
	mkdir -p $(dir $@)
	cp -f $< $@

all: $(ALL_BINARIES)
assets: $(ASSETS)

dist: bin/release.tar.gz

bin/release.tar.gz: $(ALL_BINARIES)
	@mkdir -p bin/release/
	cd bin/; @tar -zcf bin/release.tar.gz go/ 

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
	@zip -qr9 testing/testassets.zip testing/*

testing/testassets.tar.gz:
	@cd testing/testassets/; GZIP=-9 tar czf ../testassets.tar.gz *
	
testing/testassets.tar.bz2:
	@cd testing/testassets/; BZIP=-9 tar cjf ../testassets.tar.bz2 *
	
testing/testassets.tar:
	@cd testing/testassets/; tar cf ../testassets.tar *

TEST_ASSETS = testing/testassets.zip testing/testassets.tar.gz testing/testassets.tar.bz2 testing/testassets.tar embedded/bindata_assetfs.go

testing/testassets.go: $(TEST_ASSETS)
	@echo "package testing" > testing/testassets.go
	@echo "const zipAssets=\"$(shell base64 -w0 testing/testassets.zip)\"" >> testing/testassets.go
	@echo "const tarAssets=\"$(shell base64 -w0 testing/testassets.tar)\"" >> testing/testassets.go
	@echo "const tarGzAssets=\"$(shell base64 -w0 testing/testassets.tar.gz)\"" >> testing/testassets.go
	@echo "const tarBzip2Assets=\"$(shell base64 -w0 testing/testassets.tar.bz2)\"" >> testing/testassets.go
	@gofmt -s -w ./testing/

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
			$(RM) package.coverage; \
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

run: clean $(ALL_ASSETS) lint
	$(GO) run cmd/gowebserver/gowebserver.go -http.port 8181

install: gowebserver
	@mkdir -p $(DESTDIR)$(bindir) $(DESTDIR)$(man1dir)
	@install ${BINARY_NAME} $(DESTDIR)$(bindir)
	@install -m 0644 ${MAN_PAGE_NAME} $(DESTDIR)$(man1dir)

deps:
	$(GO) mod tidy
	$(GO) mod download

BUILDX_BUILDER = buildx-builder
LINUX_PLATFORMS = amd64 arm64 ppc64le s390x arm/v5 arm/v6 arm/v7
space := $(null) #
comma := ,

ensure-builder:
	-$(DOCKER) buildx create --name $(BUILDX_BUILDER)

# https://github.com/docker-library/official-images#architectures-other-than-amd64
linux-images: ensure-builder $(foreach platform,$(LINUX_PLATFORMS),bin/image-artifacts/linux/$(platform)/gowebserver)
	$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform $(subst $(space),$(comma),$(strip $(foreach platform,$(LINUX_PLATFORMS),linux/$(platform)))) -f cmd/gowebserver/Dockerfile -t $(GOWEBSERVER_IMAGE):$(TAG) . $(DOCKER_PUSH)

WINDOWS_VERSIONS = 1809 2004 20H2
windows-images: bin/image-artifacts/windows/amd64/gowebserver.exe
	for winver in $(WINDOWS_VERSIONS) ; do \
		$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform windows/amd64 -f cmd/gowebserver/Dockerfile.windows --build-arg WINDOWS_VERSION=$$winver -t $(GOWEBSERVER_IMAGE):$(TAG)-windows_amd64-$$winver . $(DOCKER_PUSH) ; \
	done

ifeq ($(DOCKER_PUSH),--push)
	$(DOCKER) manifest create $(GOWEBSERVER_IMAGE):$(TAG) $(foreach winver,$(WINDOWS_VERSIONS),$(GOWEBSERVER_IMAGE):$(TAG)-windows_amd64-$(winver))
	$(DOCKER) manifest push $(GOWEBSERVER_IMAGE):$(TAG) $(GOWEBSERVER_IMAGE):$(TAG)
endif

push-image: DOCKER_PUSH = --push
push-image: linux-images windows-images
	

.PHONY : all assets dist lint clean check test test-10 coverage bench benchmark test-all package-legacy package install run deps
