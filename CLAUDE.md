# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Superwheelie is a Git-based Python wheel build system designed to build wheels for 10k+ packages across multiple Python versions (3.10-3.13). It uses an agent-based architecture where Go agents orchestrate builds and delegate iterative debugging to Claude Code.

## Build Commands

```bash
# Build the Docker container
docker build -t superwheelie-build .

# Run Go tests (once implemented)
go test ./...

# Build all agents
go build ./cmd/...

# Build specific agent
go build ./cmd/build-agent
```

## Architecture

### Two-Branch Model
- **main**: Package configs, queue, code
- **claims**: Ephemeral work tracking (agents claim packages here to prevent races)

### Agent System
Go agents in `cmd/` orchestrate work, invoke Claude Code CLI for iterative debugging:
- **build-agent**: Claims packages from queue.txt, produces config.yaml
- **version-agent**: Adds new upstream versions to existing packages
- **fixer-agent**: Fixes skipped version/Python combinations in skips.yaml
- **review-agent**: Validates and merges agent PRs
- **gc**: Recovers packages from crashed agents (TTL-based)

### Shared Libraries (pkg/)
- **config**: YAML schema types (Config, Skips, Claims, Override)
- **builder**: Wheel build orchestration
- **gcs**: Upload wheels/logs to gs://dlorenc-superwheelie
- **git**: Git operations, claims branch, PR creation via gh CLI

### Claude Code Integration
Agents invoke Claude Code as subprocess with `--print --output-format json`. Claude generates config.yaml/skips.yaml files that agents read back (file-based contract, not parsing conversational output).

## Key Config Schemas

**config.yaml**: repo, versions (tag→version list), system_deps, env, patches, script, overrides (PEP 440 matching)

**skips.yaml**: Known failures with version, python versions, reason, log path, attempts count

**Override behavior**: Lists merge, scalars replace, maps merge (override wins)

## Container Environment

Wolfi-base with coexisting Python versions via `-base` packages:
- `python3.10`, `python3.11`, `python3.12`, `python3.13`
- Wheels tagged `linux_aarch64` (Wolfi glibc, not manylinux)

## GCS Layout

```
gs://dlorenc-superwheelie/
├── wheels/{package}/{version}/{wheel}.whl
└── logs/{package}/{version}/{python}/build.log
```
