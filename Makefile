# Copyright 2026 Cloudfra
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

REGISTRY = docker.io/cloudfra
PROTOS =
TEST_ASSETS = $(TEST_ARCHIVES)
ASSETS = $(PROTOS)
GO_PACKAGE = github.com/cloudfra/gowebserver
ALL_APPS = gowebserver
PRODUCTION=1
RUMDL_IGNORE = internal/gowebserver/testing/testassets/**
PLAN9_PLATFORMS = 
IOS_PLATFORMS = 
NETBSD_PLATFORMS = netbsd/amd64 netbsd/arm64 netbsd/arm/v5 netbsd/arm/v6 netbsd/arm/v7 # netbsd/386 

JAVASCRIPT_DIR = pkg/gowebserver/template
JAVASCRIPT_SERVER = build/bin/$(HOST_OS)/$(HOST_ARCH)/gowebserver$(EXE)
# What the scripts read beyond what the template passes (NODE_PATH,
# PLAYWRIGHT_BROWSERS_PATH and JAVASCRIPT_SERVER itself): browser.js names
# the server binary GOWEBSERVER, and browserbench.js takes its workload from
# the environment.
JAVASCRIPT_EXTRA_ENV = GOWEBSERVER="$(REPOSITORY_ROOT)/$(JAVASCRIPT_SERVER)" PHOTOS=$(PHOTOS) PHOTO_SIZE=$(PHOTO_SIZE) THUMBNAILS=$(THUMBNAILS)

# The one tool this project needs that the template's toolchain doesn't
# ship. The tests that make real thumbnails find ffmpeg on PATH, which
# already includes build/toolchain/bin, and skip without it.
PROJECT_TOOLS = build/toolchain/bin/ffmpeg$(EXE)

include Makefile_build.mk

# pkg/ffmpeg does the downloading, so the license is accepted here and the
# archive is kept in the user cache directory: `make clean` removes the
# installed binary but not the download, which is fetched again only for a
# new ffmpeg version. This rule sits below the include so it doesn't become
# what a bare `make` builds.
build/toolchain/bin/ffmpeg$(EXE):
	$(GO) run ./internal/tools/installffmpeg -o "$@"

run: clean assets lint
	$(GO) run cmd/gowebserver/gowebserver.go -http.port 8181 -path=. -verbose -debug -enhancedindex=true

multirun: clean assets lint
	$(GO) run cmd/gowebserver/gowebserver.go -path=./cmd/,./pkg/,. -verbose=true -servepath=mains,code,root -http.port 8181 -enhancedindex=true -debug

run-wasm: clean assets lint build/bin/js/wasm/gowebserver build/bin/js/wasm/gowebserver.html
	$(GO) run cmd/gowebserver/gowebserver.go -http.port 8181 -path=$(REPOSITORY_ROOT)/build/bin/js/wasm/ -verbose

kind-create:
	$(KIND) create cluster --config=$(REPOSITORY_ROOT)/install/kind/kind-cluster.yaml
# kubectl config set clusters.kind-kind.server https://192.168.86.36:6443

kind-delete:
	$(KIND) delete cluster

install/kubernetes.yaml:
	$(HELM) template gowebserver install/helm > install/kubernetes.yaml

template: install/kubernetes.yaml

test-codecov:
	curl -X POST --data-binary @codecov.yml https://codecov.io/validate

.PHONY : run multirun run-wasm kind-create kind-delete template test-codecov
