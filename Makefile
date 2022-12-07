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

RM = rm
ZIP = zip
TAR = tar
SEVENZIP = 7z
ECHO = @echo
GO = GO111MODULE=on go
DOCKER = DOCKER_CLI_EXPERIMENTAL=enabled docker
KIND = kind
HELM = helm

EXE_EXTENSION =
SHORT_SHA = $(shell git rev-parse --short=7 HEAD | tr -d [:punct:])
DIRTY_VERSION = v0.0.0-$(SHORT_SHA)
VERSION = $(shell git describe --tags || (echo $(DIRTY_VERSION) && exit 1))
BUILD_DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
TAG := $(VERSION)
PKG := github.com/jeremyje/gowebserver/v2

SOURCE_DIRS=$(shell go list ./... | grep -v '/vendor/')
export PATH := $(PWD)/bin/toolchain:$(PATH):/root/go/bin:/usr/lib/go-1.9/bin:/usr/local/go/bin:/usr/go/bin
BINARY_NAME=gowebserver
MAN_PAGE_NAME=${BINARY_NAME}.1
REPOSITORY_ROOT := $(patsubst %/,%,$(dir $(abspath Makefile)))

REGISTRY = docker.io/jeremyje
CERTTOOL_IMAGE = $(REGISTRY)/certtool
GOWEBSERVER_IMAGE = $(REGISTRY)/gowebserver
HTTPPROBE_IMAGE = $(REGISTRY)/httpprobe

GO_TOOLCHAIN_DIR = $(dir $(abspath golang.mk))bin/toolchain

LINUX_PLATFORMS = linux_386 linux_amd64 linux_arm_v5 linux_arm_v6 linux_arm_v7 linux_arm64 linux_s390x linux_ppc64 linux_ppc64le linux_riscv64 linux_mips64le linux_mips linux_mipsle linux_mips64
ANDROID_PLATFORMS = android_arm64 # android_386 android_amd64 android_arm android_arm_v5 android_arm_v6 android_arm_v7
WINDOWS_PLATFORMS = windows_386 windows_amd64 windows_arm64 windows_arm_v5 windows_arm_v6 windows_arm_v7
MAIN_PLATFORMS = windows_amd64 linux_amd64 linux_arm64
IOS_PLATFORMS = #ios_amd64 ios_arm64
DARWIN_PLATFORMS = darwin_amd64 darwin_arm64
DRAGONFLY_PLATFORMS = dragonfly_amd64
FREEBSD_PLATFORMS = freebsd_386 freebsd_amd64 freebsd_arm_v5 freebsd_arm_v6 freebsd_arm_v7 freebsd_arm64
NETBSD_PLATFORMS = netbsd_386 netbsd_amd64 netbsd_arm_v5 netbsd_arm_v6 netbsd_arm_v7 netbsd_arm64
OPENBSD_PLATFORMS = openbsd_386 openbsd_amd64 openbsd_arm_v5 openbsd_arm_v6 openbsd_arm_v7 openbsd_arm64 openbsd_mips64
PLAN9_PLATFORMS = plan9_386 plan9_amd64 plan9_arm_v5 plan9_arm_v6 plan9_arm_v7
NICHE_PLATFORMS = js_wasm solaris_amd64 illumos_amd64 aix_ppc64 $(ANDROID_PLATFORMS) $(DARWIN_PLATFORMS) $(IOS_PLATFORMS) $(DRAGONFLY_PLATFORMS) $(FREEBSD_PLATFORMS) $(NETBSD_PLATFORMS) $(OPENBSD_PLATFORMS) $(PLAN9_PLATFORMS)
ALL_PLATFORMS = $(LINUX_PLATFORMS) $(WINDOWS_PLATFORMS) $(NICHE_PLATFORMS)
TEST_ARCHIVES = internal/gowebserver/testing/testassets.zip
TEST_ARCHIVES += internal/gowebserver/testing/testassets.tar.gz
TEST_ARCHIVES += internal/gowebserver/testing/testassets.tar.bz2
TEST_ARCHIVES += internal/gowebserver/testing/testassets.tar
TEST_ARCHIVES += internal/gowebserver/testing/testassets.7z
TEST_ARCHIVES += internal/gowebserver/testing/testassets.tar.xz
TEST_ARCHIVES += internal/gowebserver/testing/testassets.tar.lz4
WASM_ASSETS = install/wasm/wasm_exec.js install/wasm/wasm_exec.html install/wasm/gowebserver.wasm
ASSETS = $(TEST_ARCHIVES) internal/gowebserver/testing/nested-testassets.zip internal/gowebserver/testing/single-testassets.zip internal/gowebserver/testing/nodir-testassets.zip
ALL_APPS = gowebserver certtool httpprobe

