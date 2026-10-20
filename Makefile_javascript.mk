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

include Makefile_core.mk

# JavaScript tests, benchmarks and the headless browsers they can run in.
#
# Node.js and npm are expected to be installed; everything else is fetched
# into build/toolchain, like the rest of the toolchain, and nothing is
# installed beside the sources (no package.json or node_modules in the
# project tree).
#
# Everything here is inert while JAVASCRIPT_DIR is empty, which is the
# template default: no JavaScript targets run, and `make test` stays exactly
# the Go and Terraform suites it is without this file. A project turns it on
# by setting, in its own Makefile (which sync-template.sh preserves):
#
#   JAVASCRIPT_DIR     the directory holding the JavaScript. The layout the
#                      targets expect is */test/*.test.js for tests that run
#                      on Node alone, */browser/*.browser.js for tests that
#                      drive a real browser, and bench.js, browserbench.js
#                      and browse.js at its top; all five names are
#                      variables below.
#   JAVASCRIPT_SERVER  a binary to build and start for the browser targets,
#                      relative to the repository root. Leave it empty when
#                      the pages come from somewhere else.
#
# Setting JAVASCRIPT_DIR also adds test-js to TEST_TARGETS, so it runs as
# part of `make test`.
#
# This file is template-managed: a template sync replaces it wholesale, so
# project settings belong in the Makefile, not here.

