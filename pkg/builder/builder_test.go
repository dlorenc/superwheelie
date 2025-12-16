package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dlorenc/superwheelie/pkg/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Repo: "https://github.com/test/pkg",
	}

	b := New("/tmp/build", "testpkg", cfg)

	if b.WorkDir != "/tmp/build" {
		t.Errorf("WorkDir = %q, want %q", b.WorkDir, "/tmp/build")
	}
	if b.PackageName != "testpkg" {
		t.Errorf("PackageName = %q, want %q", b.PackageName, "testpkg")
	}
	if b.SourceDir != "/tmp/build/src" {
		t.Errorf("SourceDir = %q, want %q", b.SourceDir, "/tmp/build/src")
	}
	if b.DistDir != "/tmp/build/dist" {
		t.Errorf("DistDir = %q, want %q", b.DistDir, "/tmp/build/dist")
	}
}

func TestSetup(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		Repo: "https://github.com/test/pkg",
	}

	b := New(dir, "testpkg", cfg)
	if err := b.Setup(); err != nil {
		t.Fatalf("Setup() failed: %v", err)
	}

	// Check directories were created
	if _, err := os.Stat(b.SourceDir); os.IsNotExist(err) {
		t.Errorf("SourceDir was not created")
	}
	if _, err := os.Stat(b.DistDir); os.IsNotExist(err) {
		t.Errorf("DistDir was not created")
	}
}

func TestGetEffectiveConfig(t *testing.T) {
	cfg := &config.Config{
		Repo:       "https://github.com/test/pkg",
		SystemDeps: []string{"libfoo"},
		Env:        map[string]string{"FOO": "bar"},
		Patches:    []string{"base.patch"},
		Overrides: []config.Override{
			{
				Match:      ">=2.0",
				SystemDeps: []string{"libfoo-new"},
				Env:        map[string]string{"FOO": "baz", "NEW": "val"},
				Patches:    []string{"new.patch"},
			},
			{
				Match:  "<1.5",
				Script: "custom build",
			},
		},
	}

	b := New("/tmp/build", "testpkg", cfg)

	tests := []struct {
		name       string
		version    string
		wantDeps   []string
		wantEnv    map[string]string
		wantPatch  []string
		wantScript string
	}{
		{
			name:       "no override match",
			version:    "1.8.0",
			wantDeps:   []string{"libfoo"},
			wantEnv:    map[string]string{"FOO": "bar"},
			wantPatch:  []string{"base.patch"},
			wantScript: "",
		},
		{
			name:       "first override matches",
			version:    "2.1.0",
			wantDeps:   []string{"libfoo", "libfoo-new"},
			wantEnv:    map[string]string{"FOO": "baz", "NEW": "val"},
			wantPatch:  []string{"base.patch", "new.patch"},
			wantScript: "",
		},
		{
			name:       "second override matches",
			version:    "1.2.0",
			wantDeps:   []string{"libfoo"},
			wantEnv:    map[string]string{"FOO": "bar"},
			wantPatch:  []string{"base.patch"},
			wantScript: "custom build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eff := b.getEffectiveConfig(tt.version)

			// Check system deps
			if len(eff.SystemDeps) != len(tt.wantDeps) {
				t.Errorf("SystemDeps = %v, want %v", eff.SystemDeps, tt.wantDeps)
			}

			// Check env
			for k, v := range tt.wantEnv {
				if eff.Env[k] != v {
					t.Errorf("Env[%q] = %q, want %q", k, eff.Env[k], v)
				}
			}

			// Check patches
			if len(eff.Patches) != len(tt.wantPatch) {
				t.Errorf("Patches = %v, want %v", eff.Patches, tt.wantPatch)
			}

			// Check script
			if eff.Script != tt.wantScript {
				t.Errorf("Script = %q, want %q", eff.Script, tt.wantScript)
			}
		})
	}
}

func TestFindWheel(t *testing.T) {
	dir := t.TempDir()
	distDir := filepath.Join(dir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{Repo: "https://github.com/test/pkg"}
	b := New(dir, "testpkg", cfg)

	// Create fake wheel files
	wheels := []string{
		"testpkg-1.0.0-cp310-cp310-linux_aarch64.whl",
		"testpkg-1.0.0-cp311-cp311-linux_aarch64.whl",
		"testpkg-1.0.0-cp312-cp312-linux_aarch64.whl",
	}
	for _, w := range wheels {
		if err := os.WriteFile(filepath.Join(distDir, w), []byte("fake"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		version string
		python  string
		wantErr bool
	}{
		{"1.0.0", "3.10", false},
		{"1.0.0", "3.11", false},
		{"1.0.0", "3.12", false},
		{"1.0.0", "3.13", true}, // No wheel for 3.13
	}

	for _, tt := range tests {
		t.Run(tt.python, func(t *testing.T) {
			path, err := b.findWheel(tt.version, tt.python)
			if (err != nil) != tt.wantErr {
				t.Errorf("findWheel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && path == "" {
				t.Errorf("findWheel() returned empty path")
			}
		})
	}
}

func TestBuildEnv(t *testing.T) {
	cfg := &config.Config{Repo: "https://github.com/test/pkg"}
	b := New("/tmp/build", "testpkg", cfg)

	env := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
	}

	result := b.buildEnv(env, "3.12")

	// Check custom env vars are included
	foundFoo := false
	foundBaz := false
	for _, e := range result {
		if e == "FOO=bar" {
			foundFoo = true
		}
		if e == "BAZ=qux" {
			foundBaz = true
		}
	}

	if !foundFoo {
		t.Error("FOO=bar not found in env")
	}
	if !foundBaz {
		t.Error("BAZ=qux not found in env")
	}
}