ALL_BINARIES = $(foreach app,$(ALL_APPS),$(foreach platform,$(ALL_PLATFORMS),bin/go/$(platform)/$(app)$(if $(findstring windows_,$(platform)),.exe,)))
WINDOWS_VERSIONS = 1709 1803 1809 1903 1909 2004 20H2 ltsc2022
BUILDX_BUILDER = buildx-builder
space := $(null) #
comma := ,

ifeq ($(OS),Windows_NT)
	HOST_OS = windows
	HOST_PLATFORM = windows_amd64
	EXE_EXTENSION = .exe
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		HOST_OS = linux
		ifeq ($(UNAME_ARCH),arm)
			HOST_PLATFORM = linux_arm
		else
			HOST_PLATFORM = linux_amd64
		endif
	endif
	ifeq ($(UNAME_S),Darwin)
		HOST_OS = darwin
		HOST_PLATFORM = darwin_amd64
	endif
endif

all: $(ALL_BINARIES) assets
assets: $(ASSETS) $(WASM_ASSETS)

bin/go/%: $(ASSETS)
	GOOS=$(firstword $(subst _, ,$(notdir $(abspath $(dir $@))))) GOARCH=$(word 2, $(subst _, ,$(notdir $(abspath $(dir $@))))) GOARM=$(subst v,,$(word 3, $(subst _, ,$(notdir $(abspath $(dir $@)))))) CGO_ENABLED=0 \
		$(GO) build -o $@ \
		-ldflags '-X $(PKG)/pkg/gowebserver.version=$(VERSION)' \
		cmd/$(basename $(notdir $@))/$(basename $(notdir $@)).go
	touch $@

SHORT_APP_NAMES = server httpprobe certtool
RELEASE_BINARY_SUFFIXES = amd64 arm arm64 386 arm amd64-darwin arm64-darwin amd64.exe 386.exe
RELEASE_BINARIES = $(foreach appname,$(SHORT_APP_NAMES),$(foreach relbin,$(RELEASE_BINARY_SUFFIXES),bin/release/$(appname)-$(relbin)))

release-binaries: $(RELEASE_BINARIES)

bin/release/certtool-amd64: bin/go/linux_amd64/certtool
	mkdir -p bin/release/ && cp $< $@

bin/release/certtool-arm: bin/go/linux_arm_v7/certtool
	mkdir -p bin/release/ && cp $< $@

bin/release/certtool-arm64: bin/go/linux_arm64/certtool
	mkdir -p bin/release/ && cp $< $@

bin/release/certtool-386: bin/go/linux_386/certtool
	mkdir -p bin/release/ && cp $< $@

bin/release/certtool-amd64-darwin: bin/go/darwin_amd64/certtool
	mkdir -p bin/release/ && cp $< $@

bin/release/certtool-arm64-darwin: bin/go/darwin_arm64/certtool
	mkdir -p bin/release/ && cp $< $@

bin/release/certtool-amd64.exe: bin/go/windows_amd64/certtool.exe
	mkdir -p bin/release/ && cp $< $@

bin/release/certtool-386.exe: bin/go/windows_386/certtool.exe
	mkdir -p bin/release/ && cp $< $@

bin/release/certtool-arm64.exe: bin/go/windows_arm64/certtool.exe
	mkdir -p bin/release/ && cp $< $@

bin/release/httpprobe-amd64: bin/go/linux_amd64/httpprobe
	mkdir -p bin/release/ && cp $< $@

bin/release/httpprobe-arm: bin/go/linux_arm_v7/httpprobe
	mkdir -p bin/release/ && cp $< $@

bin/release/httpprobe-arm64: bin/go/linux_arm64/httpprobe
	mkdir -p bin/release/ && cp $< $@

