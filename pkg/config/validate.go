package config

import (
	"fmt"
	"regexp"
	"strings"
)

// pep440Pattern matches PEP 440 version specifiers.
// Supports: ==, !=, <, <=, >, >=, ~=, and combinations with commas.
var pep440Pattern = regexp.MustCompile(`^([<>=!~]+\s*[\d\w.*]+)(,\s*[<>=!~]+\s*[\d\w.*]+)*$`)

// ValidateConfig validates a Config for required fields and correct formats.
func ValidateConfig(cfg *Config) error {
	if cfg.Repo == "" {
		return fmt.Errorf("repo is required")
	}

	if len(cfg.Versions) == 0 {
		return fmt.Errorf("at least one version is required")
	}

	seen := make(map[string]bool)
	for i, v := range cfg.Versions {
		if v.Tag == "" {
			return fmt.Errorf("version[%d]: tag is required", i)
		}
		if v.Version == "" {
			return fmt.Errorf("version[%d]: version is required", i)
		}
		if seen[v.Version] {
			return fmt.Errorf("version[%d]: duplicate version %q", i, v.Version)
		}
		seen[v.Version] = true
	}

	for i, o := range cfg.Overrides {
		if o.Match == "" {
			return fmt.Errorf("override[%d]: match is required", i)
		}
		if !isValidPEP440(o.Match) {
			return fmt.Errorf("override[%d]: invalid PEP 440 specifier %q", i, o.Match)
		}
	}

	return nil
}

// ValidateSkips validates a Skips for required fields.
func ValidateSkips(skips *Skips) error {
	for i, s := range skips.Skips {
		if s.Version == "" {
			return fmt.Errorf("skip[%d]: version is required", i)
		}
		if len(s.Python) == 0 {
			return fmt.Errorf("skip[%d]: at least one python version is required", i)
		}
		if s.Reason == "" {
			return fmt.Errorf("skip[%d]: reason is required", i)
		}
	}
	return nil
}

// ValidateClaim validates a Claim for required fields.
func ValidateClaim(claim *Claim) error {
	if claim.Agent == "" {
		return fmt.Errorf("agent is required")
	}
	if claim.ClaimedAt.IsZero() {
		return fmt.Errorf("claimed_at is required")
	}
	return nil
}

// isValidPEP440 checks if a string is a valid PEP 440 version specifier.
func isValidPEP440(spec string) bool {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return false
	}
	return pep440Pattern.MatchString(spec)
}

// MatchesVersion checks if a version matches a PEP 440 specifier.
// This is a simplified implementation that handles common cases.
func MatchesVersion(version, specifier string) (bool, error) {
	specifier = strings.TrimSpace(specifier)
	version = strings.TrimSpace(version)

	// Handle comma-separated specifiers
	parts := strings.Split(specifier, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		matches, err := matchSingleSpec(version, part)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}
	return true, nil
}

// matchSingleSpec matches a version against a single specifier.
func matchSingleSpec(version, spec string) (bool, error) {
	spec = strings.TrimSpace(spec)

	// Extract operator and version
	var op, specVer string
	for _, prefix := range []string{"==", "!=", "<=", ">=", "<", ">", "~="} {
		if strings.HasPrefix(spec, prefix) {
			op = prefix
			specVer = strings.TrimSpace(spec[len(prefix):])
			break
		}
	}

	if op == "" {
		return false, fmt.Errorf("invalid specifier: %q", spec)
	}

	cmp := compareVersions(version, specVer)

	switch op {
	case "==":
		return cmp == 0, nil
	case "!=":
		return cmp != 0, nil
	case "<":
		return cmp < 0, nil
	case "<=":
		return cmp <= 0, nil
	case ">":
		return cmp > 0, nil
	case ">=":
		return cmp >= 0, nil
	case "~=":
		// Compatible release: ~=X.Y means >=X.Y, ==X.*
		if cmp < 0 {
			return false, nil
		}
		// Check prefix match
		parts := strings.Split(specVer, ".")
		if len(parts) > 1 {
			prefix := strings.Join(parts[:len(parts)-1], ".")
			return strings.HasPrefix(version, prefix), nil
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported operator: %q", op)
	}
}

// compareVersions compares two version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aVal, bVal int
		if i < len(aParts) {
			fmt.Sscanf(aParts[i], "%d", &aVal)
		}
		if i < len(bParts) {
			fmt.Sscanf(bParts[i], "%d", &bVal)
		}

		if aVal < bVal {
			return -1
		}
		if aVal > bVal {
			return 1
		}
	}
	return 0
}
