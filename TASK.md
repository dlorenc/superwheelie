# superwheelie

Git based Python wheel build system to build wheels for multiple versions of 10k+ projects.

* All packages are tracked in git.
* Every package definition merged to main must build. CI enforces this.
* Packages are stored in GCS.
* Track a list of all packages left to build in the repo.
* An agent atomically deletes a package from this list and adds the build configs for it in one PR.
* Build configs should allow:
    * adding new system packages with apk
    * env vars
    * patches to apply
    * a script fallback mode
* Build configs can be empty (use a default config)
* Build configs can contain overrides for specific packages, versions, or version ranges. Keep config DRY. Think through a nice inheritance model here.
* They must contain a repo URL and a list of versions/ABI targets to build
* Versions should contain a tag/ref mapping in the repo

## Build Agent flow
* Grabs a package from the list
* Downloads it and inspects the repo
* Iterates on the build until it gets it working
* Generates a working config based on that
* Tests the config one more time
* Outputs the config
* Sends a PR
* If we can build some versions/targets but not others, submit that and track the missing ones in another file in the repo. We can make incremental progress. A future agent can iterate here. Include a pointer to a summary of what was tried, what worked, and what failed. This can be stored on GCS.
* This should run inside claude code using the same container outlined below.

## Version completeness agent
* Attempts to add support for more versions, either because new ones have been released or because the first build agent failed on those.
* Use the logs from GCS outlined above.

## Code Review Agent
* Reviews sent PRs and merges them if tests pass.
* Can apply style fixes as well.

## Requirements

* All builds happen in a pre-built container based on cgr.dev/chainguard/wolfi-base. Create a Dockerfile here and store the image in ghcr.io.
* Use the -base targets in the apks (play around to see how they work) for multi-version support.
* Adding new packages to the packages list should trigger builds
* We should have a mechanism to check for new versions and build those.