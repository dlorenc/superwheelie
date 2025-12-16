# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Superwheelie is a Git-based Python wheel build system designed to build wheels for 10k+ packages across multiple Python versions (3.10-3.13). It uses an agent-based architecture where Go agents orchestrate builds and delegate iterative debugging to Claude Code.

## Build Commands

```bash
# Build the Docker container
docker build -t superwheelie-build .

# Run Go tests
go test ./...

# Build all agents (once implemented)
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

**Implemented:**
- **pkg/config**: YAML schema types and parsing
  - `Config`, `Version`, `Override` types (config.yaml)
  - `Skips`, `Skip` types (skips.yaml)
  - `Claim` type (claims branch)
  - `LoadConfig`, `SaveConfig`, `LoadSkips`, `SaveSkips`, `LoadClaim`, `SaveClaim`
  - `ValidateConfig`, `ValidateSkips`, `ValidateClaim`
  - `MatchesVersion` for PEP 440 version specifier matching

- **pkg/builder**: Wheel build orchestration
  - `Builder` type with `New`, `Setup`, `Build`, `BuildAll`
  - `CloneSource`, `Checkout`, `InstallSystemDeps`, `ApplyPatches`
  - `BuildResult` type with Version, Python, WheelPath, Success, Log, Error
  - Config override merging (PEP 440 matching)
  - `PythonBinary`, `PythonCPVersion`, `WheelFilename` helpers
  - `Exec`, `ExecWithTimeout`, `ExecSimple` command helpers

**Not yet implemented:**
- **pkg/gcs**: Upload wheels/logs to gs://dlorenc-superwheelie (issue #3)
- **pkg/git**: Git operations, claims branch, PR creation (issue #4)
- **pkg/claude**: Claude Code CLI integration (issue #5)

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

## Current Status

**Completed:**
- Issue #1: pkg/config - YAML types and parsing ✓
- Issue #2: pkg/builder - Wheel build orchestration ✓

**Open issues (in suggested order):**
- Issue #3: pkg/gcs - GCS upload/download utilities
- Issue #4: pkg/git - Git and GitHub operations
- Issue #5: pkg/claude - Claude Code CLI integration
- Issue #6: cmd/build-agent - Build agent implementation
- Issue #7: cmd/version-agent - Version agent implementation
- Issue #8: cmd/fixer-agent - Fixer agent implementation
- Issue #9: cmd/review-agent - Review agent implementation
- Issue #10: cmd/gc - Garbage collector implementation
- Issue #11: CI - GitHub Actions workflows
- Issue #12: Bootstrap - Initial queue.txt with packages

## Infrastructure

- **GitHub repo**: https://github.com/dlorenc/superwheelie
- **GCS bucket**: gs://dlorenc-superwheelie
- **OIDC**: GitHub Actions can authenticate via Workload Identity Federation
- **Service account**: superwheelie-gha@dlorenc-chainguard.iam.gserviceaccount.com
