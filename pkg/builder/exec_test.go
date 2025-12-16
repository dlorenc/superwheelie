package builder

import (
	"context"
	"testing"
	"time"
)

func TestExec(t *testing.T) {
	ctx := context.Background()

	result := Exec(ctx, "", nil, "echo", "hello")

	if !result.Success() {
		t.Errorf("Exec(echo hello) failed: %v", result.Error)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "hello\n")
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestExecFailure(t *testing.T) {
	ctx := context.Background()

	result := Exec(ctx, "", nil, "false")

	if result.Success() {
		t.Error("Exec(false) should have failed")
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
}

func TestExecWithTimeout(t *testing.T) {
	result := ExecWithTimeout(5*time.Second, "", nil, "echo", "test")

	if !result.Success() {
		t.Errorf("ExecWithTimeout failed: %v", result.Error)
	}
}

func TestExecSimple(t *testing.T) {
	output, err := ExecSimple("", "echo", "hello")
	if err != nil {
		t.Errorf("ExecSimple failed: %v", err)
	}
	if output != "hello\n" {
		t.Errorf("output = %q, want %q", output, "hello\n")
	}
}

func TestExecSimpleFailure(t *testing.T) {
	_, err := ExecSimple("", "false")
	if err == nil {
		t.Error("ExecSimple(false) should have failed")
	}
}

func TestExecResultCombinedOutput(t *testing.T) {
	tests := []struct {
		name   string
		stdout string
		stderr string
		want   string
	}{
		{"stdout only", "out", "", "out"},
		{"stderr only", "", "err", "err"},
		{"both", "out", "err", "out\nerr"},
		{"empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ExecResult{Stdout: tt.stdout, Stderr: tt.stderr}
			got := r.CombinedOutput()
			if got != tt.want {
				t.Errorf("CombinedOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}
