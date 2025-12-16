// Package builder provides wheel build orchestration for Python packages.
package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dlorenc/superwheelie/pkg/config"
)

// Builder orchestrates wheel builds for a package.
type Builder struct {
	// WorkDir is the base working directory for builds.
	WorkDir string

	// Config is the package build configuration.
	Config *config.Config

	// PackageName is the name of the package being built.
	PackageName string

	// SourceDir is the directory containing the cloned source.
	SourceDir string

	// DistDir is the directory where wheels are output.
	DistDir string
}

// BuildResult contains the result of building a single version/Python combination.
type BuildResult struct {
	// Version is the package version that was built.
	Version string

	// Python is the Python version used (e.g., "3.12").
	Python string

	// WheelPath is the path to the built wheel file, if successful.
	WheelPath string

	// Success indicates whether the build succeeded.
	Success bool

	// Log contains the build output.
	Log string

	// Error contains any error that occurred.
	Error error
}

// New creates a new Builder for a package.
func New(workDir, packageName string, cfg *config.Config) *Builder {
	return &Builder{
		WorkDir:     workDir,
		Config:      cfg,
		PackageName: packageName,
		SourceDir:   filepath.Join(workDir, "src"),
		DistDir:     filepath.Join(workDir, "dist"),
	}
}

// Setup prepares the build environment by creating directories.
func (b *Builder) Setup() error {
	for _, dir := range []string{b.SourceDir, b.DistDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}
	return nil
}

// CloneSource clones the source repository.
func (b *Builder) CloneSource() error {
	if b.Config.Repo == "" {
		return fmt.Errorf("no repo URL configured")
	}

	cmd := exec.Command("git", "clone", "--depth", "1", b.Config.Repo, b.SourceDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cloning repo: %w\n%s", err, output)
	}
	return nil
}

// Checkout checks out a specific tag/ref in the source directory.
func (b *Builder) Checkout(ref string) error {
	cmd := exec.Command("git", "fetch", "--depth", "1", "origin", "tag", ref)
	cmd.Dir = b.SourceDir
	if output, err := cmd.CombinedOutput(); err != nil {
		// Try fetching as a regular ref if tag fetch fails
		cmd = exec.Command("git", "fetch", "--depth", "1", "origin", ref)
		cmd.Dir = b.SourceDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("fetching ref %s: %w\n%s", ref, err, output)
		}
		_ = output
	}

	cmd = exec.Command("git", "checkout", "FETCH_HEAD")
	cmd.Dir = b.SourceDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("checking out %s: %w\n%s", ref, err, output)
	}

	// Clean any untracked files from previous builds
	cmd = exec.Command("git", "clean", "-fdx")
	cmd.Dir = b.SourceDir
	cmd.Run() // Ignore errors

	return nil
}

// InstallSystemDeps installs system dependencies via apk.
func (b *Builder) InstallSystemDeps(deps []string) error {
	if len(deps) == 0 {
		return nil
	}

	args := append([]string{"add", "--no-cache"}, deps...)
	cmd := exec.Command("apk", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installing system deps: %w\n%s", err, output)
	}
	return nil
}

// ApplyPatches applies patch files in order.
func (b *Builder) ApplyPatches(patches []string) error {
	for _, patch := range patches {
		patchPath := filepath.Join(b.WorkDir, patch)
		cmd := exec.Command("git", "apply", patchPath)
		cmd.Dir = b.SourceDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("applying patch %s: %w\n%s", patch, err, output)
		}
	}
	return nil
}

// Build builds wheels for a specific version across all Python versions.
func (b *Builder) Build(version config.Version, pythonVersions []string) []BuildResult {
	results := make([]BuildResult, 0, len(pythonVersions))

	// Checkout the tag
	if err := b.Checkout(version.Tag); err != nil {
		// Return failure for all Python versions
		for _, py := range pythonVersions {
			results = append(results, BuildResult{
				Version: version.Version,
				Python:  py,
				Success: false,
				Log:     err.Error(),
				Error:   err,
			})
		}
		return results
	}

	// Get effective config for this version (apply overrides)
	effectiveCfg := b.getEffectiveConfig(version.Version)

	// Install system dependencies
	if err := b.InstallSystemDeps(effectiveCfg.SystemDeps); err != nil {
		for _, py := range pythonVersions {
			results = append(results, BuildResult{
				Version: version.Version,
				Python:  py,
				Success: false,
				Log:     err.Error(),
				Error:   err,
			})
		}
		return results
	}

	// Apply patches
	if err := b.ApplyPatches(effectiveCfg.Patches); err != nil {
		for _, py := range pythonVersions {
			results = append(results, BuildResult{
				Version: version.Version,
				Python:  py,
				Success: false,
				Log:     err.Error(),
				Error:   err,
			})
		}
		return results
	}

	// Build for each Python version
	for _, py := range pythonVersions {
		result := b.buildForPython(version.Version, py, effectiveCfg)
		results = append(results, result)
	}

	return results
}

