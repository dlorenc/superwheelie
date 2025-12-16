package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig reads and parses a config.yaml file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Apply defaults
	if cfg.VersionCount == 0 {
		cfg.VersionCount = DefaultVersionCount
	}

	return &cfg, nil
}

// SaveConfig writes a Config to a YAML file.
func SaveConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// LoadSkips reads and parses a skips.yaml file.
func LoadSkips(path string) (*Skips, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Skips{}, nil
		}
		return nil, fmt.Errorf("reading skips file: %w", err)
	}

	var skips Skips
	if err := yaml.Unmarshal(data, &skips); err != nil {
		return nil, fmt.Errorf("parsing skips file: %w", err)
	}

	return &skips, nil
}

// SaveSkips writes a Skips to a YAML file.
// If skips is empty, the file is removed.
func SaveSkips(skips *Skips, path string) error {
	if len(skips.Skips) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing empty skips file: %w", err)
		}
		return nil
	}

	data, err := yaml.Marshal(skips)
	if err != nil {
		return fmt.Errorf("marshaling skips: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing skips file: %w", err)
	}

	return nil
}

// LoadClaim reads and parses a claim file from the claims branch.
func LoadClaim(path string) (*Claim, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading claim file: %w", err)
	}

	var claim Claim
	if err := yaml.Unmarshal(data, &claim); err != nil {
		return nil, fmt.Errorf("parsing claim file: %w", err)
	}

	return &claim, nil
}

// SaveClaim writes a Claim to a YAML file.
func SaveClaim(claim *Claim, path string) error {
	data, err := yaml.Marshal(claim)
	if err != nil {
		return fmt.Errorf("marshaling claim: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing claim file: %w", err)
	}

	return nil
}

// LoadPackageConfig loads a package's config.yaml from the packages directory.
func LoadPackageConfig(packagesDir, packageName string) (*Config, error) {
	path := filepath.Join(packagesDir, packageName, "config.yaml")
	return LoadConfig(path)
}

// LoadPackageSkips loads a package's skips.yaml from the packages directory.
func LoadPackageSkips(packagesDir, packageName string) (*Skips, error) {
	path := filepath.Join(packagesDir, packageName, "skips.yaml")
	return LoadSkips(path)
}
