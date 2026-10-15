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
# Node.js and npm are expected to be installed; everything else is fetched
# into build/toolchain, like the rest of the toolchain, and nothing is
# installed beside the sources (no package.json or node_modules there).
#
# The Makefile that includes this sets:
#   JAVASCRIPT_DIR     where the tests are: */test/*.test.js run on Node
#                      alone, */browser/*.browser.js in a real browser
#                      against JAVASCRIPT_SERVER, and bench.js,
#                      browserbench.js and browse.js are at its top.
#   JAVASCRIPT_SERVER  the binary the browser tests start.
# and includes this before Makefile_build.mk, so that test-js is among the
# TEST_TARGETS when the test rule there is read.

NODE = node
NPM = npm

# https://github.com/microsoft/playwright/releases
PLAYWRIGHT_VERSION = 1.64.0

# The browsers to install and test in: chromium, firefox and/or webkit.
# Chromium and Firefox run on a plain Linux install. WebKit needs system
# libraries that only root can add ("playwright install-deps webkit"), so
# it's left out unless asked for: make test-browser HEADLESS_BROWSERS=webkit
HEADLESS_BROWSERS = chromium firefox
# The one a benchmark or a look at a page uses.
BROWSER = $(firstword $(HEADLESS_BROWSERS))

NODE_TOOLCHAIN_DIR = $(TOOLCHAIN_DIR)/node
BROWSERS_DIR = $(TOOLCHAIN_DIR)/browsers

PLAYWRIGHT = build/toolchain/node/node_modules/playwright/package.json
BROWSER_TOOLCHAIN = $(foreach browser,$(HEADLESS_BROWSERS),build/toolchain/browsers/$(browser).installed)
JAVASCRIPT_TOOLCHAIN = $(PLAYWRIGHT) $(BROWSER_TOOLCHAIN)

# Lets a script require('playwright') from the toolchain and find the
# browsers there, and tells it which server to start.
JAVASCRIPT_ENV = NODE_PATH="$(NODE_TOOLCHAIN_DIR)/node_modules" PLAYWRIGHT_BROWSERS_PATH="$(BROWSERS_DIR)" GOWEBSERVER="$(REPOSITORY_ROOT)/$(JAVASCRIPT_SERVER)"

TEST_TARGETS += test-js

# Through the download mirror (DOWNLOAD_MIRROR, see Makefile_core.mk) when
# it's up: neither falls back by itself. npm fetches packages from the
# registry it's given, and Playwright its browsers from the host it's
# given: Chromium from one server, everything else from a path on another.
NPM_MIRROR_FLAGS = $(if $(MIRROR_UP),--registry "$(call mirror_url,https://registry.npmjs.org)/")
PLAYWRIGHT_MIRROR_ENV = $(if $(MIRROR_UP),PLAYWRIGHT_CHROMIUM_DOWNLOAD_HOST="$(call mirror_url,https://cdn.playwright.dev)" PLAYWRIGHT_DOWNLOAD_HOST="$(call mirror_url,https://cdn.playwright.dev/dbazure/download/playwright)" )

$(PLAYWRIGHT):
	mkdir -p "$(NODE_TOOLCHAIN_DIR)"
	$(NPM) install --prefix "$(NODE_TOOLCHAIN_DIR)" --no-save --no-audit --no-fund $(NPM_MIRROR_FLAGS) playwright@$(PLAYWRIGHT_VERSION)
	touch "$(REPOSITORY_ROOT)/$@"

# Headless Chromium is its own, smaller download.
build/toolchain/browsers/%.installed: $(PLAYWRIGHT)
	mkdir -p "$(dir $@)"
	PLAYWRIGHT_BROWSERS_PATH="$(BROWSERS_DIR)" $(PLAYWRIGHT_MIRROR_ENV)"$(NODE_TOOLCHAIN_DIR)/node_modules/.bin/playwright" install $(patsubst chromium,chromium-headless-shell,$*)
	touch "$(REPOSITORY_ROOT)/$@"

tools-js: $(JAVASCRIPT_TOOLCHAIN)
browsers: $(BROWSER_TOOLCHAIN)

test-js:
	cd "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)" && $(NODE) --test

benchmark-js:
	cd "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)" && $(NODE) bench.js $(BENCH)

# The server the browser looks at, built again each time: its rule has no
# sources to tell it the embedded templates changed, and go build does
# nothing when they haven't.
javascript-server:
	rm -f "$(REPOSITORY_ROOT)/$(JAVASCRIPT_SERVER)"
	$(MAKE) $(JAVASCRIPT_SERVER)

# The pages in each of HEADLESS_BROWSERS, served by JAVASCRIPT_SERVER. It
# isn't one of the TEST_TARGETS: it downloads browsers, which `make test`
# shouldn't need.
test-browser: $(JAVASCRIPT_TOOLCHAIN) javascript-server
	cd "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)" && for browser in $(HEADLESS_BROWSERS); do \
		echo "== $$browser"; \
		$(JAVASCRIPT_ENV) BROWSER=$$browser $(NODE) --test --test-concurrency=1 "*/browser/*.browser.js" || exit 1; \
	done

# Page timings in BROWSER:
#   make benchmark-browser [BROWSER=firefox] [PHOTOS=5000] [PHOTO_SIZE=3000x2000] [THUMBNAILS=off]
benchmark-browser: $(JAVASCRIPT_TOOLCHAIN) javascript-server
	cd "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)" && $(JAVASCRIPT_ENV) BROWSER=$(BROWSER) PHOTOS=$(PHOTOS) PHOTO_SIZE=$(PHOTO_SIZE) THUMBNAILS=$(THUMBNAILS) $(NODE) browserbench.js

# A look at a page in BROWSER: what it logged, what failed to load, how
# long it took and a screenshot. TARGET is a URL, or a directory to serve
# with JAVASCRIPT_SERVER and open.
#   make browse TARGET=~/Pictures [SCREENSHOT=build/page.png] [BROWSER=firefox]
browse: $(JAVASCRIPT_TOOLCHAIN) javascript-server
	$(JAVASCRIPT_ENV) BROWSER=$(BROWSER) $(NODE) "$(REPOSITORY_ROOT)/$(JAVASCRIPT_DIR)/browse.js" "$(TARGET)" "$(or $(SCREENSHOT),$(BUILD_DIR)/browse.png)"

# This is included first, so its first rule would become what a bare
# `make` builds. Leave that to the makefile included next, as before.
.DEFAULT_GOAL :=

.PHONY: tools-js browsers javascript-server test-js benchmark-js test-browser benchmark-browser browse
