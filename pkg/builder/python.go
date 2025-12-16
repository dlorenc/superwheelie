package builder

import (
	"fmt"
	"os/exec"
	"strings"
)

// SupportedPythonVersions is the list of Python versions we build for.
var SupportedPythonVersions = []string{"3.10", "3.11", "3.12", "3.13"}

// PythonBinary returns the path to the Python binary for a version.
func PythonBinary(version string) string {
	return fmt.Sprintf("/usr/bin/python%s", version)
}

// PythonCPVersion returns the CPython version tag (e.g., "cp312").
func PythonCPVersion(version string) string {
	return "cp" + strings.Replace(version, ".", "", 1)
}

// PythonABI returns the ABI tag for a Python version (e.g., "cp312").
func PythonABI(version string) string {
	// For CPython 3.8+, ABI tag matches the Python version
	return PythonCPVersion(version)
}

// WheelFilename generates the expected wheel filename.
func WheelFilename(packageName, version, pythonVersion, platform string) string {
	// Normalize package name (PEP 427: replace - and . with _)
	normalized := strings.ReplaceAll(packageName, "-", "_")
	normalized = strings.ReplaceAll(normalized, ".", "_")

	// Normalize version (replace - with _)
	normalizedVersion := strings.ReplaceAll(version, "-", "_")

	cp := PythonCPVersion(pythonVersion)
	abi := PythonABI(pythonVersion)

	return fmt.Sprintf("%s-%s-%s-%s-%s.whl", normalized, normalizedVersion, cp, abi, platform)
}

// DefaultPlatform is the default platform tag for wheels built in the container.
const DefaultPlatform = "linux_aarch64"

// IsPythonAvailable checks if a Python version is available.
func IsPythonAvailable(version string) bool {
	bin := PythonBinary(version)
	cmd := exec.Command(bin, "--version")
	return cmd.Run() == nil
}

// GetAvailablePythonVersions returns the list of available Python versions.
func GetAvailablePythonVersions() []string {
	available := make([]string, 0, len(SupportedPythonVersions))
	for _, v := range SupportedPythonVersions {
		if IsPythonAvailable(v) {
			available = append(available, v)
		}
	}
	return available
}

// PythonVersionInfo holds information about a Python version.
type PythonVersionInfo struct {
	Version   string // e.g., "3.12"
	Binary    string // e.g., "/usr/bin/python3.12"
	CPVersion string // e.g., "cp312"
	ABI       string // e.g., "cp312"
	Available bool
}

// GetPythonInfo returns information about a Python version.
func GetPythonInfo(version string) PythonVersionInfo {
	return PythonVersionInfo{
		Version:   version,
		Binary:    PythonBinary(version),
		CPVersion: PythonCPVersion(version),
		ABI:       PythonABI(version),
		Available: IsPythonAvailable(version),
	}
}