// buildForPython builds a wheel for a specific Python version.
func (b *Builder) buildForPython(version, python string, cfg *effectiveConfig) BuildResult {
	result := BuildResult{
		Version: version,
		Python:  python,
	}

	var logBuf bytes.Buffer
	var cmd *exec.Cmd

	pythonBin := PythonBinary(python)

	if cfg.Script != "" {
		// Use custom script
		cmd = exec.Command("sh", "-c", cfg.Script)
	} else {
		// Default pip wheel command
		cmd = exec.Command(pythonBin, "-m", "pip", "wheel",
			"--no-deps",
			"--no-binary", ":all:",
			"-w", b.DistDir,
			".")
	}

	cmd.Dir = b.SourceDir
	cmd.Env = b.buildEnv(cfg.Env, python)
	cmd.Stdout = &logBuf
	cmd.Stderr = &logBuf

	err := cmd.Run()
	result.Log = logBuf.String()

	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("build failed: %w", err)
		return result
	}

	// Find the built wheel
	wheelPath, err := b.findWheel(version, python)
	if err != nil {
		result.Success = false
		result.Error = err
		result.Log += "\n" + err.Error()
		return result
	}

	result.Success = true
	result.WheelPath = wheelPath
	return result
}

// buildEnv constructs the environment for a build.
func (b *Builder) buildEnv(env map[string]string, python string) []string {
	// Start with current environment
	result := os.Environ()

	// Add configured environment variables
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}

	// Ensure the correct Python is used
	pythonBin := PythonBinary(python)
	pythonDir := filepath.Dir(pythonBin)
	for i, e := range result {
		if strings.HasPrefix(e, "PATH=") {
			result[i] = fmt.Sprintf("PATH=%s:%s", pythonDir, e[5:])
			break
		}
	}

	return result
}

// findWheel finds the built wheel file for a version/Python combination.
func (b *Builder) findWheel(version, python string) (string, error) {
	cpVersion := "cp" + strings.Replace(python, ".", "", 1)
	pattern := filepath.Join(b.DistDir, fmt.Sprintf("*-%s-%s-*.whl", cpVersion, cpVersion))

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("searching for wheel: %w", err)
	}

	// Look for a wheel matching this version
	for _, m := range matches {
		base := filepath.Base(m)
		// Wheel filename format: {name}-{version}-{python}-{abi}-{platform}.whl
		if strings.Contains(base, "-"+version+"-") || strings.Contains(base, "-"+strings.Replace(version, ".", "_", -1)+"-") {
			return m, nil
		}
	}

	// If no exact match, return any wheel with matching Python version
	if len(matches) > 0 {
		return matches[0], nil
	}

	return "", fmt.Errorf("no wheel found for Python %s", python)
}

// effectiveConfig holds the merged configuration for a specific version.
type effectiveConfig struct {
	SystemDeps []string
	Env        map[string]string
	Patches    []string
	Script     string
}

// getEffectiveConfig merges base config with version-specific overrides.
func (b *Builder) getEffectiveConfig(version string) *effectiveConfig {
	cfg := &effectiveConfig{
		SystemDeps: append([]string{}, b.Config.SystemDeps...),
		Env:        make(map[string]string),
		Patches:    append([]string{}, b.Config.Patches...),
		Script:     b.Config.Script,
	}

	// Copy base env
	for k, v := range b.Config.Env {
		cfg.Env[k] = v
	}

	// Apply overrides
	for _, override := range b.Config.Overrides {
		matches, err := config.MatchesVersion(version, override.Match)
		if err != nil || !matches {
			continue
		}

		// Merge lists
		cfg.SystemDeps = append(cfg.SystemDeps, override.SystemDeps...)
		cfg.Patches = append(cfg.Patches, override.Patches...)

		// Merge env (override wins)
		for k, v := range override.Env {
			cfg.Env[k] = v
		}

		// Replace script
		if override.Script != "" {
			cfg.Script = override.Script
		}

		// First match wins
		break
	}

	return cfg
}

// BuildAll builds all configured versions for all Python versions.
func (b *Builder) BuildAll(pythonVersions []string) map[string][]BuildResult {
	results := make(map[string][]BuildResult)

	for _, v := range b.Config.Versions {
		results[v.Version] = b.Build(v, pythonVersions)
	}

	return results
}
