# This Dockerfile creates a C archive and its header file from the Go wrapper
# of Kagome.

FROM golang:latest AS build_base
# This stage (layer) is intended to cache the Go module downloads.

RUN apt update && apt upgrade -y

WORKDIR /app/libkagome/go_wrapper

# We do not `COPY` the local `go.mod` and `go.sum` here because
RUN \
    go mod init github.com/KEINOS/kagome-py/libkagome && \
    go get "github.com/ikawaha/kagome/v2" && \
    go get "github.com/ikawaha/kagome-dict/ipa" && \
    go mod download

# Be careful about the indentation in the heredoc below. It must be a tab character.
COPY --chmod=755 <<'HEREDOC' /app/Makefile
.PHONY: all
all:
	@printf '%s\n' "[!] ERROR: Please mount the repository root directory to /app" >&2
	@exit 1
%: all
	@:
	# unreachable: all always exits with error
HEREDOC

FROM build_base AS build

WORKDIR /app

ENV CGO_ENABLED=1 CC=cc

CMD ["make", "build"]
