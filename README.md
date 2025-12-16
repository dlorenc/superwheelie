# superwheelie

Git-based Python wheel build system for building wheels across multiple Python versions for 10k+ packages.

## Overview

- All package definitions tracked in git
- Every package merged to main must build (CI enforced)
- Built wheels stored in GCS (PyPI-style layout)
- Agent-based workflow: agents claim packages, iterate on builds, submit PRs
- Supports last 4 Python versions (3.10, 3.11, 3.12, 3.13)
- Builds last 10 versions per package by default

## Repository Structure

**main branch:**
```
/
├── queue.txt                    # packages waiting to be built
├── packages/
│   └── {package-name}/
│       ├── config.yaml          # build configuration
│       ├── skips.yaml           # known failures (optional)
│       └── patches/             # optional patches
│           └── *.patch
├── Dockerfile                   # build container (wolfi-base)
├── cmd/
│   ├── build-agent/             # claims packages, iterates on builds, sends PRs
│   │   └── main.go
│   ├── version-agent/           # adds new versions to existing packages
│   │   └── main.go
│   ├── fixer-agent/             # attempts to fix skipped version/python combos
│   │   └── main.go
│   ├── review-agent/            # reviews PRs, merges if tests pass
│   │   └── main.go
│   └── gc/                      # garbage collector for stale claims
│       └── main.go
├── pkg/
│   ├── config/                  # YAML schema types and parsing
│   ├── builder/                 # wheel build orchestration
│   ├── gcs/                     # GCS upload/download
│   └── git/                     # git/GitHub operations
├── go.mod
└── go.sum
```

**claims branch** (ephemeral work tracking):
```
claims/
  numpy.yaml
  requests.yaml
```

## GCS Structure

Wheels and logs stored in PyPI-style layout:

```
gs://bucket/
├── wheels/
│   └── {package}/
│       └── {version}/
│           ├── {package}-{version}-cp310-cp310-linux_x86_64.whl
│           ├── {package}-{version}-cp311-cp311-linux_x86_64.whl
│           └── ...
└── logs/
    └── {package}/
        └── {version}/
            └── build.log
```

## File Formats

### queue.txt

One package name per line:

```
numpy
requests
flask
```

### packages/{name}/config.yaml

