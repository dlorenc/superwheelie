package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := `repo: https://github.com/numpy/numpy
version_count: 5
versions:
  - tag: v2.1.0
    version: 2.1.0
  - tag: v2.0.0
    version: 2.0.0
system_deps:
  - openblas-dev
env:
  CFLAGS: "-O2"
patches:
  - patches/fix.patch
overrides:
  - match: ">=2.0"
    system_deps:
      - extra-dep
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Repo != "https://github.com/numpy/numpy" {
		t.Errorf("Repo = %q, want %q", cfg.Repo, "https://github.com/numpy/numpy")
	}
	if cfg.VersionCount != 5 {
		t.Errorf("VersionCount = %d, want 5", cfg.VersionCount)
	}
	if len(cfg.Versions) != 2 {
		t.Errorf("len(Versions) = %d, want 2", len(cfg.Versions))
	}
	if cfg.Versions[0].Tag != "v2.1.0" {
		t.Errorf("Versions[0].Tag = %q, want %q", cfg.Versions[0].Tag, "v2.1.0")
	}
	if len(cfg.SystemDeps) != 1 || cfg.SystemDeps[0] != "openblas-dev" {
		t.Errorf("SystemDeps = %v, want [openblas-dev]", cfg.SystemDeps)
	}
	if cfg.Env["CFLAGS"] != "-O2" {
		t.Errorf("Env[CFLAGS] = %q, want %q", cfg.Env["CFLAGS"], "-O2")
	}
	if len(cfg.Overrides) != 1 || cfg.Overrides[0].Match != ">=2.0" {
		t.Errorf("Overrides = %v, want [{Match: >=2.0, ...}]", cfg.Overrides)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := `repo: https://github.com/example/pkg
versions:
  - tag: v1.0.0
    version: 1.0.0
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.VersionCount != DefaultVersionCount {
		t.Errorf("VersionCount = %d, want default %d", cfg.VersionCount, DefaultVersionCount)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "pkg", "config.yaml")

	cfg := &Config{
		Repo:         "https://github.com/test/pkg",
		VersionCount: 10,
		Versions: []Version{
			{Tag: "v1.0.0", Version: "1.0.0"},
		},
		SystemDeps: []string{"libfoo"},
	}

	if err := SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.Repo != cfg.Repo {
		t.Errorf("Repo = %q, want %q", loaded.Repo, cfg.Repo)
	}
	if len(loaded.Versions) != 1 {
		t.Errorf("len(Versions) = %d, want 1", len(loaded.Versions))
	}
}

func TestLoadSkips(t *testing.T) {
	dir := t.TempDir()
	skipsPath := filepath.Join(dir, "skips.yaml")

	content := `skips:
  - version: "1.19.0"
    python: ["3.10", "3.11"]
    reason: "incompatible typing"
    log: gs://bucket/logs/test.log
    attempts: 2
`
	if err := os.WriteFile(skipsPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	skips, err := LoadSkips(skipsPath)
	if err != nil {
		t.Fatalf("LoadSkips failed: %v", err)
	}

	if len(skips.Skips) != 1 {
		t.Errorf("len(Skips) = %d, want 1", len(skips.Skips))
	}
	if skips.Skips[0].Version != "1.19.0" {
		t.Errorf("Skips[0].Version = %q, want %q", skips.Skips[0].Version, "1.19.0")
	}
	if len(skips.Skips[0].Python) != 2 {
		t.Errorf("len(Skips[0].Python) = %d, want 2", len(skips.Skips[0].Python))
	}
	if skips.Skips[0].Attempts != 2 {
		t.Errorf("Skips[0].Attempts = %d, want 2", skips.Skips[0].Attempts)
	}
}

func TestLoadSkipsNotExist(t *testing.T) {
	skips, err := LoadSkips("/nonexistent/skips.yaml")
	if err != nil {
		t.Fatalf("LoadSkips failed: %v", err)
	}
	if len(skips.Skips) != 0 {
		t.Errorf("len(Skips) = %d, want 0", len(skips.Skips))
	}
}

func TestSaveSkipsEmpty(t *testing.T) {
	dir := t.TempDir()
	skipsPath := filepath.Join(dir, "skips.yaml")

	// Create a file first
	if err := os.WriteFile(skipsPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Save empty skips should remove the file
	if err := SaveSkips(&Skips{}, skipsPath); err != nil {
		t.Fatalf("SaveSkips failed: %v", err)
	}

	if _, err := os.Stat(skipsPath); !os.IsNotExist(err) {
		t.Errorf("skips file should be removed when empty")
	}
}

func TestLoadClaim(t *testing.T) {
	dir := t.TempDir()
	claimPath := filepath.Join(dir, "numpy.yaml")

	content := `agent: build-agent-abc123
claimed_at: 2025-01-15T10:30:00Z
type: build
`
	if err := os.WriteFile(claimPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	claim, err := LoadClaim(claimPath)
	if err != nil {
		t.Fatalf("LoadClaim failed: %v", err)
	}

	if claim.Agent != "build-agent-abc123" {
		t.Errorf("Agent = %q, want %q", claim.Agent, "build-agent-abc123")
	}
	if claim.Type != "build" {
		t.Errorf("Type = %q, want %q", claim.Type, "build")
	}
	if claim.ClaimedAt.IsZero() {
		t.Errorf("ClaimedAt should not be zero")
	}
}

func TestSaveAndLoadClaim(t *testing.T) {
	dir := t.TempDir()
	claimPath := filepath.Join(dir, "claims", "numpy.yaml")

	claim := &Claim{
		Agent:     "test-agent",
		ClaimedAt: time.Now().UTC().Truncate(time.Second),
		Type:      ClaimTypeBuild,
	}

	if err := SaveClaim(claim, claimPath); err != nil {
		t.Fatalf("SaveClaim failed: %v", err)
	}

	loaded, err := LoadClaim(claimPath)
	if err != nil {
		t.Fatalf("LoadClaim failed: %v", err)
	}

	if loaded.Agent != claim.Agent {
		t.Errorf("Agent = %q, want %q", loaded.Agent, claim.Agent)
	}
	if loaded.Type != claim.Type {
		t.Errorf("Type = %q, want %q", loaded.Type, claim.Type)
	}
}
