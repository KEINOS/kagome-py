# Detect platform once
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Platform-specific library extension
ifeq ($(GOOS),darwin)
    LIB_EXT := dylib
else ifeq ($(GOOS),windows)
    LIB_EXT := dll
else ifeq ($(GOOS),linux)
    LIB_EXT := so
else
    $(error Unsupported GOOS: $(GOOS))
endif

.PHONY: clean build-archive build-shared stage-lib build test-go test-python test wheel sdist
.PHONY: docker-pull docker-clean docker-build docker-test

# ---------------------------------------------------------------------------
# Clean
# ---------------------------------------------------------------------------

clean:
	@rm -rf ./_src_c/build/libkagome*
	@rm -f ./src/libkagome/lib/libkagome.*
	@rm -rf ./build ./dist ./*.egg-info ./src/*.egg-info

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------

build-archive:
	@mkdir -p ./_src_c/build
	@rm -rf ./_src_c/build/lib*
	@cd ./_src_c/go && \
	go clean -testcache && go test -race -v ./... && \
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=c-archive -o ../build/libkagome.a ./main.go

build-shared: build-archive
	@cd ./_src_c && \
	cc -fPIC \
		-fvisibility=hidden \
		-shared \
		./c/kagome_wrapper.c ./build/libkagome.a \
		-o ./build/libkagome.$(LIB_EXT)

stage-lib:
	@mkdir -p ./src/libkagome/lib
	@cp ./_src_c/build/libkagome.$(LIB_EXT) ./src/libkagome/lib/libkagome.$(LIB_EXT)

build: build-shared stage-lib

# ---------------------------------------------------------------------------
# Test
# ---------------------------------------------------------------------------

test-go:
	@if ! command -v go >/dev/null 2>&1; then \
		echo "[SKIP] Go is not installed."; \
		exit 0; \
	fi; \
	echo "**NOTE**: On macOS, 'ld: warning: ...' can be ignored if tests pass."; \
	cd ./_src_c/go && \
	go clean -testcache && go test -race ./...

test-python:
	@if ! command -v python3 >/dev/null 2>&1; then \
		echo "[SKIP] Python3 is not installed."; \
		exit 0; \
	fi; \
	if [ -f ./src/libkagome/lib/libkagome.so ] || [ -f ./src/libkagome/lib/libkagome.dll ] || [ -f ./src/libkagome/lib/libkagome.dylib ]; then \
		PYTHONPATH=./src python3 ./tests/libkagome_test.py ; \
	else \
		echo "No libkagome.(so|dll|dylib) found in src/libkagome/lib/. Please run 'make build' first." ; exit 1 ; \
	fi

test: test-go test-python

# ---------------------------------------------------------------------------
# Python packaging
# ---------------------------------------------------------------------------

wheel: stage-lib
	python3 -m pip wheel . --no-deps --wheel-dir ./dist/

sdist:
	python3 -m build --sdist --outdir ./dist/

# ---------------------------------------------------------------------------
# Docker
# ---------------------------------------------------------------------------

docker-pull:
	# Pull latest base images for stability
	@docker pull golang:latest
	@docker pull python:latest

docker-clean: clean
	@docker compose down --rmi all --volumes --remove-orphans
	@docker container prune -f
	@docker image prune -f
	@docker volume prune -f
	@docker network prune -f

docker-build: docker-pull
	# Build all images without cache
	@docker compose build --no-cache

docker-test:
	@docker compose up --exit-code-from test-python