See [Build Configuration](#build-configuration) below.

### claims/{name}.yaml (on claims branch)

```yaml
agent: build-agent-abc123
claimed_at: 2025-01-15T10:30:00Z
```

## Build Configuration

### packages/{name}/config.yaml

```yaml
# packages/numpy/config.yaml
repo: https://github.com/numpy/numpy
version_count: 10  # optional, default 10

versions:
  - tag: v2.1.0
    version: 2.1.0
  - tag: v2.0.2
    version: 2.0.2
  - tag: v1.19.0
    version: 1.19.0

# Base build config (all fields below are optional)
system_deps:
  - openblas-dev
  - gcc-gfortran

env:
  CFLAGS: "-O2"

patches:
  - patches/fix-build.patch
  - patches/wolfi-compat.patch

script: |  # if set, replaces normal build entirely
  python setup.py bdist_wheel

# Version overrides (matched in order, first match wins)
overrides:
  - match: ">=2.0"
    system_deps:
      - openblas-dev=0.3.26

  - match: "<1.24"
    env:
      USE_LEGACY_BUILD: "1"
    patches:
      - patches/legacy-fix.patch
```

### packages/{name}/skips.yaml

Known failures that the fixer agent will retry:

```yaml
# packages/numpy/skips.yaml
skips:
  - version: "1.19.0"
    python: ["3.10"]
    reason: "incompatible with Python 3.10 typing changes"
    log: gs://bucket/logs/numpy/1.19.0/3.10/build.log
    attempts: 2

  - version: "<1.18"
    python: ["3.12", "3.13"]
    reason: "requires removed distutils"
    attempts: 1
```

### Config Fields

| Field | Required | Description |
|-------|----------|-------------|
| `repo` | yes | Git repository URL |
| `version_count` | no | Number of versions to build (default: 10) |
| `versions` | yes | List of tag/version mappings |
| `system_deps` | no | APK packages to install (supports pinning: `pkg=1.0`) |
| `env` | no | Environment variables for build |
| `patches` | no | Patches to apply in order |
| `script` | no | Custom build script (replaces default `pip wheel`) |
| `overrides` | no | Version-specific overrides (PEP 440 matching) |

### Skip Fields (skips.yaml)

| Field | Required | Description |
|-------|----------|-------------|
| `version` | yes | Version or PEP 440 range to skip |
| `python` | yes | List of Python versions to skip for this version |
| `reason` | yes | Human-readable explanation of failure |
| `log` | no | GCS path to build log for debugging |
| `attempts` | no | Number of times fixer agent has tried (default: 0) |

### Override Behavior

- **Lists** (system_deps, patches): merged with base config
- **Scalars** (script): replaced entirely
- **Maps** (env): merged (override keys win)
- Overrides matched in order; first match wins per version

## Agents

### Build Agent

The build agent claims packages from the queue and iterates until it produces a working build configuration.

**Flow:**

1. **Select** - Read `queue.txt` from `main`, pick a random package
2. **Claim** - Push `claims/{package}.yaml` to `claims` branch
   - If push fails (file exists), another agent claimed it; go back to step 1
3. **Clone** - Clone the package's source repo (discovered via PyPI API)
4. **Discover versions** - List tags, select last N versions
5. **Iterate on build** - For each version × Python combination:
   - Attempt build with default config (just `pip wheel`)
   - On failure, analyze error and adjust:
     - Missing headers → add `system_deps`
     - Compiler flags → add `env`
     - Build system issues → try `script` override
   - Retry until success or max attempts
6. **Generate config** - Produce minimal `config.yaml` that builds all successful versions
7. **Final validation** - Clean rebuild with generated config
8. **Upload artifacts** - Push wheels and logs to GCS
9. **Submit PR** - Create PR against `main`:
   - Add `packages/{name}/config.yaml`
   - Add `packages/{name}/skips.yaml` if any failures
   - Remove package from `queue.txt`
10. **Release claim** - Delete `claims/{package}.yaml` from `claims` branch

**Partial success:** If only some versions/Python combinations succeed, the agent submits what works. Failed combinations are added to `skips.yaml` with reason and log pointer for the fixer agent to retry later.

**Environment:** Runs inside the wolfi-base build container with Claude Code for iterative debugging.

### Garbage Collector

Recovers packages from crashed or abandoned agents.

**Flow:**

1. Scan `claims/` directory on `claims` branch
2. For each claim older than TTL (default: 4 hours):
   - Delete the claim file from `claims` branch
   - If package not in `packages/` on `main`, ensure it's in `queue.txt`

**Deployment:** Runs on a cron schedule (e.g., every 30 minutes).

### Version Agent

Adds new versions to existing packages when upstream releases them.

**Flow:**

1. **Select** - Pick a random package from `packages/` that has available upstream versions not in config
2. **Claim** - Push `claims/{package}.yaml` to `claims` branch (type: `version`)
3. **Discover** - Query upstream repo for new tags since last tracked version
4. **Build** - Attempt to build new versions using existing config
5. **Update config** - Add successful versions to `versions` list, failures to `skips.yaml`
6. **Submit PR** - Create PR with updated `config.yaml`
7. **Release claim** - Delete claim file

**Trigger:** Runs periodically or on-demand to pick up new releases.

### Fixer Agent

Attempts to fix skipped version/Python combinations.

**Flow:**

1. **Select** - Scan `packages/` for `skips.yaml` files, prioritize low `attempts` count
2. **Claim** - Push `claims/{package}.yaml` to `claims` branch (type: `fixer`)
3. **Analyze** - Read the skip's `reason` and fetch `log` from GCS
4. **Iterate** - Try different approaches to fix the build:
   - Add missing `system_deps`
   - Add version-specific `overrides`
   - Create a `patch` for source compatibility
   - Try alternative build `script`
5. **On success**:
   - Remove entry from `skips.yaml` (delete file if empty)
   - Add any required `overrides` or `patches` to `config.yaml`
   - Upload wheels to GCS
6. **On failure** - Update skip's `reason` with new findings, increment `attempts`
7. **Submit PR** - Create PR with updated config/skips
8. **Release claim** - Delete claim file

**Strategy:** Uses previous build logs and error messages to guide fixes. May try:
- Backporting patches from newer versions
- Adding compatibility shims for removed stdlib modules (e.g., distutils → setuptools)
- Adjusting compiler flags for older code

### Review Agent

Reviews and merges PRs from other agents.

**Flow:**

1. **List PRs** - Query open PRs from build/version/fixer agents
2. **For each PR**:
   - Check CI status (must pass)
   - Validate config schema
   - Verify wheels exist in GCS
   - Apply style fixes if needed
   - Approve and merge

**Deployment:** Runs continuously or on PR webhook.

## Claude Code Integration

Agents delegate iterative debugging to Claude Code via CLI subprocess.

### Invocation

```go
cmd := exec.Command("claude",
    "-p", prompt,
    "--output-format", "json",
    "--allowedTools", "Bash,Read,Edit,Write",
    "--permission-mode", "acceptEdits",
    "--max-turns", "30",
)
cmd.Dir = workDir
```

### JSON Response

```json
{
  "type": "result",
  "subtype": "success",
  "is_error": false,
  "result": "Built 8/10 versions successfully...",
  "num_turns": 24,
  "session_id": "abc123"
}
```

### Session Resumption

If Claude hits max turns without completing, resume up to 3 times:

```go
cmd := exec.Command("claude",
    "-p", "Continue where you left off",
    "--resume", sessionID,
    "--output-format", "json",
    // ... same flags
)
```

### Build Agent Prompt

```
You are building Python wheels for package "{name}".

Source cloned to: /build/{name}/src
Output directory: /build/{name}/dist
Python versions: 3.10, 3.11, 3.12, 3.13

Versions to build (newest first):
- v2.1.0 (2.1.0)
- v2.0.2 (2.0.2)
- v2.0.1 (2.0.1)
... (10 versions)

For each version × Python combination:
1. Check out the tag
2. Run: pip wheel --no-deps -w ../dist .
3. If it fails, try fixes:
   - Install system packages: apk add <pkg>
   - Set environment variables
   - Apply source patches
4. Move on after 3 failed attempts per combination

Partial success is fine - build what you can.

When done, create two files in /build/{name}/output:

1. config.yaml with this structure:
   repo: https://github.com/...
   versions:
     - tag: v2.1.0
       version: 2.1.0
     ...only versions that built for ALL Python versions...
   system_deps: [...if any...]
   env: {...if any...}
   patches: [...if any...]
   overrides: [...if needed for specific versions...]

2. skips.yaml (if any failures):
   skips:
     - version: "1.19.0"
       python: ["3.10"]
       reason: "error message summary"
       attempts: 3

Output a final summary of what worked and what didn't.
```

### Fixer Agent Prompt

```
You are fixing a failed Python wheel build.

Package: {name}
Source: /build/{name}/src
Previous failure info:
  Version: {version}
  Python: {python_versions}
  Reason: {reason}
  Log: (contents of build log)

The existing config.yaml is at /build/{name}/config.yaml

Try to fix this build. Approaches:
1. Add missing system dependencies
2. Add version-specific overrides
3. Create a patch file in /build/{name}/patches/
4. Try alternative build commands

If you succeed:
- Update config.yaml with any new overrides/patches
- Remove this entry from skips.yaml

If you fail after 5 attempts:
- Update skips.yaml with new reason and increment attempts

Output what you tried and the result.
```

### Output Parsing

Go agent reads files from `/build/{name}/output/`:
- `config.yaml` → validates and includes in PR
- `skips.yaml` → includes in PR if present
- `patches/*.patch` → includes in PR if created

The agent does NOT parse Claude's conversational output - only the files it generates.
