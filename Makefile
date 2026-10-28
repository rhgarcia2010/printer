# Book Printer build tasks. Run `make help` for the list.
#
# Works on Linux and macOS. On Windows, run the wails commands from the README.

UNAME := $(shell uname -s)
GOBIN := $(shell go env GOBIN)
GOPATH := $(shell go env GOPATH)
# Use wails from PATH, or where `go install` puts it.
WAILS := $(or $(shell command -v wails 2>/dev/null),$(if $(GOBIN),$(GOBIN)/wails,$(GOPATH)/bin/wails))

# Linux ships WebKitGTK 4.1; Wails needs this tag to use it.
ifeq ($(UNAME),Linux)
TAGS := -tags webkit2_41
endif

.DEFAULT_GOAL := build
.PHONY: help setup doctor build build-universal dev test check run clean

help: ## Show available targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-16s %s\n", $$1, $$2}'

setup: ## Install the Wails CLI and frontend dependencies
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	cd frontend && npm install

doctor: ## Check that the build prerequisites are installed
	@ok=1; \
	check() { if command -v $$1 >/dev/null 2>&1; then printf '  ok       %s\n' "$$2"; else printf '  MISSING  %s\n' "$$2  ($$3)"; ok=0; fi; }; \
	echo "Checking build prerequisites on $(UNAME):"; \
	check go "Go $$(go env GOVERSION 2>/dev/null) (need 1.26+)" "https://go.dev/dl"; \
	check node "Node.js $$(node --version 2>/dev/null) (need 20.19+ or 22.12+)" "https://nodejs.org"; \
	check npm "npm" "comes with Node.js"; \
	if [ -x "$(WAILS)" ]; then echo "  ok       Wails CLI"; else echo "  MISSING  Wails CLI  (run: make setup)"; ok=0; fi; \
	if [ "$(UNAME)" = Darwin ]; then \
		if xcode-select -p >/dev/null 2>&1; then echo "  ok       Xcode command line tools"; \
		else echo "  MISSING  Xcode command line tools  (run: xcode-select --install)"; ok=0; fi; \
	fi; \
	if [ "$(UNAME)" = Linux ]; then \
		check gcc "gcc" "install your distro's build tools"; \
		if pkg-config --exists gtk+-3.0 webkit2gtk-4.1 2>/dev/null; then echo "  ok       GTK 3 + WebKitGTK 4.1"; \
		else echo "  MISSING  GTK 3 / WebKitGTK 4.1  (Arch: pacman -S gtk3 webkit2gtk-4.1; Debian/Ubuntu: apt install libgtk-3-dev libwebkit2gtk-4.1-dev)"; ok=0; fi; \
	fi; \
	[ $$ok = 1 ] && echo "All set." || { echo "Install the missing pieces above, then run make build."; exit 1; }

# `wails build -clean` empties frontend/dist, placeholder included; it is put
# back so plain `go build` / `go test` keep compiling (main.go embeds dist).
build: ## Build the app for this machine into build/bin
	$(WAILS) build $(TAGS) -clean
	@touch frontend/dist/gitkeep
	@echo; echo "Built: $$(ls -d build/bin/* | head -1)"

build-universal: ## macOS only: one app that runs on Intel and Apple Silicon
	@[ "$(UNAME)" = Darwin ] || { echo "build-universal only works on macOS"; exit 1; }
	$(WAILS) build -platform darwin/universal -clean
	@touch frontend/dist/gitkeep

dev: ## Run with live reload while editing
	$(WAILS) dev $(TAGS)

test: ## Run the Go tests
	go test ./...

check: ## Vet Go code and type-check the frontend
	go vet $(TAGS) ./...
	cd frontend && npx svelte-check --tsconfig ./tsconfig.json

run: build ## Build, then start the app
ifeq ($(UNAME),Darwin)
	open build/bin/*.app
else
	./build/bin/book-printer
endif

clean: ## Remove build output
	rm -rf build/bin
	find frontend/dist -mindepth 1 ! -name gitkeep -exec rm -rf {} +
	touch frontend/dist/gitkeep
