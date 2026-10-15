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

COMMA := ,
EXE =
FIND = find

REPOSITORY_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
BUILD_DIR = $(REPOSITORY_ROOT)/build
ARCHIVES_DIR = $(BUILD_DIR)/archives
TOOLCHAIN_DIR = $(BUILD_DIR)/toolchain
TOOLCHAIN_BIN = $(TOOLCHAIN_DIR)/bin
THIRDPARTY_DIR = $(REPOSITORY_ROOT)/third_party

TOOLCHAIN_GO = go
# Install to GOBIN (build/toolchain/bin, on PATH) instead of pinning GOPATH to
# the repo: Go stores the read-only module cache under $GOPATH/pkg/mod, so
# parking it in the repo means `actions/checkout`'s `git clean -ffdx` and `make
# clean` (rm -rf build/) have to delete thousands of read-only files — what
# makes the Windows self-hosted runner's checkout step take over a minute.
# Using the default GOPATH keeps the cache out of the repo (persistent across
# runs, never touched by clean); the proto plugins now match every other
# toolchain rule, which already uses GOBIN.
TOOLCHAIN_GO_INSTALL = GOBIN=$(TOOLCHAIN_BIN) $(GO_MIRROR_ENV)$(TOOLCHAIN_GO) install
CURL = curl --retry 5 --retry-connrefused

# DOWNLOAD_MIRROR, when set in the environment (e.g. http://localhost:3142),
# is a cache the build's downloads go through: internal/tools/downloadcache,
# run on each CI machine. Files, Go modules, npm packages, Playwright's
# browsers and ffmpeg are fetched through it, and straight from their
# servers when it fails. Unset, nothing changes.
DOWNLOAD_MIRROR ?=
# mirror_url is where the mirror serves URL $(1), or $(1) without one.
mirror_url = $(if $(DOWNLOAD_MIRROR),$(patsubst %/,%,$(DOWNLOAD_MIRROR))/$(subst ://,/,$(1)),$(1))
# download fetches URL $(2) to the file $(1), through the mirror if there
# is one, and straight from the URL if that fails. The mirror is asked only
# once: when it's down, retrying it only delays the direct download.
download = $(if $(DOWNLOAD_MIRROR),$(CURL) --retry 0 -fL -o "$(1)" "$(call mirror_url,$(2))" || )$(CURL) -fL -o "$(1)" "$(2)"
# MIRROR_UP is "yes" when the mirror answers, for clients that can't fall
# back by themselves (npm and Playwright). It's asked only where used.
MIRROR_UP = $(if $(DOWNLOAD_MIRROR),$(shell curl -sf --max-time 3 -o /dev/null "$(patsubst %/,%,$(DOWNLOAD_MIRROR))/healthz" && echo yes))
# GO_MIRROR_ENV goes before a go command that fetches modules: through the
# mirror, then the module proxy itself when the mirror fails.
GO_MIRROR_ENV = $(if $(DOWNLOAD_MIRROR),GOPROXY="$(call mirror_url,https://proxy.golang.org)|https://proxy.golang.org$(comma)direct" )
comma := ,

SHORT_SHA = $(shell git rev-parse --short=7 HEAD | tr -d [:punct:])
DIRTY_VERSION = v0.0.0-$(SHORT_SHA)
VERSION = $(shell git describe --tags || (echo $(DIRTY_VERSION) && exit 1))
BUILD_DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
TAG := $(VERSION)

export PATH := $(PWD)/build/toolchain/bin:$(PATH)
SOURCE_DIRS=$(shell go list ./... | grep -v '/vendor/')

ifeq ($(OS),Windows_NT)
	HOST_OS = windows
	HOST_PLATFORM = windows_amd64
	HOST_ARCH = amd64
	# Give priority to /usr/bin because it conflicts with C:\Windows\system32 within Msys32 environment.
	FIND = /usr/bin/find.exe
	EXE = .exe
	SED_REPLACE = sed -i
else
	UNAME_S := $(shell uname -s)
	UNAME_ARCH := $(shell uname -m)
	ifeq ($(UNAME_S),Linux)
		HOST_OS = linux
		SED_REPLACE = sed -i
		ifneq ($(filter arm aarch64,$(UNAME_ARCH)),)
			HOST_PLATFORM = linux_arm
			HOST_ARCH = arm
		else
			HOST_PLATFORM = linux_amd64
			HOST_ARCH = amd64
		endif
	endif
	ifeq ($(UNAME_S),Darwin)
		HOST_OS = darwin
		SED_REPLACE = sed -i ''
		ifeq ($(UNAME_ARCH),arm64)
			HOST_PLATFORM = darwin_arm64
			HOST_ARCH = arm64
		else
			HOST_PLATFORM = darwin_amd64
			HOST_ARCH = amd64
		endif
	endif
endif
