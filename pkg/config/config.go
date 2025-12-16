// Package config provides types and parsing for superwheelie configuration files.
package config

// Config represents a package build configuration (packages/{name}/config.yaml).
type Config struct {
	// Repo is the Git repository URL for the package source.
	Repo string `yaml:"repo"`

	// VersionCount is the number of versions to build (default: 10).
	VersionCount int `yaml:"version_count,omitempty"`

	// Versions is the list of tag/version mappings to build.
	Versions []Version `yaml:"versions"`

	// SystemDeps are APK packages to install before building.
	// Supports pinning: "pkg=1.0"
	SystemDeps []string `yaml:"system_deps,omitempty"`

	// Env contains environment variables to set during build.
	Env map[string]string `yaml:"env,omitempty"`

	// Patches is a list of patch files to apply in order.
	Patches []string `yaml:"patches,omitempty"`

	// Script is a custom build script that replaces the default pip wheel command.
	Script string `yaml:"script,omitempty"`

	// Overrides contains version-specific build configuration overrides.
	Overrides []Override `yaml:"overrides,omitempty"`
}

// Version represents a tag-to-version mapping.
type Version struct {
	// Tag is the git tag or ref to checkout.
	Tag string `yaml:"tag"`

	// Version is the PyPI version string.
	Version string `yaml:"version"`
}

// Override represents version-specific build configuration.
// Overrides are matched in order using PEP 440 version specifiers.
type Override struct {
	// Match is a PEP 440 version specifier (e.g., ">=2.0", "<1.24", "==1.19.5").
	Match string `yaml:"match"`

	// SystemDeps are additional APK packages (merged with base config).
	SystemDeps []string `yaml:"system_deps,omitempty"`

	// Env contains additional environment variables (merged with base config).
	Env map[string]string `yaml:"env,omitempty"`

	// Patches are additional patch files (merged with base config).
	Patches []string `yaml:"patches,omitempty"`

	// Script replaces the base script entirely.
	Script string `yaml:"script,omitempty"`
}

// DefaultVersionCount is the default number of versions to build.
const DefaultVersionCount = 10
