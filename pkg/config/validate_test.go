package config

import (
	"testing"
	"time"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Repo: "https://github.com/test/pkg",
				Versions: []Version{
					{Tag: "v1.0.0", Version: "1.0.0"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing repo",
			cfg: &Config{
				Versions: []Version{
					{Tag: "v1.0.0", Version: "1.0.0"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing versions",
			cfg: &Config{
				Repo:     "https://github.com/test/pkg",
				Versions: []Version{},
			},
			wantErr: true,
		},
		{
			name: "missing tag",
			cfg: &Config{
				Repo: "https://github.com/test/pkg",
				Versions: []Version{
					{Version: "1.0.0"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			cfg: &Config{
				Repo: "https://github.com/test/pkg",
				Versions: []Version{
					{Tag: "v1.0.0"},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate versions",
			cfg: &Config{
				Repo: "https://github.com/test/pkg",
				Versions: []Version{
					{Tag: "v1.0.0", Version: "1.0.0"},
					{Tag: "v1.0.0-rc1", Version: "1.0.0"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid override",
			cfg: &Config{
				Repo: "https://github.com/test/pkg",
				Versions: []Version{
					{Tag: "v1.0.0", Version: "1.0.0"},
				},
				Overrides: []Override{
					{Match: ">=1.0"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid override match",
			cfg: &Config{
				Repo: "https://github.com/test/pkg",
				Versions: []Version{
					{Tag: "v1.0.0", Version: "1.0.0"},
				},
				Overrides: []Override{
					{Match: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty override match",
			cfg: &Config{
				Repo: "https://github.com/test/pkg",
				Versions: []Version{
					{Tag: "v1.0.0", Version: "1.0.0"},
				},
				Overrides: []Override{
					{Match: ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSkips(t *testing.T) {
	tests := []struct {
		name    string
		skips   *Skips
		wantErr bool
	}{
		{
			name: "valid skips",
			skips: &Skips{
				Skips: []Skip{
					{Version: "1.0.0", Python: []string{"3.10"}, Reason: "test"},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty skips",
			skips:   &Skips{},
			wantErr: false,
		},
		{
			name: "missing version",
			skips: &Skips{
				Skips: []Skip{
					{Python: []string{"3.10"}, Reason: "test"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing python",
			skips: &Skips{
				Skips: []Skip{
					{Version: "1.0.0", Reason: "test"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing reason",
			skips: &Skips{
				Skips: []Skip{
					{Version: "1.0.0", Python: []string{"3.10"}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSkips(tt.skips)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSkips() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateClaim(t *testing.T) {
	tests := []struct {
		name    string
		claim   *Claim
		wantErr bool
	}{
		{
			name: "valid claim",
			claim: &Claim{
				Agent:     "test-agent",
				ClaimedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing agent",
			claim: &Claim{
				ClaimedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing claimed_at",
			claim: &Claim{
				Agent: "test-agent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClaim(tt.claim)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClaim() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidPEP440(t *testing.T) {
	tests := []struct {
		spec  string
		valid bool
	}{
		{">=1.0", true},
		{">=1.0.0", true},
		{"<2.0", true},
		{"<=2.0.0", true},
		{">1.5", true},
		{"==1.0.0", true},
		{"!=1.0.0", true},
		{"~=1.4.2", true},
		{">=1.0,<2.0", true},
		{">=1.0, <2.0", true},
		{"", false},
		{"invalid", false},
		{"1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			got := isValidPEP440(tt.spec)
			if got != tt.valid {
				t.Errorf("isValidPEP440(%q) = %v, want %v", tt.spec, got, tt.valid)
			}
		})
	}
}

func TestMatchesVersion(t *testing.T) {
	tests := []struct {
		version   string
		specifier string
		want      bool
		wantErr   bool
	}{
		{"1.0.0", ">=1.0", true, false},
		{"1.0.0", ">=1.0.0", true, false},
		{"1.0.0", ">1.0.0", false, false},
		{"2.0.0", ">1.0.0", true, false},
		{"1.0.0", "<2.0", true, false},
		{"2.0.0", "<2.0", false, false},
		{"1.0.0", "<=1.0.0", true, false},
		{"1.0.0", "==1.0.0", true, false},
		{"1.0.1", "==1.0.0", false, false},
		{"1.0.0", "!=1.0.0", false, false},
		{"1.0.1", "!=1.0.0", true, false},
		{"1.5.0", ">=1.0,<2.0", true, false},
		{"2.0.0", ">=1.0,<2.0", false, false},
		{"0.9.0", ">=1.0,<2.0", false, false},
		{"1.4.5", "~=1.4.2", true, false},
		{"1.5.0", "~=1.4.2", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.version+"_"+tt.specifier, func(t *testing.T) {
			got, err := MatchesVersion(tt.version, tt.specifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchesVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchesVersion(%q, %q) = %v, want %v", tt.version, tt.specifier, got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0", "1.0.0", 0},
		{"1.10.0", "1.9.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
