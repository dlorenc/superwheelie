package config

import "time"

// Claim represents a package claim on the claims branch (claims/{name}.yaml).
type Claim struct {
	// Agent is the identifier of the agent that claimed the package.
	Agent string `yaml:"agent"`

	// ClaimedAt is when the package was claimed.
	ClaimedAt time.Time `yaml:"claimed_at"`

	// Type is the type of claim (build, version, fixer).
	Type string `yaml:"type,omitempty"`
}

// Claim types.
const (
	ClaimTypeBuild   = "build"
	ClaimTypeVersion = "version"
	ClaimTypeFixer   = "fixer"
)