bin/release/httpprobe-386: bin/go/linux_386/httpprobe
	mkdir -p bin/release/ && cp $< $@

bin/release/httpprobe-amd64-darwin: bin/go/darwin_amd64/httpprobe
	mkdir -p bin/release/ && cp $< $@

bin/release/httpprobe-arm64-darwin: bin/go/darwin_arm64/httpprobe
	mkdir -p bin/release/ && cp $< $@

bin/release/httpprobe-amd64.exe: bin/go/windows_amd64/httpprobe.exe
	mkdir -p bin/release/ && cp $< $@

bin/release/httpprobe-386.exe: bin/go/windows_386/httpprobe.exe
	mkdir -p bin/release/ && cp $< $@

bin/release/httpprobe-arm64.exe: bin/go/windows_arm64/httpprobe.exe
	mkdir -p bin/release/ && cp $< $@

bin/release/server-amd64: bin/go/linux_amd64/gowebserver
	mkdir -p bin/release/ && cp $< $@

bin/release/server-arm: bin/go/linux_arm_v7/gowebserver
	mkdir -p bin/release/ && cp $< $@

bin/release/server-arm64: bin/go/linux_arm64/gowebserver
	mkdir -p bin/release/ && cp $< $@

bin/release/server-386: bin/go/linux_386/gowebserver
	mkdir -p bin/release/ && cp $< $@

bin/release/server-amd64-darwin: bin/go/darwin_amd64/gowebserver
	mkdir -p bin/release/ && cp $< $@

bin/release/server-arm64-darwin: bin/go/darwin_arm64/gowebserver
	mkdir -p bin/release/ && cp $< $@

bin/release/server-amd64.exe: bin/go/windows_amd64/gowebserver.exe
	mkdir -p bin/release/ && cp $< $@

bin/release/server-386.exe: bin/go/windows_386/gowebserver.exe
	mkdir -p bin/release/ && cp $< $@

bin/release/server-arm64.exe: bin/go/windows_arm64/gowebserver.exe
	mkdir -p bin/release/ && cp $< $@

dist: bin/release.tar.gz

bin/release.tar.gz: $(ALL_BINARIES)
	mkdir -p bin/
	cd bin/go/; $(TAR) -I 'gzip -9' -cf ../release.tar.gz *

lint: $(ASSETS)
	$(GO) fmt ${SOURCE_DIRS}
	$(GO) vet ${SOURCE_DIRS}

clean:
	$(RM) -f ${BINARY_NAME} ${BINARY_NAME}-* cert.pem rsa.pem release.tar.gz $(ASSETS) *.tar.bz2 *.snap
	$(RM) -rf parts/ prime/ snap/.snapcraft/ stage/ *.snap
	$(RM) -rf upload/
	$(RM) -rf toolchain/
	$(RM) -rf bin/
	$(RM) -f install/wasm/wasm_exec.js
	$(RM) -f install/wasm/wasm_exec.html

check: test

install/wasm/gowebserver.wasm: bin/go/js_wasm/gowebserver
	cp -f $< $@

install/wasm/wasm_exec.js:
	cp -f $(shell go env GOROOT)/misc/wasm/wasm_exec.js $@

install/wasm/wasm_exec.html:
	cp -f $(shell go env GOROOT)/misc/wasm/wasm_exec.html $@

internal/gowebserver/testing/nodir-testassets.zip: $(TEST_ARCHIVES) internal/gowebserver/testing/single-testassets.zip internal/gowebserver/testing/nested-testassets.zip
	cd internal/gowebserver/testing/testassets; $(ZIP) -qr9 ../../nodir-testassets.zip index.html assets/1.txt assets/2.txt site.js "weird #1.txt" weird#.txt weird$$.txt assets/more/3.txt assets/four/4.txt assets/fivesix/5.txt assets/fivesix/6.txt
	mv internal/gowebserver/nodir-testassets.zip internal/gowebserver/testing/nodir-testassets.zip

internal/gowebserver/testing/single-testassets.zip: $(TEST_ARCHIVES)
	cd internal/gowebserver/testing/; $(ZIP) -qr9 ../single-testassets.zip testassets/
	mv internal/gowebserver/single-testassets.zip internal/gowebserver/testing/single-testassets.zip

