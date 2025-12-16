FROM cgr.dev/chainguard/wolfi-base:latest

# Base build tools (changes rarely)
RUN apk add --no-cache \
    build-base \
    git \
    curl \
    wget

# Python base versions - these can coexist (changes rarely)
RUN apk add --no-cache \
    python-3.10-base \
    python-3.11-base \
    python-3.12-base \
    python-3.13-base

# Python dev headers for native extensions (changes rarely)
RUN apk add --no-cache \
    python-3.10-base-dev \
    python-3.11-base-dev \
    python-3.12-base-dev \
    python-3.13-base-dev

# Pip for each version (changes rarely)
RUN apk add --no-cache \
    py3.10-pip-base \
    py3.11-pip-base \
    py3.12-pip-base \
    py3.13-pip-base

# Build tools
RUN apk add --no-cache \
    py3-build \
    py3-wheel

# Common native build dependencies (add more as needed)
RUN apk add --no-cache \
    openssl-dev \
    libffi-dev \
    zlib-dev \
    bzip2-dev \
    xz-dev \
    readline-dev \
    sqlite-dev \
    ncurses-dev

# Agent tools (for CI/CD workflows)
RUN apk add --no-cache \
    gh \
    google-cloud-sdk \
    nodejs-22 \
    npm

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

WORKDIR /build
