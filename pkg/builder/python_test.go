package builder

import (
	"testing"
)

func TestPythonBinary(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{"3.10", "/usr/bin/python3.10"},
		{"3.11", "/usr/bin/python3.11"},
		{"3.12", "/usr/bin/python3.12"},
		{"3.13", "/usr/bin/python3.13"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := PythonBinary(tt.version)
			if got != tt.want {
				t.Errorf("PythonBinary(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestPythonCPVersion(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{"3.10", "cp310"},
		{"3.11", "cp311"},
		{"3.12", "cp312"},
		{"3.13", "cp313"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := PythonCPVersion(tt.version)
			if got != tt.want {
				t.Errorf("PythonCPVersion(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestWheelFilename(t *testing.T) {
	tests := []struct {
		name      string
		pkg       string
		version   string
		python    string
		platform  string
		want      string
	}{
		{
			name:     "simple",
			pkg:      "numpy",
			version:  "1.26.0",
			python:   "3.12",
			platform: "linux_aarch64",
			want:     "numpy-1.26.0-cp312-cp312-linux_aarch64.whl",
		},
		{
			name:     "hyphenated name",
			pkg:      "my-package",
			version:  "1.0.0",
			python:   "3.11",
			platform: "linux_aarch64",
			want:     "my_package-1.0.0-cp311-cp311-linux_aarch64.whl",
		},
		{
			name:     "dotted name",
			pkg:      "zope.interface",
			version:  "6.0",
			python:   "3.10",
			platform: "linux_aarch64",
			want:     "zope_interface-6.0-cp310-cp310-linux_aarch64.whl",
		},
		{
			name:     "version with hyphen",
			pkg:      "foo",
			version:  "1.0.0-beta1",
			python:   "3.12",
			platform: "linux_aarch64",
			want:     "foo-1.0.0_beta1-cp312-cp312-linux_aarch64.whl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WheelFilename(tt.pkg, tt.version, tt.python, tt.platform)
			if got != tt.want {
				t.Errorf("WheelFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetPythonInfo(t *testing.T) {
	info := GetPythonInfo("3.12")

	if info.Version != "3.12" {
		t.Errorf("Version = %q, want %q", info.Version, "3.12")
	}
	if info.Binary != "/usr/bin/python3.12" {
		t.Errorf("Binary = %q, want %q", info.Binary, "/usr/bin/python3.12")
	}
	if info.CPVersion != "cp312" {
		t.Errorf("CPVersion = %q, want %q", info.CPVersion, "cp312")
	}
	if info.ABI != "cp312" {
		t.Errorf("ABI = %q, want %q", info.ABI, "cp312")
	}
}