internal/gowebserver/testing/nested-testassets.zip: $(TEST_ARCHIVES) internal/gowebserver/testing/single-testassets.zip
	cd internal/gowebserver/testing/; $(ZIP) -qr9 ../nested-testassets.zip *
	mv internal/gowebserver/nested-testassets.zip internal/gowebserver/testing/nested-testassets.zip

internal/gowebserver/testing/testassets.zip:
	cd internal/gowebserver/testing/testassets/; $(ZIP) -qr9 ../testassets.zip *

internal/gowebserver/testing/testassets.tar.gz:
	cd internal/gowebserver/testing/testassets/; $(TAR) -I 'gzip -9' -cf ../testassets.tar.gz *

internal/gowebserver/testing/testassets.tar.bz2:
	cd internal/gowebserver/testing/testassets/; BZIP=-9 $(TAR) cjf ../testassets.tar.bz2 *

internal/gowebserver/testing/testassets.tar.xz:
	cd internal/gowebserver/testing/testassets/; $(TAR) cJf ../testassets.tar.xz *

internal/gowebserver/testing/testassets.tar.lz4:
	cd internal/gowebserver/testing/testassets/; $(TAR) cf ../testassets.tar.lz4 -I 'lz4' *

internal/gowebserver/testing/testassets.tar:
	cd internal/gowebserver/testing/testassets/; $(TAR) cf ../testassets.tar *

internal/gowebserver/testing/testassets.7z:
	cd internal/gowebserver/testing/testassets/; $(SEVENZIP) a ../testassets.7z *

test: $(ASSETS)
	$(GO) test -race ${SOURCE_DIRS}

test-10: $(ASSETS)
	$(GO) test -race ${SOURCE_DIRS} -count 10

coverage: $(ASSETS)
	$(GO) test -cover ${SOURCE_DIRS}

coverage.txt: $(ASSETS)
	for sfile in ${SOURCE_DIRS} ; do \
		go test -race "$$sfile" -coverprofile=package.coverage -covermode=atomic; \
		if [ -f package.coverage ]; then \
			cat package.coverage >> coverage.txt; \
			$(RM) package.coverage; \
		fi; \
	done

bench: benchmark
benchmark: $(ASSETS)
	$(GO) test -benchmem -bench=. ${SOURCE_DIRS}

test-all: test test-10 benchmark coverage

run: clean $(ASSETS) lint
	$(GO) run cmd/gowebserver/gowebserver.go -http.port 8181 -path=git@github.com:jeremyje/gowebserver.git -verbose

multirun: clean $(ASSETS) lint
	$(GO) run cmd/gowebserver/gowebserver.go -path=./cmd/,./pkg/,. -servepath=mains,code,root -http.port 8181 -enhancedindex=true

install: gowebserver
	mkdir -p $(DESTDIR)$(bindir) $(DESTDIR)$(man1dir)
	install ${BINARY_NAME} $(DESTDIR)$(bindir)
	install -m 0644 ${MAN_PAGE_NAME} $(DESTDIR)$(man1dir)

deps:
	$(GO) mod tidy
	$(GO) mod download

ensure-builder:
	-$(DOCKER) buildx create --name $(BUILDX_BUILDER)

ALL_IMAGES = $(GOWEBSERVER_IMAGE) $(CERTTOOL_IMAGE) $(HTTPPROBE_IMAGE)
# https://github.com/docker-library/official-images#architectures-other-than-amd64
images: DOCKER_PUSH = --push
images: linux-images windows-images
	-$(DOCKER) manifest rm $(GOWEBSERVER_IMAGE):$(TAG)
	-$(DOCKER) manifest rm $(CERTTOOL_IMAGE):$(TAG)
	-$(DOCKER) manifest rm $(HTTPPROBE_IMAGE):$(TAG)

	for image in $(ALL_IMAGES) ; do \
		$(DOCKER) manifest create $$image:$(TAG) $(foreach winver,$(WINDOWS_VERSIONS),$${image}:$(TAG)-windows_amd64-$(winver)) $(foreach platform,$(LINUX_PLATFORMS),$${image}:$(TAG)-$(platform)) ; \
		for winver in $(WINDOWS_VERSIONS) ; do \
			windows_version=`$(DOCKER) manifest inspect mcr.microsoft.com/windows/nanoserver:$${winver} | jq -r '.manifests[0].platform["os.version"]'`; \
			$(DOCKER) manifest annotate --os-version $${windows_version} $${image}:$(TAG) $${image}:$(TAG)-windows_amd64-$${winver} ; \
		done ; \
		$(DOCKER) manifest push $$image:$(TAG) ; \
	done

