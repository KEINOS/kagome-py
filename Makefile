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

.PHONY: clean build-archive build-shared build test-go test-python test clean

clean:
	@rm -rf ./libkagome/bin

build-archive:
	@mkdir -p ./libkagome/bin
	@rm -rf ./libkagome/bin/lib*
	@cd ./libkagome/go_wrapper && \
	go clean -testcache && go test -race -v ./... && \
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=c-archive -o ../bin/libkagome.a ./main.go

build-shared: build-archive
	@cd ./libkagome && \
	cc -fPIC \
		-fvisibility=hidden \
		-shared \
		./c_wrapper/kagome_wrapper.c ./bin/libkagome.a \
		-o ./bin/libkagome.$(LIB_EXT)

build: build-shared
	@mkdir -p ./libkagome/dist
	@rm -rf ./libkagome/dist/libkagome*
	@cp ./libkagome/bin/libkagome.$(LIB_EXT) ./libkagome/dist/libkagome.$(LIB_EXT)
	@cp ./libkagome/python_wrapper/libkagome.py ./libkagome/dist/libkagome.py

test-go:
	@if ! command -v go >/dev/null 2>&1; then \
		echo "[SKIP] Go is not installed."; \
		exit 0; \
	fi; \
	echo "**NOTE**: On macOS, 'ld: warning: ...' can be ignored if tests pass."; \
	cd ./libkagome/go_wrapper && \
	go clean -testcache && go test -race ./...

test-python:
	@if ! command -v python3 >/dev/null 2>&1; then \
		echo "[SKIP] Python3 is not installed."; \
		exit 0; \
	fi; \
	if [ -f ./libkagome/bin/libkagome.so ] || [ -f ./libkagome/bin/libkagome.dll ] || [ -f ./libkagome/bin/libkagome.dylib ]; then \
		PYTHONPATH=./libkagome/python_wrapper python3 ./tests/libkagome_test.py ; \
	else \
		echo "No libkagome.(so|dll|dylib) found. Please run make build first." ; exit 1 ; \
	fi

test: test-go test-python


.PHONY: docker-pull docker-clean docker-build docker-test

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
