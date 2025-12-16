package config

// Skips represents known build failures (packages/{name}/skips.yaml).
type Skips struct {
	// Skips is the list of known failures.
	Skips []Skip `yaml:"skips"`
}

// Skip represents a known build failure for a version/Python combination.
type Skip struct {
	// Version is the package version or PEP 440 range that fails.
	Version string `yaml:"version"`

	// Python is the list of Python versions that fail for this package version.
	Python []string `yaml:"python"`

	// Reason is a human-readable explanation of the failure.
	Reason string `yaml:"reason"`

	// Log is the GCS path to the build log for debugging.
	Log string `yaml:"log,omitempty"`

	// Attempts is the number of times the fixer agent has tried.
	Attempts int `yaml:"attempts,omitempty"`
}
