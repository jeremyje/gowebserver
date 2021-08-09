# Copyright 2021 Jeremy Edwards
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

EXE_EXTENSION =
GO = GO111MODULE=on go
_INTERNAL_GO = $(GO)
GO_TOOLCHAIN_DIR = $(dir $(abspath golang.mk))bin/toolchain

# Constant modtime value so that the created files are consistent.
BINDATA_MODTIME := 1557978307

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

bin/toolchain/go-bindata$(EXE_EXTENSION):
	mkdir -p $(dir $(abspath $@))
	cd $(dir $(abspath $@)) && $(_INTERNAL_GO) build -pkgdir . github.com/go-bindata/go-bindata/go-bindata
	touch $@

bin/toolchain/go-bindata-assetfs$(EXE_EXTENSION):
	mkdir -p $(dir $(abspath $@))
	cd $(dir $(abspath $@)) && $(_INTERNAL_GO) build -pkgdir . github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs
	touch $@

bin/go/%:
	GOOS=$(firstword $(subst _, ,$(notdir $(abspath $(dir $@))))) GOARCH=$(word 2, $(subst _, ,$(notdir $(abspath $(dir $@))))) GOARM=$(subst v,,$(word 3, $(subst _, ,$(notdir $(abspath $(dir $@)))))) CGO_ENABLED=0 $(_INTERNAL_GO) build -o $@ cmd/$(basename $(notdir $@))/$(basename $(notdir $@)).go
	touch $@

%/bindata.go: bin/toolchain/go-bindata$(EXE_EXTENSION)
	cd $(dir $@); $(GO_TOOLCHAIN_DIR)/go-bindata$(EXE_EXTENSION) -modtime $(BINDATA_MODTIME) -pkg $(notdir $(abspath $(dir $@))) -o bindata.go data/...
	$(_INTERNAL_GO) fmt $@
	touch $@

%/bindata_assetfs.go: %/bindata.go bin/toolchain/go-bindata-assetfs$(EXE_EXTENSION)
	cd $(dir $@); $(GO_TOOLCHAIN_DIR)/go-bindata-assetfs$(EXE_EXTENSION) -modtime $(BINDATA_MODTIME) -pkg $(notdir $(abspath $(dir $@))) -o bindata_assetfs.go data/...
	rm -f $*/bindata.go
	$(_INTERNAL_GO) fmt $@
	touch $@