ALL_LINUX_IMAGES = $(foreach app,$(ALL_APPS),$(foreach platform,$(LINUX_PLATFORMS),linux-image-$(app)-$(platform)))
linux-images: $(ALL_LINUX_IMAGES)

linux-image-certtool-%: bin/go/%/certtool ensure-builder
	$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform $(subst _,/,$*) --build-arg BINARY_PATH=$< -f cmd/certtool/Dockerfile -t $(CERTTOOL_IMAGE):$(TAG)-$* . $(DOCKER_PUSH)

linux-image-gowebserver-%: bin/go/%/gowebserver ensure-builder
	$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform $(subst _,/,$*) --build-arg BINARY_PATH=$< -f cmd/gowebserver/Dockerfile -t $(GOWEBSERVER_IMAGE):$(TAG)-$* . $(DOCKER_PUSH)

linux-image-httpprobe-%: bin/go/%/httpprobe ensure-builder
	$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform $(subst _,/,$*) --build-arg BINARY_PATH=$< -f cmd/httpprobe/Dockerfile -t $(HTTPPROBE_IMAGE):$(TAG)-$* . $(DOCKER_PUSH)

ALL_WINDOWS_IMAGES = $(foreach app,$(ALL_APPS),$(foreach winver,$(WINDOWS_VERSIONS),windows-image-$(app)-$(winver)))
windows-images: $(ALL_WINDOWS_IMAGES)

windows-image-certtool-%: bin/go/windows_amd64/certtool.exe ensure-builder
	$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform windows/amd64 -f cmd/certtool/Dockerfile.windows --build-arg WINDOWS_VERSION=$* -t $(CERTTOOL_IMAGE):$(TAG)-windows_amd64-$* . $(DOCKER_PUSH)

windows-image-gowebserver-%: bin/go/windows_amd64/gowebserver.exe ensure-builder
	$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform windows/amd64 -f cmd/gowebserver/Dockerfile.windows --build-arg WINDOWS_VERSION=$* -t $(GOWEBSERVER_IMAGE):$(TAG)-windows_amd64-$* . $(DOCKER_PUSH)

windows-image-httpprobe-%: bin/go/windows_amd64/httpprobe.exe ensure-builder
	$(DOCKER) buildx build --builder $(BUILDX_BUILDER) --platform windows/amd64 -f cmd/httpprobe/Dockerfile.windows --build-arg WINDOWS_VERSION=$* -t $(HTTPPROBE_IMAGE):$(TAG)-windows_amd64-$* . $(DOCKER_PUSH)

presubmit: clean check coverage all release-binaries images

kind-create:
	$(KIND) create cluster --config=$(REPOSITORY_ROOT)/install/kind/kind-cluster.yaml
# kubectl config set clusters.kind-kind.server https://192.168.86.36:6443

kind-delete:
	$(KIND) delete cluster

install/kubernetes.yaml:
	$(HELM) template gowebserver install/helm > install/kubernetes.yaml

template: install/kubernetes.yaml

bs: pkg/gowebserver/static/bootstrap/
pkg/gowebserver/static/bootstrap/:
	mkdir -p pkg/gowebserver/static/
	cd pkg/gowebserver/static/ \
		&& curl -o bootstrap.zip -L https://github.com/twbs/bootstrap/releases/download/v5.1.3/bootstrap-5.1.3-dist.zip \
		&& unzip bootstrap.zip \
		&& mv bootstrap-5.1.3-dist/ bootstrap/ \
		&& rm -f bootstrap.zip

test-codecov:
	curl -X POST --data-binary @codecov.yml https://codecov.io/validate

run-wasm: clean assets lint
	$(GO) run cmd/gowebserver/gowebserver.go -http.port 8181 -path=install/wasm/ -verbose

.PHONY : all assets dist lint clean check test test-10 coverage bench benchmark test-all install run deps presubmit