# What is run within JAVASCRIPT_DIR. Node globs the patterns it is given, so
# these are relative to JAVASCRIPT_DIR, not the repository root.
#
# JAVASCRIPT_DIR and JAVASCRIPT_SERVER are not assigned here at all, and
# everything else is assigned with ?=: a project sets these in its own
# Makefile, which includes this one, so a plain assignment here would
# overwrite whatever the project just set above it.
JAVASCRIPT_TEST ?= */test/*.test.js
JAVASCRIPT_BROWSER_TEST ?= */browser/*.browser.js
JAVASCRIPT_BENCH ?= bench.js
JAVASCRIPT_BROWSER_BENCH ?= browserbench.js
JAVASCRIPT_BROWSE ?= browse.js

# Arguments passed to the benchmark scripts; a project's own knobs go here.
JAVASCRIPT_BENCH_ARGS ?=
# Anything else the scripts read from the environment, as VAR=value pairs.
JAVASCRIPT_EXTRA_ENV ?=

NODE ?= node
NPM ?= npm

# https://github.com/microsoft/playwright/releases
PLAYWRIGHT_VERSION ?= 1.64.0

# The browsers to install and test in: chromium, firefox and/or webkit.
# Chromium and Firefox run on a plain Linux install. WebKit needs system
# libraries that only root can add ("playwright install-deps webkit"), so
# it's left out unless asked for: make test-browser HEADLESS_BROWSERS=webkit
HEADLESS_BROWSERS ?= chromium firefox
# The one a benchmark or a look at a page uses.
BROWSER ?= $(firstword $(HEADLESS_BROWSERS))

NODE_TOOLCHAIN_DIR = $(TOOLCHAIN_DIR)/node
BROWSERS_DIR = $(TOOLCHAIN_DIR)/browsers

PLAYWRIGHT = build/toolchain/node/node_modules/playwright/package.json
BROWSER_TOOLCHAIN = $(foreach browser,$(HEADLESS_BROWSERS),build/toolchain/browsers/$(browser).installed)
JAVASCRIPT_TOOLCHAIN = $(PLAYWRIGHT) $(BROWSER_TOOLCHAIN)

# Lets a script require('playwright') from the toolchain and find the
# browsers there, and tells it which server to start.
JAVASCRIPT_ENV = NODE_PATH="$(NODE_TOOLCHAIN_DIR)/node_modules" PLAYWRIGHT_BROWSERS_PATH="$(BROWSERS_DIR)" $(JAVASCRIPT_EXTRA_ENV) $(if $(JAVASCRIPT_SERVER),JAVASCRIPT_SERVER="$(REPOSITORY_ROOT)/$(JAVASCRIPT_SERVER)",)

ifeq ($(strip $(JAVASCRIPT_DIR)),)

# Not configured. The targets are still defined so that `make test-js` says
# what to set rather than "No rule to make target".
javascript-not-configured:
	@echo "No JavaScript configured: set JAVASCRIPT_DIR in the Makefile,"
	@echo "and JAVASCRIPT_SERVER for the browser targets, to enable them."

test-js benchmark-js test-browser benchmark-browser browse tools-js browsers javascript-server: javascript-not-configured

else

TEST_TARGETS += test-js

$(PLAYWRIGHT):
	mkdir -p "$(NODE_TOOLCHAIN_DIR)"
	$(NPM) install --prefix "$(NODE_TOOLCHAIN_DIR)" --no-save --no-audit --no-fund playwright@$(PLAYWRIGHT_VERSION)
	touch "$(REPOSITORY_ROOT)/$@"

# Headless Chromium is its own, smaller download.
build/toolchain/browsers/%.installed: $(PLAYWRIGHT)
	mkdir -p "$(dir $@)"
	PLAYWRIGHT_BROWSERS_PATH="$(BROWSERS_DIR)" "$(NODE_TOOLCHAIN_DIR)/node_modules/.bin/playwright$(EXE)" install $(patsubst chromium,chromium-headless-shell,$*)
	touch "$(REPOSITORY_ROOT)/$@"

tools-js: $(JAVASCRIPT_TOOLCHAIN)
browsers: $(BROWSER_TOOLCHAIN)

test-js:
	cd "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)" && $(NODE) --test "$(JAVASCRIPT_TEST)"

benchmark-js:
	cd "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)" && $(NODE) "$(JAVASCRIPT_BENCH)" $(JAVASCRIPT_BENCH_ARGS)

# The server the browser looks at, built again each time: its rule has no
# sources to tell it the embedded templates changed, and go build does
# nothing when they haven't.
ifneq ($(strip $(JAVASCRIPT_SERVER)),)
javascript-server:
	rm -f "$(REPOSITORY_ROOT)/$(JAVASCRIPT_SERVER)"
	$(MAKE) $(JAVASCRIPT_SERVER)
else
javascript-server:
endif

# The pages in each of HEADLESS_BROWSERS, served by JAVASCRIPT_SERVER. It
# isn't one of the TEST_TARGETS: it downloads browsers, which `make test`
# shouldn't need.
test-browser: $(JAVASCRIPT_TOOLCHAIN) javascript-server
	cd "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)" && for browser in $(HEADLESS_BROWSERS); do \
		echo "== $$browser"; \
		$(JAVASCRIPT_ENV) BROWSER=$$browser $(NODE) --test --test-concurrency=1 "$(JAVASCRIPT_BROWSER_TEST)" || exit 1; \
	done

# Page timings in BROWSER:
#   make benchmark-browser [BROWSER=firefox] [JAVASCRIPT_BENCH_ARGS=...]
benchmark-browser: $(JAVASCRIPT_TOOLCHAIN) javascript-server
	cd "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)" && $(JAVASCRIPT_ENV) BROWSER=$(BROWSER) $(NODE) "$(JAVASCRIPT_BROWSER_BENCH)" $(JAVASCRIPT_BENCH_ARGS)

# A look at a page in BROWSER: what it logged, what failed to load, how
# long it took and a screenshot. TARGET is a URL, or a directory to serve
# with JAVASCRIPT_SERVER and open.
#   make browse TARGET=~/Pictures [SCREENSHOT=build/page.png] [BROWSER=firefox]
browse: $(JAVASCRIPT_TOOLCHAIN) javascript-server
	$(JAVASCRIPT_ENV) BROWSER=$(BROWSER) $(NODE) "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)/$(JAVASCRIPT_BROWSE)" "$(TARGET)" "$(or $(SCREENSHOT),$(BUILD_DIR)/browse.png)"

endif

.PHONY: tools-js browsers javascript-server test-js benchmark-js test-browser benchmark-browser browse javascript-not-configured
